package syncer

import (
	"anki-sync-go/internal/anki"
	"anki-sync-go/internal/config"
	"anki-sync-go/internal/executor"
	"anki-sync-go/internal/plan"
	"anki-sync-go/internal/sheet"
	"anki-sync-go/internal/status"
	"fmt"
	"time"
)

type Syncer struct {
	cfg          config.Config
	ankiClient   *anki.Client
	sheetFetcher *sheet.Fetcher
	tts          plan.TTSProvider
	reporter     status.Reporter
}

func NewSyncer(cfg config.Config, ankiClient *anki.Client, sheetFetcher *sheet.Fetcher, tts plan.TTSProvider, reporter status.Reporter) *Syncer {
	return &Syncer{
		cfg:          cfg,
		ankiClient:   ankiClient,
		sheetFetcher: sheetFetcher,
		tts:          tts,
		reporter:     reporter,
	}
}

func (s *Syncer) report(msg status.ProgressMsg) {
	if s.reporter != nil {
		s.reporter.Report(msg)
	}
}

func (s *Syncer) Validate() error {
	s.report(status.ProgressMsg{Status: status.StatusValidating, Stage: "Config Validation"})
	// 1. Basic Structure Validation (Local)
	if err := s.cfg.Validate(); err != nil {
		return fmt.Errorf("config format error: %w", err)
	}

	s.report(status.ProgressMsg{Status: status.StatusValidating, Stage: "Checking Anki..."})
	// 2. Anki Environment Validation (Remote)
	if err := s.cfg.ValidateWithAnki(s.ankiClient); err != nil {
		return fmt.Errorf("anki error: %w", err)
	}

	s.report(status.ProgressMsg{Status: status.StatusValidating, Stage: "Checking Google Sheet..."})
	// 3. Sheet Column Validation (Remote Data)
	headers, err := s.sheetFetcher.GetHeaders()
	if err != nil {
		return fmt.Errorf("sheet error: %w", err)
	}

	if len(headers) > 0 {
		if err := s.cfg.ValidateWithSheet(headers); err != nil {
			return fmt.Errorf("sheet column error: %w", err)
		}
	} else {
		s.report(status.ProgressMsg{Status: status.StatusValidating, Log: "Warning: Sheet is empty, cannot validate columns."})
	}

	s.report(status.ProgressMsg{Status: status.StatusValidating, Log: "Validation completed"})
	return nil
}

func (s *Syncer) Run() error {
	start := time.Now()

	// Step 1: Get sheet headers
	s.report(status.ProgressMsg{Status: status.StatusFetching, Stage: "Fetching sheet headers"})
	wordCol, err := s.sheetFetcher.GetFirstColumnName()
	if err != nil {
		return fmt.Errorf("failed to get sheet headers: %w", err)
	}

	// Step 2: Fetch notes from Anki
	s.report(status.ProgressMsg{Status: status.StatusFetching, Stage: "Fetching notes from Anki"})
	ankiData, err := s.getAnkiData(wordCol)
	if err != nil {
		return fmt.Errorf("failed to get anki notes from deck %s ", s.cfg.AnkiDeck)
	}

	// Step 3: Build sync plan
	s.report(status.ProgressMsg{Status: status.StatusPlanning, Stage: "Building sync plan"})
	syncPlan, err := s.createPlan(ankiData, wordCol)
	if err != nil {
		return fmt.Errorf("failed to build sync plan: %w", err)
	}

	if s.cfg.DryRun {
		s.report(status.ProgressMsg{Status: status.StatusDone, Log: fmt.Sprintf("Dry run: %d new, %d updates", len(syncPlan.NewNotes), len(syncPlan.UpdatedNotes))})
		return nil
	}

	// Step 4: Confirm sync
	if s.reporter != nil {
		summary := fmt.Sprintf("Plan built: %d new, %d updates", len(syncPlan.NewNotes), len(syncPlan.UpdatedNotes))
		if !s.reporter.Confirm(summary) {
			s.report(status.ProgressMsg{Status: status.StatusDone, Log: "Sync cancelled by user"})
			return nil
		}
	}

	// Step 5: Execute sync
	s.report(status.ProgressMsg{Status: status.StatusExecuting, Stage: "Executing sync"})
	if err := s.executePlan(syncPlan); err != nil {
		return fmt.Errorf("failed to execute sync: %w", err)
	}

	s.report(status.ProgressMsg{Status: status.StatusDone, Log: fmt.Sprintf("Synced %d new and %d updated notes in %v", len(syncPlan.NewNotes), len(syncPlan.UpdatedNotes), time.Since(start).Round(time.Millisecond))})
	return nil
}

func (s *Syncer) getAnkiData(wordCol string) (anki.NoteMap, error) {
	rawNotes, err := s.ankiClient.GetNotesFromDeck(s.cfg.AnkiDeck)
	if err != nil {
		return nil, err
	}
	ankiData, err := anki.NormalizeAnkiNotes(rawNotes, wordCol)
	if err != nil {
		return nil, err
	}
	s.report(status.ProgressMsg{Status: status.StatusFetching, Log: fmt.Sprintf("Fetched %d notes from Anki", len(ankiData))})
	return ankiData, nil
}

func (s *Syncer) createPlan(ankiData anki.NoteMap, wordCol string) (plan.Plan, error) {
	builder := plan.NewPlanBuilder(s.cfg, s.tts, ankiData, wordCol)

	// We need a way to count records first if we want a progress bar here
	// For now, let's just report each record if possible, or just a spinner
	if err := s.sheetFetcher.ForEachRecord(func(r sheet.Record) error {
		if err := builder.Process(r); err != nil {
			return err
		}
		// Optional: report progress if we knew the total
		return nil
	}); err != nil {
		return plan.Plan{}, err
	}

	builtPlan := builder.GetPlan()
	s.report(status.ProgressMsg{Status: status.StatusPlanning, Log: fmt.Sprintf("Built plan: %d new, %d updates", len(builtPlan.NewNotes), len(builtPlan.UpdatedNotes))})
	return builtPlan, nil
}

func (s *Syncer) executePlan(p plan.Plan) error {
	exec := executor.NewExecutor(s.ankiClient, s.cfg, s.reporter)
	if err := exec.Execute(p); err != nil {
		return err
	}
	return nil
}

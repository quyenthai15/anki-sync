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

func (s *Syncer) report(msg status.Event) {
	if s.reporter != nil {
		s.reporter.Report(msg)
	}
}

func (s *Syncer) Validate() error {
	s.report(status.PhaseMsg{Step: status.Validating, Label: "Config Validation"})
	// 1. Basic Structure Validation (Local)
	if err := s.cfg.Validate(); err != nil {
		return fmt.Errorf("config format error: %w", err)
	}

	s.report(status.PhaseMsg{Step: status.Validating, Label: "Checking Anki..."})
	// 2. Anki Environment Validation (Remote)
	if err := s.cfg.ValidateWithAnki(s.ankiClient); err != nil {
		return fmt.Errorf("anki error: %w", err)
	}

	s.report(status.PhaseMsg{Step: status.Validating, Label: "Checking Google Sheet..."})
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
		s.report(status.LogMsg{Text: "Warning: Sheet is empty, cannot validate columns."})
	}

	s.report(status.LogMsg{Text: "Validation completed"})
	return nil
}

func (s *Syncer) Run() error {
	start := time.Now()

	// Step 1: Get sheet headers
	s.report(status.PhaseMsg{Step: status.Fetching, Label: "Fetching sheet headers"})
	wordCol, err := s.sheetFetcher.GetFirstColumnName()
	if err != nil {
		return fmt.Errorf("failed to get sheet headers: %w", err)
	}

	// Step 2: Fetch notes from Anki
	s.report(status.PhaseMsg{Step: status.Fetching, Label: "Fetching notes from Anki"})
	ankiData, err := s.getAnkiData(wordCol)
	if err != nil {
		return fmt.Errorf("failed to get anki notes from deck %s ", s.cfg.AnkiDeck)
	}

	// Step 3: Build sync plan
	s.report(status.PhaseMsg{Step: status.Planning, Label: "Building sync plan"})
	syncPlan, err := s.createPlan(ankiData, wordCol)
	if err != nil {
		return fmt.Errorf("failed to build sync plan: %w", err)
	}

	if s.cfg.DryRun {
		s.report(status.LogMsg{Text: fmt.Sprintf("Dry run: %d new, %d updates", len(syncPlan.NewNotes), len(syncPlan.UpdatedNotes))})
		s.report(status.PhaseMsg{Step: status.Done, Label: "Dry run complete"})
		return nil
	}

	// Step 4: Confirm sync (skip if nothing to do)
	if len(syncPlan.NewNotes) == 0 && len(syncPlan.UpdatedNotes) == 0 {
		s.report(status.LogMsg{Text: "Nothing to sync"})
		s.report(status.PhaseMsg{Step: status.Done, Label: "Nothing to sync"})
		return nil
	}

	if s.reporter != nil {
		summary := syncPlan.String()
		if !s.reporter.Confirm(summary) {
			s.report(status.LogMsg{Text: "Sync cancelled by user"})
			s.report(status.PhaseMsg{Step: status.Done, Label: "Cancelled"})
			return nil
		}
	}

	// Step 5: Execute sync
	s.report(status.PhaseMsg{Step: status.Executing, Label: "Executing sync"})
	if err := s.executePlan(syncPlan); err != nil {
		return fmt.Errorf("failed to execute sync: %w", err)
	}

	s.report(status.LogMsg{Text: fmt.Sprintf("Synced %d new and %d updated notes in %v", len(syncPlan.NewNotes), len(syncPlan.UpdatedNotes), time.Since(start).Round(time.Millisecond))})
	s.report(status.PhaseMsg{Step: status.Done, Label: "Sync complete"})
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
	s.report(status.PhaseMsg{Step: status.Fetching, Label: fmt.Sprintf("Fetched %d notes from Anki", len(ankiData))})
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
	return builtPlan, nil
}

func (s *Syncer) executePlan(p plan.Plan) error {
	exec := executor.NewExecutor(s.ankiClient, s.cfg, s.reporter)
	if err := exec.Execute(p); err != nil {
		return err
	}
	return nil
}

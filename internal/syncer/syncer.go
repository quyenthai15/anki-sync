package syncer

import (
	"anki-sync-go/internal/anki"
	"anki-sync-go/internal/config"
	"anki-sync-go/internal/executor"
	"anki-sync-go/internal/plan"
	"anki-sync-go/internal/sheet"
	"fmt"
)

type Syncer struct {
	cfg          *config.Config
	ankiClient   *anki.Client
	sheetFetcher *sheet.Fetcher
	tts          plan.TTSProvider
}

func NewSyncer(cfg *config.Config, ankiClient *anki.Client, sheetFetcher *sheet.Fetcher, tts plan.TTSProvider) *Syncer {
	return &Syncer{
		cfg:          cfg,
		ankiClient:   ankiClient,
		sheetFetcher: sheetFetcher,
		tts:          tts,
	}
}

func (s *Syncer) Validate() error {
	// 1. Basic Structure Validation (Local)
	if err := s.cfg.Validate(); err != nil {
		return fmt.Errorf("config format error: %w", err)
	}

	// 2. Anki Environment Validation (Remote)
	fmt.Println("🔍 Checking Anki...")
	if err := s.cfg.ValidateWithAnki(s.ankiClient); err != nil {
		return fmt.Errorf("anki error: %w", err)
	}

	// 3. Sheet Column Validation (Remote Data)
	fmt.Println("🔍 Checking Google Sheet...")
	headers, err := s.sheetFetcher.GetHeaders()
	if err != nil {
		return fmt.Errorf("sheet error: %w", err)
	}

	if len(headers) > 0 {
		if err := s.cfg.ValidateWithSheet(headers); err != nil {
			return fmt.Errorf("sheet column error: %w", err)
		}
	} else {
		fmt.Println("⚠️  Warning: Sheet is empty, cannot validate columns.")
	}

	return nil
}

func (s *Syncer) Run() error {
	wordCol, err := s.sheetFetcher.GetFirstColumnName()
	if err != nil {
		return fmt.Errorf("failed to get sheet headers: %w", err)
	}

	rawNotes, err := s.ankiClient.GetNotesFromDeck(s.cfg.AnkiDeck)
	if err != nil {
		return fmt.Errorf("failed to get anki notes from deck %s ", s.cfg.AnkiDeck)
	}
	ankiData, err := anki.NormalizeAnkiNotes(rawNotes, wordCol)
	if err != nil {
		return fmt.Errorf("failed to get anki notes detail: %w", err)
	}
	fmt.Printf("Fetched %d notes from Anki\n", len(ankiData))

	builder := plan.NewPlanBuilder(s.cfg, s.tts, ankiData, wordCol)
	err = s.sheetFetcher.ForEachRecord(builder.Process)
	if err != nil {
		return fmt.Errorf("failed to build sync plan: %w", err)
	}

	syncPlan := builder.GetPlan()

	if s.cfg.DryRun {
		fmt.Printf("- Adding %d notes: %+v\n", len(syncPlan.NewNotes), syncPlan.NewNotes)
		fmt.Printf("- Updating %d notes: %+v\n", len(syncPlan.UpdatedNotes), syncPlan.UpdatedNotes)
		fmt.Println("(Dry run mode - not adding notes to Anki)")
		return nil
	}

	// Post plan to anki
	exec := executor.NewExecutor(s.ankiClient, s.cfg)
	if err := exec.Execute(syncPlan); err != nil {
		return fmt.Errorf("failed to execute sync: %w", err)
	}

	fmt.Printf("Synced %d new notes and %d updated notes\n", len(syncPlan.NewNotes), len(syncPlan.UpdatedNotes))
	return nil
}

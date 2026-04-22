package main

import (
	"anki-sync-go/internal/anki"
	"anki-sync-go/internal/config"
	"anki-sync-go/internal/executor"
	"anki-sync-go/internal/plan"
	"anki-sync-go/internal/sheet"
	"flag"
	"fmt"
	"log"
)

type SheetFetcher interface {
	ForEachRecord(callback func(sheet.Record) error) error
}

func main() {
	force := flag.Bool("force", false, "Force sync")
	dryRun := flag.Bool("dry-run", false, "Dry run")
	flag.Parse()

	cfg, err := config.LoadConfig("config.json")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	cfg.ForceSync = *force
	cfg.DryRun = *dryRun

	// Fetch sheet records & anki notes to compare get diff
	ankiClient := anki.NewClient()
	sheetFetcher := sheet.NewFetcher(cfg.SheetCSVURL)

	primarySheetCol, err := sheetFetcher.GetFirstColumnName()
	if err != nil {
		log.Fatalf("Failed to get sheet headers: %v", err)
	}

	ankiData, err := anki.NormalizeAnkiNotes(ankiClient, cfg.AnkiDeck, primarySheetCol)
	if err != nil {
		log.Fatalf("Failed to get anki notes detail: %v", err)
	}
	fmt.Printf("Fetched %d notes from Anki\n", len(ankiData))

	syncPlan := plan.NewSyncPlan(cfg)
	sheetFetcher.ForEachRecord(func(r sheet.Record) error {
		targetFields := cfg.MapRecord(r)
		
		existing, found := ankiData[r[primarySheetCol]]
		if !found {
			// Build new note action
			syncPlan.AddNewNote(targetFields)
		} else if cfg.ForceSync {
			// Check if note is different
			diff := plan.Diff(targetFields, existing.Fields)
			if diff != nil {
				// Build updated note action
				syncPlan.AddUpdatedNote(existing.NoteID, diff)
			}
		}
		return nil
	})

	if cfg.DryRun {
		fmt.Printf("- Adding %d notes: %v\n", len(syncPlan.NewNotes), syncPlan.NewNotes)
		fmt.Printf("- Updating %d notes: %v\n", len(syncPlan.UpdatedNotes), syncPlan.UpdatedNotes)
		fmt.Println("(Dry run mode - not adding notes to Anki)")
		return
	}
	// Post plan to anki
	executor := executor.NewExecutor(ankiClient)
	if len(syncPlan.NewNotes) > 0 {
		if err := executor.ExecuteAdditions(syncPlan.NewNotes); err != nil {
			log.Fatalf("Failed to execute additions: %v", err)
		}
	}
	if len(syncPlan.UpdatedNotes) > 0 {
		if err := executor.ExecuteUpdates(syncPlan.UpdatedNotes); err != nil {
			log.Fatalf("Failed to execute updates: %v", err)
		}
	}
	fmt.Printf("Synced %d new notes and %d updated notes\n", len(syncPlan.NewNotes), len(syncPlan.UpdatedNotes))
}

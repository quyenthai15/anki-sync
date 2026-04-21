package main

import (
	"anki-sync-go/internal/anki"
	"anki-sync-go/internal/config"
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

	primaryColumn, err := sheetFetcher.GetFirstColumnName()
	if err != nil {
		log.Fatalf("Failed to get sheet headers: %v", err)
	}

	ankiData, err := anki.NormalizeAnkiNotes(ankiClient, cfg.AnkiDeck, primaryColumn)
	if err != nil {
		log.Printf("Failed to get anki notes detail: %v", err)
	}
	fmt.Printf("Fetched %d notes from Anki\n", len(ankiData))

	syncPlan := plan.SyncPlan{}
	sheetFetcher.ForEachRecord(func(r sheet.Record) error {
		existing, found := ankiData[r[primaryColumn]]
		if !found {
			// Build new note action
			syncPlan.AddNewNote(r)
		} else if cfg.ForceSync {
			// Check if note is different
			diff := plan.GetDiff(r, existing)
			if diff != nil {
				// Build updated note action
				syncPlan.AddUpdatedNote(existing.NoteID, diff)
			}
		}
		return nil
	})

	if cfg.DryRun {
		fmt.Println("Dry run mode - not adding notes to Anki")
		return
	}
	// Post plan to anki
	if len(syncPlan.NewNotes) > 0 {
		// TODO: Post new notes to anki
	}
	if len(syncPlan.UpdatedNotes) > 0 {
		// TODO: Post updated notes to anki
	}
}

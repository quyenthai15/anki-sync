package main

import (
	"anki-sync-go/internal/anki"
	"anki-sync-go/internal/config"
	"anki-sync-go/internal/sheet"
	"flag"
	"fmt"
	"log"
)


type AnkiReader interface {
	GetModels() ([]string, error)
	GetDeckNames() ([]string, error)
	GetModelFields(modelName string) ([]string, error) // TODO: this is to initial setup, not sure if still needed
	GetNotes(deckName string) ([]int, error)
	GetNoteInfo(noteId int) (map[string]any, error)
}

type AnkiWriter interface {
	AddNotes(notes []anki.Note) error
	MultiUpdate(actions []anki.Request[map[string]any]) ([]any, error)
}

type SheetFetcher interface {
	GetRecords() ([]sheet.Record, error)
}

/**
* This function should:
* 1. Load config
* 2. Validate config
* 3. Create anki client
* 4. Create sheet client
* 5. Fetch records
* 6. Create notes
* 7. Add notes to anki
 */
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

	if err := config.Check(cfg); err != nil {
		log.Fatalf("Validation failed: %v", err)
	} else {
		fmt.Println("Validation passed")
	}

	// Fetch sheet records & anki notes to compare get diff
	ankiClient := anki.NewClient()
	sheetFetcher := sheet.NewFetcher(cfg.SheetCSVURL)

	records, err := sheetFetcher.GetRecords()
	if err != nil {
		log.Fatalf("Failed to fetch sheet records: %v", err)
	}
	fmt.Printf("Fetched %d records from sheet\n", len(records))

	models, err := ankiClient.GetModels()
	if err != nil {
		log.Fatalf("Failed to get models: %v", err)
	}
	fmt.Println("Anki Models:", models)

	if cfg.DryRun {
		fmt.Println("Dry run mode - not adding notes to Anki")
		return
	}

	if cfg.ForceSync {
		// if force sync, update existing notes with new data
	}

	// Create notes
	// Add notes to anki
}


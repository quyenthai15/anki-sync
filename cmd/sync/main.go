package main

import (
	"anki-sync-go/internal/anki"
	"anki-sync-go/internal/config"
	"anki-sync-go/internal/sheet"
	"flag"
	"fmt"
	"log"
)


type AnkiClient interface {
	AddNotes(notes []anki.Note) error
	// TODO: these are for validation, not sure if needed
	// GetModels() ([]string, error)
	// GetDeckNames() ([]string, error)
	// GetModelFields(modelName string) ([]string, error) -> this is to initial setup, not sure if still needed
	GetNotes(deckName string) ([]int, error)
	GetNoteInfo(noteId int) (map[string]any, error)
	MultiUpdate(actions []anki.Action) ([]any, error)
}


type SheetFetcher interface {
	GetRecords(url string) ([]sheet.Record, error)
}

func main() {
	force := flag.Bool("force", false, "Force sync")
	dryRun := flag.Bool("dry-run", true, "Dry run")
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
}


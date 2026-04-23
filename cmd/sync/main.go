package main

import (
	"anki-sync-go/internal/anki"
	"anki-sync-go/internal/config"
	"anki-sync-go/internal/plan"
	"anki-sync-go/internal/sheet"
	"anki-sync-go/internal/syncer"
	"flag"
	"log"
)

func main() {
	force := flag.Bool("force", false, "Force sync")
	dryRun := flag.Bool("dry-run", false, "Dry run")
	skipValidate := flag.Bool("skip-validate", false, "Skip validation")
	flag.Parse()

	cfg, err := config.LoadConfig("config.json")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	cfg.ForceSync = *force
	cfg.DryRun = *dryRun

	ankiClient := anki.NewClient()
	sheetFetcher := sheet.NewFetcher(cfg.SheetCSVURL)
	tts := plan.NewGoogleTTS("ja")

	syncSvc := syncer.NewSyncer(cfg, ankiClient, sheetFetcher, tts)

	if !*skipValidate {
		if err := syncSvc.Validate(); err != nil {
			log.Fatalf("Validation failed: %v", err)
		}
	}

	if err := syncSvc.Run(); err != nil {
		log.Fatalf("Sync failed: %v", err)
	}
}

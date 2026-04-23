package main

import (
	"anki-sync-go/internal/anki"
	"anki-sync-go/internal/config"
	"anki-sync-go/internal/plan"
	"anki-sync-go/internal/sheet"
	"anki-sync-go/internal/syncer"
	"flag"
	"fmt"
	"log"
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: sync [options]\n\n")
		fmt.Fprintf(flag.CommandLine.Output(), "Anki sync tool synchronizes Google Sheet data with Anki flashcards.\n\n")
		fmt.Fprintf(flag.CommandLine.Output(), "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(flag.CommandLine.Output(), "\nExamples:\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  sync                           # Normal sync\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  sync --dry-run                # Preview changes without executing\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  sync --force                  # Force update existing notes\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  sync --skip-validate          # Skip validation checks (faster)\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  sync --dry-run --force        # Preview forced updates\n")
	}

	force := flag.Bool("force", false, "Force sync of existing notes (updates audio and fields)")
	dryRun := flag.Bool("dry-run", false, "Preview changes without executing them")
	skipValidate := flag.Bool("skip-validate", false, "Skip validation checks (deck, model, columns)")
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

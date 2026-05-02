package main

import (
	"anki-sync-go/internal/anki"
	"anki-sync-go/internal/config"
	"anki-sync-go/internal/plan"
	"anki-sync-go/internal/sheet"
	"anki-sync-go/internal/status"
	"anki-sync-go/internal/syncer"
	"anki-sync-go/internal/ui"
	"flag"
	"fmt"
	"log"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	// ... (flag parsing remains the same)
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: sync [options]\n\n")
		fmt.Fprintf(flag.CommandLine.Output(), "Anki sync tool synchronizes Google Sheet data with Anki flashcards.\n\n")
		fmt.Fprintf(flag.CommandLine.Output(), "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(flag.CommandLine.Output(), "\nExamples:\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  sync                           # Normal sync\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  sync --skip-validate          # Skip validation checks (faster)\n")
	}

	force := flag.Bool("force", true, "Force sync of existing notes (updates audio and fields)")
	skipValidate := flag.Bool("skip-validate", false, "Skip validation checks (deck, model, columns)")
	flag.Parse()

	cfg, err := config.LoadConfig("config.json")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	cfg.ForceSync = *force

	tui := ui.NewTUI()
	p := tea.NewProgram(tui)

	ankiClient := anki.NewClient()
	sheetFetcher := sheet.NewFetcher(cfg.SheetCSVURL)
	tts := plan.NewGoogleTTS("ja")

	syncSvc := syncer.NewSyncer(cfg, ankiClient, sheetFetcher, tts, tui)

	go func() {
		if !*skipValidate {
			if err := syncSvc.Validate(); err != nil {
				tui.Report(status.ErrMsg{Error: err})
				return
			}
		}

		if err := syncSvc.Run(); err != nil {
			tui.Report(status.ErrMsg{Error: err})
			return
		}
	}()

	if _, err := p.Run(); err != nil {
		log.Fatalf("Error running TUI: %v", err)
	}
}

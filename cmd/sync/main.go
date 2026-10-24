package main

import (
	"anki-sync-go/internal/anki"
	"anki-sync-go/internal/config"
	"anki-sync-go/internal/importer"
	"anki-sync-go/internal/plan"
	"anki-sync-go/internal/sheet"
	"anki-sync-go/internal/status"
	"anki-sync-go/internal/syncer"
	"anki-sync-go/internal/ui"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	// Handle subcommands first (import, check, profiles)
	if len(os.Args) > 1 && !strings.HasPrefix(os.Args[1], "-") {
		cmd := strings.ToUpper(os.Args[1])

		switch cmd {
		case "CHECK":
			client := anki.NewClient()
			if err := client.Check(); err != nil {
				fmt.Fprintf(os.Stderr, "✘ AnkiConnect not reachable: %v\n(Make sure Anki desktop is open with AnkiConnect enabled on localhost:8765)\n", err)
				os.Exit(1)
			}
			fmt.Println("✔ AnkiConnect is running on localhost:8765")
			return

		case "PROFILES", "PROFILE":
			profiles, err := importer.LoadProfiles(".profiles.json")
			if err != nil {
				fmt.Fprintf(os.Stderr, "✘ Failed to load profiles: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Configured Profiles (%d):\n", len(profiles))
			for name, p := range profiles {
				fmt.Printf("  • %-10s Deck: %-15s Model: %-12s Lang: %s -> %s\n", name, p.Deck, p.Model, p.TargetLang, p.TranslationLang)
			}
			return

		case "IMPORT":
			handleImport(os.Args[2:])
			return

		default:
			fmt.Fprintf(os.Stderr, "Unknown command '%s'. Available commands: import, check, profiles\n", os.Args[1])
			os.Exit(1)
		}
	}

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: sync [command|options]\n\n")
		fmt.Fprintf(flag.CommandLine.Output(), "Commands:\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  import [file] [--dry-run]     Import draft YAML cards into Anki\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  check                         Check AnkiConnect availability\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  profiles                      List configured language profiles\n\n")
		fmt.Fprintf(flag.CommandLine.Output(), "Options (Sheet Sync):\n")
		flag.PrintDefaults()
	}

	// Legacy Google Sheet sync pipeline with TUI
	force := flag.Bool("force", false, "Force sync of existing notes (updates audio and fields)")
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

func handleImport(args []string) {
	importFlags := flag.NewFlagSet("import", flag.ExitOnError)
	dryRun := importFlags.Bool("dry-run", false, "Validate and preview notes without adding to Anki")
	importFlags.BoolVar(dryRun, "n", false, "Alias for --dry-run")
	_ = importFlags.Parse(args)

	draftPath := ".anki-draft.yaml"
	if len(importFlags.Args()) > 0 {
		draftPath = importFlags.Args()[0]
	}

	result, err := importer.Import(draftPath, *dryRun)
	if err != nil {
		fmt.Fprintf(os.Stderr, "✘ Import failed: %v\n", err)
		os.Exit(1)
	}

	if result.DryRun {
		fmt.Printf("[DRY RUN] Validated draft '%s' for profile '%s' (Deck: %s, Model: %s)\n", draftPath, result.Profile, result.Deck, result.Model)
		fmt.Printf("Cards to import (%d):\n\n", result.CardCount)
		for i, note := range result.Notes {
			expr := note.Fields["Expression"]
			meaning := note.Fields["Meaning"]
			reading := note.Fields["Reading"]
			fmt.Printf("  %d. %s (%s)\n", i+1, expr, meaning)
			if reading != "" {
				fmt.Printf("     • Reading: %s\n", reading)
			}
			for _, aud := range note.Audio {
				if len(aud.Fields) > 0 {
					fmt.Printf("     • Audio [%s]: %s\n", aud.Fields[0], aud.Filename)
				}
			}
			fmt.Println()
		}
		fmt.Println("✔ Dry-run successful. No changes made to Anki.")
		return
	}

	fmt.Printf("✔ Successfully imported %d card(s) into Anki deck '%s' (Model: %s)!\n", result.CardCount, result.Deck, result.Model)
}

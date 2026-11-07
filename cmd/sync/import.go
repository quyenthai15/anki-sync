package main

import (
	"anki-sync-go/internal/importer"
	"flag"
	"fmt"
	"os"
)

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

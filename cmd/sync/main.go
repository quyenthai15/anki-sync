package main

import (
	"anki-sync-go/internal/anki"
	"anki-sync-go/internal/importer"
	"fmt"
	"io"
	"os"
	"strings"
)

func main() {
	if len(os.Args) < 2 || strings.HasPrefix(os.Args[1], "-") {
		printUsage(os.Stdout)
		return
	}

	switch strings.ToLower(os.Args[1]) {
	case "check":
		cmdCheck()
	case "profiles", "profile":
		cmdProfiles()
	case "import":
		handleImport(os.Args[2:])
	case "search":
		handleSearch(os.Args[2:])
	default:
		fmt.Fprintf(os.Stderr, "✘ Unknown command %q\n\n", os.Args[1])
		printUsage(os.Stderr)
		os.Exit(1)
	}
}

func cmdCheck() {
	if err := anki.NewClient().Check(); err != nil {
		fmt.Fprintf(os.Stderr, "✘ AnkiConnect not reachable: %v\n(Make sure Anki desktop is open with AnkiConnect enabled on localhost:8765)\n", err)
		os.Exit(1)
	}
	fmt.Println("✔ AnkiConnect is running on localhost:8765")
}

func cmdProfiles() {
	profiles, err := importer.LoadProfiles(".profiles.json")
	if err != nil {
		fmt.Fprintf(os.Stderr, "✘ Failed to load profiles: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Configured Profiles (%d):\n", len(profiles))
	for name, p := range profiles {
		fmt.Printf("  • %-10s Deck: %-15s Model: %-12s Lang: %s -> %s\n", name, p.Deck, p.Model, p.TargetLang, p.TranslationLang)
	}
}

func printUsage(w io.Writer) {
	fmt.Fprintf(w, "Usage: anki-sync <command> [options]\n\n")
	fmt.Fprintf(w, "Commands:\n")
	fmt.Fprintf(w, "  check                         Check AnkiConnect availability\n")
	fmt.Fprintf(w, "  profiles                      List configured language profiles\n")
	fmt.Fprintf(w, "  import [file] [--dry-run]     Import draft YAML cards into Anki\n")
	fmt.Fprintf(w, "  search <words...>             Search notes in a deck\n\n")
	fmt.Fprintf(w, "Search options:\n")
	fmt.Fprintf(w, "  --deck <name>                 Search this deck\n")
	fmt.Fprintf(w, "  --profile <name>              Search the profile's deck\n")
}

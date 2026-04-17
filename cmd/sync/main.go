package main

import (
	"anki-sync-go/internal/anki"
	"anki-sync-go/internal/config"
	"flag"
	"fmt"
	"log"
)

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

	anki.AnkiAction("action", nil)
}


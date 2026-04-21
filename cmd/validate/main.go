package main

import (
	"anki-sync-go/internal/anki"
	"anki-sync-go/internal/config"
	"anki-sync-go/internal/sheet"
	"fmt"
	"log"
)

func main() {
	log.SetFlags(0)

	cfg, err := config.LoadConfig("config.json")
	if err != nil {
		log.Fatalf("❌ Failed to load config: %v", err)
	}

	// 1. Basic Structure Validation (Local)
	if err := cfg.Validate(); err != nil {
		log.Fatalf("❌ Config format error: %v", err)
	}

	ankiClient := anki.NewClient()
	sheetFetcher := sheet.NewFetcher(cfg.SheetCSVURL)

	// 2. Anki Environment Validation (Remote)
	fmt.Println("🔍 Checking Anki...")
	if err := cfg.ValidateWithAnki(ankiClient); err != nil {
		log.Fatalf("❌ Anki error: %v", err)
	}

	// 3. Sheet Column Validation (Remote Data)
	fmt.Println("🔍 Checking Google Sheet...")
	headers, err := sheetFetcher.GetHeaders()
	if err != nil {
		log.Fatalf("❌ Sheet error: %v", err)
	}

	if len(headers) > 0 {
		if err := cfg.ValidateWithSheet(headers); err != nil {
			log.Fatalf("❌ Sheet column error: %v", err)
		}
	} else {
		fmt.Println("⚠️  Warning: Sheet is empty, cannot validate columns.")
	}

	fmt.Println("✅ Configuration verified! Ready to sync.")
}

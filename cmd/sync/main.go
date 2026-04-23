package main

import (
	"anki-sync-go/internal/anki"
	"anki-sync-go/internal/audio"
	"anki-sync-go/internal/config"
	"anki-sync-go/internal/executor"
	"anki-sync-go/internal/plan"
	"anki-sync-go/internal/sheet"
	"flag"
	"fmt"
	"log"
)

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

	// Fetch sheet records & anki notes to compare get diff
	ankiClient := anki.NewClient()
	sheetFetcher := sheet.NewFetcher(cfg.SheetCSVURL)
	tts := audio.NewGoogleTTS("ja")

	wordCol, err := sheetFetcher.GetFirstColumnName()
	if err != nil {
		log.Fatalf("Failed to get sheet headers: %v", err)
	}

	ankiData, err := anki.NormalizeAnkiNotes(ankiClient, cfg.AnkiDeck, wordCol)
	if err != nil {
		log.Fatalf("Failed to get anki notes detail: %v", err)
	}
	fmt.Printf("Fetched %d notes from Anki\n", len(ankiData))

	wordAudio := audio.NewAudioProcessor(tts, "ja_word_tts")
	sentAudio := audio.NewAudioProcessor(tts, "ja_sent_tts")

	builder := plan.NewPlanBuilder(cfg, wordAudio, sentAudio, ankiData, wordCol)
	err = sheetFetcher.ForEachRecord(builder.Process)
	if err != nil {
		log.Fatalf("Failed to build sync plan: %v", err)
	}

	syncPlan := builder.GetPlan()

	if cfg.DryRun {
		fmt.Printf("- Adding %d notes: %+v\n", len(syncPlan.NewNotes), syncPlan.NewNotes)
		fmt.Printf("- Updating %d notes: %+v\n", len(syncPlan.UpdatedNotes), syncPlan.UpdatedNotes)
		fmt.Println("(Dry run mode - not adding notes to Anki)")
		return
	}

	// Post plan to anki
	exec := executor.NewExecutor(ankiClient, cfg)
	if err := exec.Execute(syncPlan); err != nil {
		log.Fatalf("Failed to execute sync: %v", err)
	}

	fmt.Printf("Synced %d new notes and %d updated notes\n", len(syncPlan.NewNotes), len(syncPlan.UpdatedNotes))
}

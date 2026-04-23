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
	"strings"
)

type SheetFetcher interface {
	ForEachRecord(callback func(sheet.Record) error) error
}

func hasSound(val string) bool {
	return strings.Contains(val, "[sound:")
}

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
	audioBuilder := audio.NewGoogleTTS("ja")

	primarySheetCol, err := sheetFetcher.GetFirstColumnName()
	if err != nil {
		log.Fatalf("Failed to get sheet headers: %v", err)
	}

	ankiData, err := anki.NormalizeAnkiNotes(ankiClient, cfg.AnkiDeck, primarySheetCol)
	if err != nil {
		log.Fatalf("Failed to get anki notes detail: %v", err)
	}
	fmt.Printf("Fetched %d notes from Anki\n", len(ankiData))

	syncPlan := plan.NewSyncPlan(cfg)
	sheetFetcher.ForEachRecord(func(r sheet.Record) error {
		targetFields := cfg.MapRecord(r)
		word := r[primarySheetCol]

		existing, found := ankiData[word]
		var audioInfo []anki.AudioInfo

		if !found {
			// New note - Word Audio
			if cfg.AudioField != "" && word != "" {
				filename := audio.SanitizeFilename(word, "ja_word_tts")
				audioInfo = append(audioInfo, anki.AudioInfo{
					URL:      audioBuilder.GetTts(word),
					Filename: filename,
					Fields:   []string{cfg.AudioField},
				})
				targetFields[cfg.AudioField] = fmt.Sprintf("[sound:%s]", filename)
			}
			// New note - Sentence Audio
			if cfg.SentenceAudioField != "" && cfg.SentenceSourceCol != "" {
				sentence := targetFields[cfg.SentenceSourceCol]
				if sentence != "" {
					filename := audio.SanitizeFilename(sentence, "ja_sent_tts")
					audioInfo = append(audioInfo, anki.AudioInfo{
						URL:      audioBuilder.GetTts(sentence),
						Filename: filename,
						Fields:   []string{cfg.SentenceAudioField},
					})
					targetFields[cfg.SentenceAudioField] = fmt.Sprintf("[sound:%s]", filename)
				}
			}
			syncPlan.AddNewNote(targetFields, audioInfo)
		} else if cfg.ForceSync {
			// Update note - Word Audio (if missing)
			if cfg.AudioField != "" && word != "" {
				if !hasSound(existing.Fields[cfg.AudioField]) {
					filename := audio.SanitizeFilename(word, "ja_word_tts")
					audioInfo = append(audioInfo, anki.AudioInfo{
						URL:      audioBuilder.GetTts(word),
						Filename: filename,
						Fields:   []string{cfg.AudioField},
					})
					targetFields[cfg.AudioField] = fmt.Sprintf("[sound:%s]", filename)
				}
			}
			// Update note - Sentence Audio (if missing or sentence changed)
			if cfg.SentenceAudioField != "" && cfg.SentenceSourceCol != "" {
				sentence := targetFields[cfg.SentenceSourceCol]
				sentenceChanged := existing.Fields[cfg.SentenceSourceCol] != sentence

				if sentence != "" && (!hasSound(existing.Fields[cfg.SentenceAudioField]) || sentenceChanged) {
					filename := audio.SanitizeFilename(sentence, "ja_sent_tts")
					audioInfo = append(audioInfo, anki.AudioInfo{
						URL:      audioBuilder.GetTts(sentence),
						Filename: filename,
						Fields:   []string{cfg.SentenceAudioField},
					})
					targetFields[cfg.SentenceAudioField] = fmt.Sprintf("[sound:%s]", filename)
				}
			}

			diff := plan.Diff(targetFields, existing.Fields)
			if diff != nil || len(audioInfo) > 0 {
				syncPlan.AddUpdatedNote(existing.NoteID, diff, audioInfo)
			}
		}
		return nil
	})

	if cfg.DryRun {
		fmt.Printf("- Adding %d notes: %+v\n", len(syncPlan.NewNotes), syncPlan.NewNotes)
		fmt.Printf("- Updating %d notes: %+v\n", len(syncPlan.UpdatedNotes), syncPlan.UpdatedNotes)
		fmt.Println("(Dry run mode - not adding notes to Anki)")
		return
	}
	// Post plan to anki
	executor := executor.NewExecutor(ankiClient, audioBuilder, cfg)
	if len(syncPlan.NewNotes) > 0 {
		if err := executor.ExecuteAdditions(syncPlan.NewNotes); err != nil {
			log.Fatalf("Failed to execute additions: %v", err)
		}
	}
	if len(syncPlan.UpdatedNotes) > 0 {
		if err := executor.ExecuteUpdates(syncPlan.UpdatedNotes); err != nil {
			log.Fatalf("Failed to execute updates: %v", err)
		}
	}
	fmt.Printf("Synced %d new notes and %d updated notes\n", len(syncPlan.NewNotes), len(syncPlan.UpdatedNotes))
}

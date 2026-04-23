package plan

import (
	"anki-sync-go/internal/anki"
	"anki-sync-go/internal/config"
	"anki-sync-go/internal/sheet"
	"testing"
)

func TestPlanBuilder_Process_Duplicates(t *testing.T) {
	cfg := &config.Config{
		AnkiDeck:           "Deck",
		AnkiModel:          "Model",
		Mapping:            config.FieldMapping{"Word": "Expression", "Meaning": "Meaning", "Sentence": "Sentence"},
		AudioField:         "Audio",
		SentenceAudioField: "SentenceAudio",
		SentenceSourceCol:  "Sentence",
		ForceSync:          true,
	}

	tts := &mockTTS{}

	existingNotes := anki.NoteMap{
		"apple": anki.NormalizedNote{
			NoteID: 1,
			Fields: map[string]string{
				"Expression": "apple",
				"Meaning":    "fruit",
				"Sentence":   "I eat an apple",
				"Audio":      "[sound:word_apple.mp3]", // ALREADY HAS SOUND
			},
		},
	}

	builder := NewPlanBuilder(cfg, tts, existingNotes, "Word")

	// Record with same word
	record := sheet.Record{
		"Word":     "apple",
		"Meaning":  "fruit",
		"Sentence": "I eat an apple",
	}

	err := builder.Process(record)
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	p := builder.GetPlan()

	// Should NOT have an update for Audio if it already exists,
	// unless something ELSE changed.
	// In this case, SentenceAudio is missing, so it SHOULD have an update for that.
	if len(p.UpdatedNotes) != 1 {
		t.Errorf("Expected 1 updated note (for sentence audio), got %d", len(p.UpdatedNotes))
	}

	if len(p.UpdatedNotes) > 0 {
		update := p.UpdatedNotes[0]
		if _, exists := update.Fields["Audio"]; exists {
			t.Errorf("Should NOT have updated Audio field as it already has a sound tag")
		}

		// Verify we didn't accidentally add two Audio infos
		audioCount := 0
		for _, info := range update.Audio {
			for _, field := range info.Fields {
				if field == "Audio" {
					audioCount++
				}
			}
		}
		if audioCount > 1 {
			t.Errorf("Found %d audio infos for 'Audio' field, expected at most 1", audioCount)
		}
	}
}

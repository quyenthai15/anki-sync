package importer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseYaml(t *testing.T) {
	raw := `
profile: japanese
cards:
  - Expression: たまご
    Reading: たまご [2]
    Meaning: trứng
    Kanji: null
    Type: Noun
    Level: 5
    Sentence_JP: あさごはんに たまごを たべます。
    Sentence_VN: Tôi ăn trứng vào bữa sáng.
`
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, ".anki-draft.yaml")
	if err := os.WriteFile(filePath, []byte(raw), 0644); err != nil {
		t.Fatalf("failed to write temp draft: %v", err)
	}

	draft, err := ParseYaml(filePath)
	if err != nil {
		t.Fatalf("ParseYaml failed: %v", err)
	}

	if draft.Profile != "japanese" {
		t.Errorf("expected profile 'japanese', got '%s'", draft.Profile)
	}
	if len(draft.Cards) != 1 {
		t.Fatalf("expected 1 card, got %d", len(draft.Cards))
	}

	card := draft.Cards[0]
	if card["Expression"] != "たまご" {
		t.Errorf("expected Expression 'たまご', got '%s'", card["Expression"])
	}
	if card["Reading"] != "たまご [2]" {
		t.Errorf("expected Reading 'たまご [2]', got '%s'", card["Reading"])
	}
	if card["Sentence_JP"] != "あさごはんに たまごを たべます。" {
		t.Errorf("expected Sentence_JP, got '%s'", card["Sentence_JP"])
	}
}

func TestBuildNotesJapanese(t *testing.T) {
	draft := &Draft{
		Profile: "japanese",
		Cards: []Card{
			{
				"Expression":  "ねこ",
				"Reading":     "ねこ [1]",
				"Meaning":     "con mèo",
				"Kanji":       "猫",
				"Sentence_JP": "しろい ねこが います。",
				"Sentence_VN": "Có một con mèo trắng.",
			},
		},
	}

	profile := &Profile{
		Deck:       "Japanese::N5",
		Model:      "JP_sheet",
		TargetLang: "ja",
		Audio: map[string]string{
			"Audio":         "Expression",
			"SentenceAudio": "Sentence_JP",
		},
	}

	notes, err := BuildNotes(draft, profile)
	if err != nil {
		t.Fatalf("BuildNotes failed: %v", err)
	}

	if len(notes) != 1 {
		t.Fatalf("expected 1 note, got %d", len(notes))
	}

	note := notes[0]
	if note.DeckName != "Japanese::N5" {
		t.Errorf("expected deck Japanese::N5, got %s", note.DeckName)
	}
	if note.ModelName != "JP_sheet" {
		t.Errorf("expected model JP_sheet, got %s", note.ModelName)
	}
	if note.Fields["Expression"] != "ねこ" {
		t.Errorf("expected Expression 'ねこ', got %s", note.Fields["Expression"])
	}

	if len(note.Audio) != 2 {
		t.Fatalf("expected 2 audio items, got %d", len(note.Audio))
	}

	// Audio for Expression
	audio0 := note.Audio[0]
	if audio0.Fields[0] != "Audio" {
		t.Errorf("expected audio field 'Audio', got %s", audio0.Fields[0])
	}
	if !strings.Contains(audio0.URL, "tl=ja") {
		t.Errorf("expected URL to contain tl=ja, got %s", audio0.URL)
	}
	if !strings.HasPrefix(audio0.Filename, "ja_Audio_") {
		t.Errorf("expected filename prefix 'ja_Audio_', got %s", audio0.Filename)
	}

	// Audio for Sentence_JP
	audio1 := note.Audio[1]
	if audio1.Fields[0] != "SentenceAudio" {
		t.Errorf("expected audio field 'SentenceAudio', got %s", audio1.Fields[0])
	}
	if !strings.HasPrefix(audio1.Filename, "ja_SentenceAudio_") {
		t.Errorf("expected filename prefix 'ja_SentenceAudio_', got %s", audio1.Filename)
	}
}

func TestBuildNotesFrench(t *testing.T) {
	draft := &Draft{
		Profile: "french",
		Cards: []Card{
			{
				"Expression":  "le livre",
				"Reading":     "/livʁ/",
				"Meaning":     "cuốn sách",
				"Sentence_FR": "Je lis un livre.",
				"Sentence_VN": "Tôi đọc một cuốn sách.",
			},
		},
	}

	profile := &Profile{
		Deck:       "French::A1",
		Model:      "French_sheet",
		TargetLang: "fr",
		Audio: map[string]string{
			"Audio":         "Expression",
			"SentenceAudio": "Sentence_FR",
		},
	}

	notes, err := BuildNotes(draft, profile)
	if err != nil {
		t.Fatalf("BuildNotes failed: %v", err)
	}

	note := notes[0]
	if note.DeckName != "French::A1" {
		t.Errorf("expected deck French::A1, got %s", note.DeckName)
	}
	if len(note.Audio) != 2 {
		t.Fatalf("expected 2 audio items, got %d", len(note.Audio))
	}
	if !strings.Contains(note.Audio[0].URL, "tl=fr") {
		t.Errorf("expected French URL tl=fr, got %s", note.Audio[0].URL)
	}
}

func TestImportDryRun(t *testing.T) {
	draftPath := filepath.Join("..", "..", ".anki-draft.yaml")
	if _, err := os.Stat(draftPath); os.IsNotExist(err) {
		t.Skip("skipping test: .anki-draft.yaml does not exist")
	}

	result, err := Import(draftPath, true)
	if err != nil {
		t.Fatalf("Import in dryRun failed: %v", err)
	}

	if !result.DryRun {
		t.Errorf("expected DryRun true")
	}
	if result.Profile != "japanese" {
		t.Errorf("expected Profile japanese, got %s", result.Profile)
	}
	if result.CardCount != 2 {
		t.Errorf("expected CardCount 2, got %d", result.CardCount)
	}
	if len(result.Notes) != 2 {
		t.Errorf("expected 2 notes, got %d", len(result.Notes))
	}
	t.Logf("Dry run result: %+v", result)
}

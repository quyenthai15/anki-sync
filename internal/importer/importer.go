package importer

import (
	"anki-sync-go/internal/anki"
	"crypto/md5"
	"fmt"
	"net/url"
	"os"
	"regexp"
	"slices"
	"unicode/utf8"

	"gopkg.in/yaml.v3"
)

var illegalChars = regexp.MustCompile(`[\\/?%*:|"<>\[\]]`)

// ParseYaml reads and parses a YAML draft file from the given file path.
func ParseYaml(filePath string) (*Draft, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read draft file %s: %w", filePath, err)
	}

	var draft Draft
	if err := yaml.Unmarshal(data, &draft); err != nil {
		return nil, fmt.Errorf("failed to parse yaml draft: %w", err)
	}

	if draft.Profile == "" {
		return nil, fmt.Errorf("draft missing required 'profile' field")
	}

	if len(draft.Cards) == 0 {
		return nil, fmt.Errorf("draft has no cards to import")
	}

	return &draft, nil
}

func buildTTSURL(text, lang string) string {
	if text == "" {
		return ""
	}
	baseURL := "https://translate.google.com/translate_tts"
	params := url.Values{}
	params.Add("ie", "UTF-8")
	params.Add("q", text)
	params.Add("tl", lang)
	params.Add("client", "tw-ob")

	return fmt.Sprintf("%s?%s", baseURL, params.Encode())
}

func sanitizeFilename(text, prefix string) string {
	safe := illegalChars.ReplaceAllString(text, "_")
	if utf8.RuneCountInString(safe) > 50 {
		runes := []rune(safe)
		safe = string(runes[:50])
	}
	return fmt.Sprintf("%s_%s.mp3", prefix, safe)
}

func calculateSkipHash(text string) string {
	h := md5.New()
	h.Write([]byte(text))
	return fmt.Sprintf("%x", h.Sum(nil))
}

// BuildNotes converts draft cards into anki.NoteToAdd with audio metadata.
func BuildNotes(draft *Draft, profile *Profile) ([]anki.NoteToAdd, error) {
	if draft == nil || profile == nil {
		return nil, fmt.Errorf("draft and profile cannot be nil")
	}

	lang := profile.TargetLang
	if lang == "" {
		lang = "ja" // Default fallback
	}

	// Sort audio field keys for deterministic order
	audioKeys := make([]string, 0, len(profile.Audio))
	for k := range profile.Audio {
		audioKeys = append(audioKeys, k)
	}
	slices.Sort(audioKeys)

	notes := make([]anki.NoteToAdd, 0, len(draft.Cards))
	for _, card := range draft.Cards {
		fields := make(map[string]string, len(card))
		for k, v := range card {
			fields[k] = v
		}

		var audioList []anki.AudioInfo
		for _, ankiAudioField := range audioKeys {
			sourceField := profile.Audio[ankiAudioField]
			text := card[sourceField]
			if text == "" {
				continue
			}

			prefix := fmt.Sprintf("%s_%s", lang, ankiAudioField)
			filename := sanitizeFilename(text, prefix)
			audioInfo := anki.AudioInfo{
				URL:            buildTTSURL(text, lang),
				Filename:       filename,
				Fields:         []string{ankiAudioField},
				SkipHash:       calculateSkipHash(text),
				DeleteExisting: true,
			}
			audioList = append(audioList, audioInfo)
		}

		note := anki.NoteToAdd{
			DeckName:  profile.Deck,
			ModelName: profile.Model,
			Fields:    fields,
			Audio:     audioList,
			Tags:      []string{"anki-sync"},
			Options: map[string]any{
				"allowDuplicate": false,
			},
		}
		notes = append(notes, note)
	}

	return notes, nil
}

type ImportResult struct {
	Profile   string
	Deck      string
	Model     string
	CardCount int
	DryRun    bool
	Notes     []anki.NoteToAdd
}

// Import reads a draft YAML file, resolves its profile, builds notes with audio,
// and imports them into AnkiConnect (or previews them if dryRun is true).
func Import(draftPath string, dryRun bool) (*ImportResult, error) {
	draft, err := ParseYaml(draftPath)
	if err != nil {
		return nil, err
	}

	profile, err := LoadProfile(".profiles.json", draft.Profile)
	if err != nil {
		return nil, err
	}

	notes, err := BuildNotes(draft, profile)
	if err != nil {
		return nil, err
	}

	result := &ImportResult{
		Profile:   draft.Profile,
		Deck:      profile.Deck,
		Model:     profile.Model,
		CardCount: len(notes),
		DryRun:    dryRun,
		Notes:     notes,
	}

	if dryRun {
		return result, nil
	}

	// Real run: connect to AnkiConnect and add notes
	client := anki.NewClient()
	if err := client.Check(); err != nil {
		return nil, fmt.Errorf("anki connect is not accessible: %w (make sure Anki is open with AnkiConnect enabled)", err)
	}

	if err := client.AddNotes(notes); err != nil {
		return nil, fmt.Errorf("failed to add notes to Anki: %w", err)
	}

	// Remove draft file
	if err := os.Remove(draftPath); err != nil {
		return nil, fmt.Errorf("failed to remove anki-draft file: %w", err)
	}
	fmt.Println("✔ Removed draft ", draftPath)

	return result, nil
}

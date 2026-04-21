package config

import (
	"encoding/json"
	"fmt"
	"os"
	"slices"
)

// AnkiProvider defines what we need from Anki to validate our config.
type AnkiProvider interface {
	GetModels() ([]string, error)
	GetModelFields(modelName string) ([]string, error)
	GetDeckNames() ([]string, error)
}

type FieldMapping map[string]string

type Config struct {
	SheetCSVURL        string       `json:"sheet_csv_url"`
	AnkiDeck           string       `json:"anki_deck"`
	AnkiModel          string       `json:"anki_model"`
	Mapping            FieldMapping `json:"mapping"`
	AudioField         string       `json:"audio_field"`
	SentenceAudioField string       `json:"sentence_audio_field"`
	SentenceSourceCol  string       `json:"sentence_source_col"`
	ForceSync          bool         `json:"-"`
	DryRun             bool         `json:"-"`
}

// Validate checks if the local configuration fields are present.
func (c *Config) Validate() error {
	if c.SheetCSVURL == "" {
		return fmt.Errorf("sheet CSV URL is required")
	}
	if c.AnkiDeck == "" {
		return fmt.Errorf("anki deck is required")
	}
	if c.AnkiModel == "" {
		return fmt.Errorf("anki model is required")
	}
	if len(c.Mapping) == 0 {
		return fmt.Errorf("mapping is required")
	}
	if c.AudioField == "" {
		return fmt.Errorf("audio field is required")
	}
	if c.SentenceAudioField == "" {
		return fmt.Errorf("sentence audio field is required")
	}
	if c.SentenceSourceCol == "" {
		return fmt.Errorf("sentence source column is required")
	}
	return nil
}

// ValidateWithAnki checks if the configured deck, model, and fields exist in Anki.
func (c *Config) ValidateWithAnki(anki AnkiProvider) error {
	// Check Deck
	decks, err := anki.GetDeckNames()
	if err != nil {
		return fmt.Errorf("failed to fetch decks: %w", err)
	}
	if !slices.Contains(decks, c.AnkiDeck) {
		return fmt.Errorf("deck '%s' not found in Anki", c.AnkiDeck)
	}

	// Check Model
	models, err := anki.GetModels()
	if err != nil {
		return fmt.Errorf("failed to fetch models: %w", err)
	}
	if !slices.Contains(models, c.AnkiModel) {
		return fmt.Errorf("model '%s' not found in Anki", c.AnkiModel)
	}

	// Check Fields
	modelFields, err := anki.GetModelFields(c.AnkiModel)
	if err != nil {
		return fmt.Errorf("failed to fetch fields for model %s: %w", c.AnkiModel, err)
	}
	for _, ankiField := range c.Mapping {
		if !slices.Contains(modelFields, ankiField) {
			return fmt.Errorf("field '%s' not found in Anki model '%s'", ankiField, c.AnkiModel)
		}
	}
	return nil
}

// ValidateWithSheet checks if the mapping keys exist as columns.
func (c *Config) ValidateWithSheet(headers []string) error {
	for sheetCol := range c.Mapping {
		if !slices.Contains(headers, sheetCol) {
			return fmt.Errorf("column '%s' not found in Google Sheet", sheetCol)
		}
	}
	return nil
}

func (c *Config) String() string {
	return fmt.Sprintf(`Config{
		SheetCSVURL: %s,
		AnkiDeck: %s,
		AnkiModel: %s,
		Mapping: %v,
		AudioField: %s,
		SentenceAudioField: %s,
		SentenceSourceCol: %s,
	}`, c.SheetCSVURL, c.AnkiDeck, c.AnkiModel, c.Mapping, c.AudioField, c.SentenceAudioField, c.SentenceSourceCol)
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}
	return &config, nil
}

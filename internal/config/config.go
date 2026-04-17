package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type Validator interface {
	Validate() error
}

type Config struct {
	SheetCSVURL        string            `json:"sheet_csv_url"`
	AnkiDeck           string            `json:"anki_deck"`
	AnkiModel          string            `json:"anki_model"`
	Mapping            map[string]string `json:"mapping"`
	AudioField         string            `json:"audio_field"`
	SentenceAudioField string            `json:"sentence_audio_field"`
	SentenceSourceCol  string            `json:"sentence_source_col"`
	ForceSync          bool              `json:"-"`
	DryRun             bool              `json:"-"`
}	

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


func Check(v Validator) error {
	if err := v.Validate(); err != nil {
		return err
	}
	return nil
}
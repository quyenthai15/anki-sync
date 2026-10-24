package importer

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Profile struct {
	Deck             string            `json:"deck"`
	Model            string            `json:"model"`
	TargetLang       string            `json:"target_lang"`
	TranslationLang  string            `json:"translation_lang"`
	Level            string            `json:"level"`
	Pronunciation    string            `json:"pronunciation"`
	Rules            string            `json:"rules"`
	SentenceField    string            `json:"sentence_field"`
	TranslationField string            `json:"translation_field"`
	Fields           map[string]string `json:"fields"`
	Audio            map[string]string `json:"audio"` // "Audio": "Expression", "SentenceAudio": "Sentence_JP"
}

type ProfilesConfig struct {
	Profiles map[string]Profile `json:"profiles"`
}

// LoadProfiles reads a profiles configuration file (e.g. .profiles.json or profiles.json).
func LoadProfiles(path string) (map[string]Profile, error) {
	var candidates []string
	if path != "" {
		candidates = append(candidates, path, filepath.Join("..", path), filepath.Join("..", "..", path))
	}
	// Default candidates in order of preference
	defaults := []string{".profiles.json", "profiles.json"}
	for _, d := range defaults {
		candidates = append(candidates, d, filepath.Join("..", d), filepath.Join("..", "..", d))
	}

	var data []byte
	for _, c := range candidates {
		if d, err := os.ReadFile(c); err == nil {
			data = d
			break
		}
	}

	if data == nil {
		return nil, fmt.Errorf("failed to find profiles file (checked .profiles.json, profiles.json)")
	}

	var cfg ProfilesConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse profiles json: %w", err)
	}

	if len(cfg.Profiles) == 0 {
		return nil, fmt.Errorf("no profiles found in configuration")
	}

	return cfg.Profiles, nil
}

// LoadProfile loads a specific profile by name from the given file path (or default).
func LoadProfile(path, profileName string) (*Profile, error) {
	profiles, err := LoadProfiles(path)
	if err != nil {
		return nil, err
	}

	p, ok := profiles[profileName]
	if !ok {
		return nil, fmt.Errorf("profile '%s' not found in profiles config", profileName)
	}

	return &p, nil
}

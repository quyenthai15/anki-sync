package audio

import (
	"fmt"
	"net/url"
	"regexp"
	"unicode/utf8"
)

type GoogleTTS struct {
	Language string
}

func NewGoogleTTS(lang string) *GoogleTTS {
	if lang == "" {
		lang = "ja" // Default to Japanese
	}
	return &GoogleTTS{Language: lang}
}

func (g *GoogleTTS) GetTts(text string) string {
	if text == "" {
		return ""
	}
	baseURL := "https://translate.google.com/translate_tts"
	params := url.Values{}
	params.Add("ie", "UTF-8")
	params.Add("q", text)
	params.Add("tl", g.Language)
	params.Add("client", "tw-ob")

	return fmt.Sprintf("%s?%s", baseURL, params.Encode())
}

var illegalChars = regexp.MustCompile(`[\\/?%*:|"<>\[\]]`)

// SanitizeFilename creates a safe filename for Anki.
func SanitizeFilename(name, prefix string) string {
	// Replace illegal characters with "_"
	safe := illegalChars.ReplaceAllString(name, "_")

	// Truncate to 50 characters safely using runes
	if utf8.RuneCountInString(safe) > 50 {
		runes := []rune(safe)
		safe = string(runes[:50])
	}

	return fmt.Sprintf("%s_%s.mp3", prefix, safe)
}

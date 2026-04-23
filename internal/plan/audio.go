package plan

import (
	"anki-sync-go/internal/anki"
	"crypto/md5"
	"fmt"
	"net/url"
	"regexp"
	"unicode/utf8"
)

type TTSProvider interface {
	GetTts(text string) string
}

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

// calculateSkipHash generates MD5 hash of the text content for skipHash
func calculateSkipHash(text string) string {
	h := md5.New()
	h.Write([]byte(text))
	return fmt.Sprintf("%x", h.Sum(nil))
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

type AudioProcessor struct {
	tts    TTSProvider
	prefix string
}

func NewAudioProcessor(tts TTSProvider, prefix string) *AudioProcessor {
	return &AudioProcessor{
		tts:    tts,
		prefix: prefix,
	}
}

// CreateAudioInfo generates anki.AudioInfo and the corresponding [sound:filename] tag.
func (p *AudioProcessor) CreateAudioInfo(text string, fields []string) (info anki.AudioInfo, soundTag string) {
	filename := SanitizeFilename(text, p.prefix)
	info = anki.AudioInfo{
		URL:            p.tts.GetTts(text),
		Filename:       filename,
		Fields:         fields,
		SkipHash:       calculateSkipHash(text), // Use text content for consistent hash
		DeleteExisting: true,
	}
	soundTag = fmt.Sprintf("[sound:%s]", filename)
	return info, soundTag
}

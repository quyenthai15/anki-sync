package plan

import (
	"testing"
)

type mockTTS struct{}

func (m *mockTTS) GetTts(text string) string {
	return "http://example.com/tts?q=" + text
}

func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		prefix   string
		expected string
	}{
		{"Normal", "hello", "pref", "pref_hello.mp3"},
		{"Illegal chars", "hello/world?", "pref", "pref_hello_world_.mp3"},
		{"Too long", "this_is_a_very_long_filename_that_should_be_truncated_at_some_point", "pref", "pref_this_is_a_very_long_filename_that_should_be_trunca.mp3"},
		{"Japanese", "日本語", "ja", "ja_日本語.mp3"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SanitizeFilename(tt.input, tt.prefix)
			if got != tt.expected {
				t.Errorf("SanitizeFilename() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestAudioProcessor_CreateAudioInfo(t *testing.T) {
	ap := NewAudioProcessor(&mockTTS{}, "test")
	info, tag := ap.CreateAudioInfo("hello", []string{"Field1"})

	expectedTag := "[sound:test_hello.mp3]"
	if tag != expectedTag {
		t.Errorf("Expected tag %s, got %s", expectedTag, tag)
	}

	if info.Filename != "test_hello.mp3" {
		t.Errorf("Expected filename test_hello.mp3, got %s", info.Filename)
	}

	if info.URL != "http://example.com/tts?q=hello" {
		t.Errorf("Expected URL to match mock, got %s", info.URL)
	}
}

package main

import "testing"

func TestResolveDeck(t *testing.T) {
	tests := []struct {
		name    string
		deck    string
		profile string
		want    string
		wantErr bool
	}{
		{name: "deck only", deck: "French::A1", want: "French::A1"},
		{name: "neither", wantErr: true},
		{name: "both", deck: "French::A1", profile: "french", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolveDeck(tt.deck, tt.profile)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got deck %q", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestBuildQuery(t *testing.T) {
	tests := []struct {
		deck string
		word string
		want string
	}{
		{deck: "French::A1", word: "le livre", want: `deck:"French::A1" "le livre"`},
		{deck: "Japanese::N5", word: "卵", want: `deck:"Japanese::N5" "卵"`},
		{deck: "French::A1", word: "Expression:livre", want: `deck:"French::A1" Expression:livre`},
		{deck: "French::A1", word: `Front:"le livre"`, want: `deck:"French::A1" Front:"le livre"`},
		{deck: "A B", word: `he said "hi"`, want: `deck:"A B" "he said \"hi\""`},
	}

	for _, tt := range tests {
		if got := buildQuery(tt.deck, tt.word); got != tt.want {
			t.Errorf("buildQuery(%q, %q) = %q, want %q", tt.deck, tt.word, got, tt.want)
		}
	}
}

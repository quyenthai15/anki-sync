package main

import (
	"anki-sync-go/internal/anki"
	"anki-sync-go/internal/importer"
	"flag"
	"fmt"
	"os"
	"strings"
)

const maxShownNotes = 10

type searchResult struct {
	word string
	ids  []int
}

func handleSearch(args []string) {
	searchFlags := flag.NewFlagSet("search", flag.ExitOnError)
	deck := searchFlags.String("deck", "", "Deck to search")
	profile := searchFlags.String("profile", "", "Profile whose deck to search")
	field := searchFlags.String("field", "", "Specific field to search")
	_ = searchFlags.Parse(args)

	words := searchFlags.Args()
	if len(words) == 0 {
		fmt.Fprintln(os.Stderr, "✘ Search needs at least one word")
		os.Exit(1)
	}

	if *field != "" {
		for i, w := range words {
			words[i] = fmt.Sprintf("%s:%s", *field, w)
		}
	}

	deckName, err := resolveDeck(*deck, *profile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "✘ %v\n", err)
		os.Exit(1)
	}

	results, err := searchWords(anki.NewClient(), deckName, words)
	if err != nil {
		fmt.Fprintf(os.Stderr, "✘ Search failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Searching %s...\n\n", deckName)
	printResults(deckName, results)
}

func resolveDeck(deck, profile string) (string, error) {
	if (deck == "") == (profile == "") {
		return "", fmt.Errorf("use exactly one of --deck or --profile")
	}
	if profile == "" {
		return deck, nil
	}

	p, err := importer.LoadProfile("", profile)
	if err != nil {
		return "", err
	}
	return p.Deck, nil
}

func buildQuery(deck, word string) string {
	if strings.Contains(word, ":") {
		return fmt.Sprintf("deck:%q %s", deck, word)
	}
	return fmt.Sprintf("deck:%q %q", deck, word)
}

func searchWords(client *anki.Client, deck string, words []string) ([]searchResult, error) {
	results := make([]searchResult, 0, len(words))
	for _, word := range words {
		ids, err := client.FindNotes(buildQuery(deck, word))
		if err != nil {
			return nil, fmt.Errorf("%q: %w", word, err)
		}
		results = append(results, searchResult{word: word, ids: ids})
	}
	return results, nil
}

func printResults(deck string, results []searchResult) {
	found := 0
	for _, result := range results {
		switch len(result.ids) {
		case 0:
			fmt.Printf("✘ %s — not found\n", result.word)
		case 1:
			found++
			fmt.Printf("✔ %s — note #%d\n", result.word, result.ids[0])
		default:
			found++
			fmt.Printf("✔ %s — %d notes (%s)\n", result.word, len(result.ids), noteIDs(result.ids))
		}
	}
	fmt.Printf("\n%d/%d words found in %s.\n", found, len(results), deck)
}

func noteIDs(ids []int) string {
	shown := ids
	suffix := ""
	if len(ids) > maxShownNotes {
		shown = ids[:maxShownNotes]
		suffix = fmt.Sprintf(", +%d more", len(ids)-maxShownNotes)
	}

	parts := make([]string, len(shown))
	for i, id := range shown {
		parts[i] = fmt.Sprintf("#%d", id)
	}
	return strings.Join(parts, ", ") + suffix
}

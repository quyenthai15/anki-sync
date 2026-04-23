package executor

import (
	"anki-sync-go/internal/anki"
	"anki-sync-go/internal/config"
	"anki-sync-go/internal/plan"
	"fmt"
)

type Executor struct {
	ankiClient AnkiClient
	config     *config.Config
}

type AnkiClient interface {
	AddNotes(notes []anki.NoteToAdd) ([]int, error)
	UpdateNotes(notes []any) error
}

func NewExecutor(client AnkiClient, cfg *config.Config) *Executor {
	return &Executor{
		ankiClient: client,
		config:     cfg,
	}
}

// Execute performs the actual updates and additions to Anki
func (e *Executor) Execute(p *plan.Plan) error {
	if err := e.ExecuteAdditions(p.NewNotes); err != nil {
		return fmt.Errorf("failed to execute additions: %w", err)
	}

	if err := e.ExecuteUpdates(p.UpdatedNotes); err != nil {
		return fmt.Errorf("failed to execute updates: %w", err)
	}

	return nil
}

func (e *Executor) ExecuteAdditions(notes []plan.NoteToAdd) error {
	if len(notes) == 0 {
		return nil
	}
	ankiNotes := make([]anki.NoteToAdd, 0, len(notes))
	for _, note := range notes {
		ankiNote, err := e.buildNoteToAdd(note)
		if err != nil {
			return fmt.Errorf("fail to build new note %w", err)
		}
		ankiNotes = append(ankiNotes, ankiNote)
	}
	if _, err := e.ankiClient.AddNotes(ankiNotes); err != nil {
		return fmt.Errorf("fail to add new notes to anki %w", err)
	}
	return nil
}

func (e *Executor) ExecuteUpdates(notes []plan.NoteToUpdate) error {
	for _, note := range notes {
		// TODO: Implement actual AnkiConnect call in internal/anki/anki.go
		// and use it here.
		fmt.Printf("Updating note %d: %v\n", note.NoteID, note.Fields)
	}
	return nil
}

func (e *Executor) buildNoteToAdd(note plan.NoteToAdd) (anki.NoteToAdd, error) {
	ankiNote := anki.NoteToAdd{
		DeckName:  e.config.AnkiDeck,
		ModelName: e.config.AnkiModel,
		Audio:     note.Audio,
		Fields:    note.Fields,
		Tags:      []string{"csv-sync"},
		Options: map[string]any{
			"allowDuplicate": false,
		},
	}

	return ankiNote, nil
}

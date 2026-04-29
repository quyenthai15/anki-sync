package executor

import (
	"anki-sync-go/internal/anki"
	"anki-sync-go/internal/config"
	"anki-sync-go/internal/plan"
	"anki-sync-go/internal/status"
	"fmt"
)

type Reporter interface {
	Report(status.Event)
}

type Executor struct {
	ankiClient AnkiClient
	config     config.Config
	reporter   Reporter
}

type AnkiClient interface {
	AddNotes(notes []anki.NoteToAdd) ([]int, error)
	UpdateNotes(notes []any) error
}

func NewExecutor(client AnkiClient, cfg config.Config, r Reporter) *Executor {
	return &Executor{
		ankiClient: client,
		config:     cfg,
		reporter:   r,
	}
}

func (e *Executor) report(msg status.Event) {
	if e.reporter != nil {
		e.reporter.Report(msg)
	}
}

// Execute performs the actual updates and additions to Anki
func (e *Executor) Execute(p plan.Plan) error {
	total := len(p.NewNotes) + len(p.UpdatedNotes)
	current := 0

	if len(p.NewNotes) > 0 {
		e.report(status.ProgressMsg{Current: current, Total: total})
		e.report(status.LogMsg{Text: "Adding notes"})
		if err := e.ExecuteAdditions(p.NewNotes); err != nil {
			return fmt.Errorf("failed to execute additions: %w", err)
		}
		current += len(p.NewNotes)
		e.report(status.ProgressMsg{Current: current, Total: total})
	}

	if len(p.UpdatedNotes) > 0 {
		e.report(status.ProgressMsg{Current: current, Total: total})
		e.report(status.LogMsg{Text: "Updating notes"})
		if err := e.ExecuteUpdates(p.UpdatedNotes); err != nil {
			return fmt.Errorf("failed to execute updates: %w", err)
		}
		current += len(p.UpdatedNotes)
		e.report(status.ProgressMsg{Current: current, Total: total})
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
	if len(notes) == 0 {
		return nil
	}
	var actions []any
	for _, note := range notes {
		// Update note fields and audio
		// To add audio to an existing note, we can use 'updateNoteFields'
		// if the audio tag is already in the field value. [sound:ja_word_tts_ほんとに.mp3]
		// Ensure fields is not nil for AnkiConnect
		fields := note.Fields
		if fields == nil {
			fields = make(map[string]string)
		}

		action := map[string]any{
			"action": "updateNoteFields",
			"params": map[string]any{
				"note": map[string]any{
					"id":     note.NoteID,
					"fields": fields,
					"audio":  note.Audio,
				},
			},
		}
		actions = append(actions, action)
	}

	if len(actions) > 0 {
		return e.ankiClient.UpdateNotes(actions)
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

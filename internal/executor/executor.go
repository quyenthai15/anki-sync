package executor

import (
	"anki-sync-go/internal/anki"
	"anki-sync-go/internal/plan"
	"fmt"
)

type Executor struct {
	ankiClient *anki.Client
}

func NewExecutor(client *anki.Client) *Executor {
	return &Executor{
		ankiClient: client,
	}
}

// Execute performs the actual updates and additions to Anki
func (e *Executor) Execute(p *plan.SyncPlan) error {
	if err := e.ExecuteAdditions(p.NewNotes); err != nil {
		return fmt.Errorf("failed to execute additions: %w", err)
	}

	if err := e.ExecuteUpdates(p.UpdatedNotes); err != nil {
		return fmt.Errorf("failed to execute updates: %w", err)
	}

	return nil
}

func (e *Executor) ExecuteAdditions(notes []plan.NoteToAdd) error {
	for _, note := range notes {
		// TODO: Implement actual AnkiConnect call in internal/anki/anki.go
		// and use it here. For now, we just log the intent.
		fmt.Printf("Adding note: %v\n", note.Fields)
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

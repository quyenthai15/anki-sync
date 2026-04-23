package plan

import (
	"anki-sync-go/internal/anki"
	"anki-sync-go/internal/config"
)

// NoteToAdd represents the data needed to create a new note in Anki
type NoteToAdd struct {
	Fields map[string]string
	Audio  []anki.AudioInfo
}

// NoteToUpdate represents a change to an existing Anki note
type NoteToUpdate struct {
	NoteID int
	Fields map[string]string // Only the fields that actually changed
	Audio  []anki.AudioInfo
}

// SyncPlan holds the pending changes identified during the diffing phase
type SyncPlan struct {
	Config       *config.Config
	NewNotes     []NoteToAdd
	UpdatedNotes []NoteToUpdate
}

func NewSyncPlan(cfg *config.Config) *SyncPlan {
	return &SyncPlan{
		Config: cfg,
	}
}

// AddNewNote adds a new note to the plan. Fields must already be mapped to Anki field names.
func (p *SyncPlan) AddNewNote(ankiFields map[string]string, audio []anki.AudioInfo) {
	p.NewNotes = append(p.NewNotes, NoteToAdd{Fields: ankiFields, Audio: audio})
}

// AddUpdatedNote adds an update to an existing note to the plan
func (p *SyncPlan) AddUpdatedNote(id int, changedFields map[string]string, audio []anki.AudioInfo) {
	p.UpdatedNotes = append(p.UpdatedNotes, NoteToUpdate{NoteID: id, Fields: changedFields, Audio: audio})
}

// Diff compares target fields with existing fields.
// It returns a map of fields that need updating, or nil if they are identical.
func Diff(target, existing map[string]string) map[string]string {
	diff := make(map[string]string)
	for key, targetVal := range target {
		existingVal, exists := existing[key]
		if !exists {
			// ignore extra sheet columns
			continue
		}
		if targetVal != existingVal {
			diff[key] = targetVal
		}
	}

	if len(diff) == 0 {
		return nil
	}
	return diff
}

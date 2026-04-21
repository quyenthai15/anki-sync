package plan

import (
	"anki-sync-go/internal/anki"
	"anki-sync-go/internal/sheet"
)

// NoteToAdd represents the data needed to create a new note in Anki
type NoteToAdd struct {
	Fields map[string]string
}

// NoteToUpdate represents a change to an existing Anki note
type NoteToUpdate struct {
	NoteID int
	Fields map[string]string // Only the fields that actually changed
}

// SyncPlan holds the pending changes identified during the diffing phase
type SyncPlan struct {
	NewNotes     []NoteToAdd
	UpdatedNotes []NoteToUpdate
}

// AddNewNote adds a new note to the plan
func (p *SyncPlan) AddNewNote(fields map[string]string) {
	p.NewNotes = append(p.NewNotes, NoteToAdd{Fields: fields})
}

// AddUpdatedNote adds an update to an existing note to the plan
func (p *SyncPlan) AddUpdatedNote(id int, fields map[string]string) {
	p.UpdatedNotes = append(p.UpdatedNotes, NoteToUpdate{NoteID: id, Fields: fields})
}

// GetDiff compares a sheet record with an existing Anki note.
// It returns a map of fields that are different, or nil if they are identical.
func GetDiff(sheetRecord sheet.Record, ankiRecord anki.NormalizedNote) map[string]string {
	diff := make(map[string]string)

	for key, sheetVal := range sheetRecord {
		ankiVal, exists := ankiRecord.Fields[key]
		// If field is missing in Anki or values differ, mark for update
		if !exists || sheetVal != ankiVal {
			diff[key] = sheetVal
		}
	}

	if len(diff) == 0 {
		return nil
	}
	return diff
}

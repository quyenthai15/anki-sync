package plan

import (
	"anki-sync-go/internal/anki"
	"anki-sync-go/internal/config"
	"anki-sync-go/internal/sheet"
	"fmt"
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
	Config       *config.Config
	NewNotes     []NoteToAdd
	UpdatedNotes []NoteToUpdate
}

func NewSyncPlan(cfg *config.Config) *SyncPlan {
	return &SyncPlan{
		Config: cfg,
	}
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
func (p *SyncPlan) GetDiff(sheetRecord sheet.Record, ankiRecord anki.NormalizedNote) map[string]string {
	diff := make(map[string]string)
	for key, sheetVal := range sheetRecord {
		ankiVal, exist := ankiRecord.Fields[p.Config.Mapping[key]]
		if !exist {
			// ignore extra sheet columns
			continue
		}
		if ankiVal != sheetVal {
			diff[key] = sheetVal
			fmt.Printf("Field %s differs: sheet=%s, anki=%s\n", key, sheetVal, ankiVal)
			// Check if sentence source is there
			if p.Config.SentenceAudioField != "" && key == p.Config.SentenceSourceCol {
				diff[p.Config.SentenceAudioField] = "[sentence_audio]"
			}
		}
	}

	if len(diff) == 0 {
		return nil
	}
	return diff
}

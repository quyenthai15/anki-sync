package plan

import (
	"anki-sync-go/internal/anki"
	"anki-sync-go/internal/audio"
	"anki-sync-go/internal/config"
	"anki-sync-go/internal/sheet"
	"strings"
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

// Plan holds the pending changes identified during the diffing phase
type Plan struct {
	NewNotes     []NoteToAdd
	UpdatedNotes []NoteToUpdate
}

func NewPlan() *Plan {
	return &Plan{}
}

func (p *Plan) AddNewNote(ankiFields map[string]string, audio []anki.AudioInfo) {
	p.NewNotes = append(p.NewNotes, NoteToAdd{Fields: ankiFields, Audio: audio})
}

func (p *Plan) AddUpdatedNote(id int, changedFields map[string]string, audio []anki.AudioInfo) {
	p.UpdatedNotes = append(p.UpdatedNotes, NoteToUpdate{NoteID: id, Fields: changedFields, Audio: audio})
}

type PlanBuilder struct {
	config        *config.Config
	wordAudio     *audio.AudioProcessor
	sentenceAudio *audio.AudioProcessor
	existingNotes map[string]anki.NormalizedNote
	wordCol       string
	plan          *Plan
}

func NewPlanBuilder(
	cfg *config.Config,
	wordAudio *audio.AudioProcessor,
	sentenceAudio *audio.AudioProcessor,
	existingNotes map[string]anki.NormalizedNote,
	primaryCol string,
) *PlanBuilder {
	return &PlanBuilder{
		config:        cfg,
		wordAudio:     wordAudio,
		sentenceAudio: sentenceAudio,
		existingNotes: existingNotes,
		wordCol:       primaryCol,
		plan:          NewPlan(),
	}
}

func (b *PlanBuilder) Process(r sheet.Record) error {
	targetFields := b.config.MapRecord(r)
	word := r[b.wordCol]

	existing, found := b.existingNotes[word]
	var audioInfo []anki.AudioInfo

	if !found {
		// New note - Word Audio
		if b.config.AudioField != "" && word != "" {
			info, tag := b.wordAudio.CreateAudioInfo(word, []string{b.config.AudioField})
			audioInfo = append(audioInfo, info)
			targetFields[b.config.AudioField] = tag
		}
		// New note - Sentence Audio
		if b.config.SentenceAudioField != "" && b.config.SentenceSourceCol != "" {
			sentence := targetFields[b.config.SentenceSourceCol]
			if sentence != "" {
				info, tag := b.sentenceAudio.CreateAudioInfo(sentence, []string{b.config.SentenceAudioField})
				audioInfo = append(audioInfo, info)
				targetFields[b.config.SentenceAudioField] = tag
			}
		}
		b.plan.AddNewNote(targetFields, audioInfo)
	} else if b.config.ForceSync {
		// Update note - Word Audio (if missing)
		if b.config.AudioField != "" && word != "" {
			if !hasSound(existing.Fields[b.config.AudioField]) {
				info, tag := b.wordAudio.CreateAudioInfo(word, []string{b.config.AudioField})
				audioInfo = append(audioInfo, info)
				targetFields[b.config.AudioField] = tag
			}
		}
		// Update note - Sentence Audio (if missing or sentence changed)
		if b.config.SentenceAudioField != "" && b.config.SentenceSourceCol != "" {
			sentence := targetFields[b.config.SentenceSourceCol]
			sentenceChanged := existing.Fields[b.config.SentenceSourceCol] != sentence

			if sentence != "" && (!hasSound(existing.Fields[b.config.SentenceAudioField]) || sentenceChanged) {
				info, tag := b.sentenceAudio.CreateAudioInfo(sentence, []string{b.config.SentenceAudioField})
				audioInfo = append(audioInfo, info)
				targetFields[b.config.SentenceAudioField] = tag
			}
		}

		diff := Diff(targetFields, existing.Fields)
		if diff != nil || len(audioInfo) > 0 {
			b.plan.AddUpdatedNote(existing.NoteID, diff, audioInfo)
		}
	}
	return nil
}

func (b *PlanBuilder) GetPlan() *Plan {
	return b.plan
}

func hasSound(val string) bool {
	return strings.Contains(val, "[sound:")
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

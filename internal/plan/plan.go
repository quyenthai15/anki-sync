package plan

import (
	"anki-sync-go/internal/anki"
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
	wordAudio     *AudioProcessor
	sentenceAudio *AudioProcessor
	existingNotes map[string]anki.NormalizedNote
	wordCol       string
	plan          *Plan
}

func NewPlanBuilder(
	cfg *config.Config,
	tts TTSProvider,
	existingNotes map[string]anki.NormalizedNote,
	primaryCol string,
) *PlanBuilder {
	wordAudio := NewAudioProcessor(tts, "ja_word_tts")
	sentenceAudio := NewAudioProcessor(tts, "ja_sent_tts")
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
	if !found {
		return b.processNewNote(word, targetFields)
	}

	return b.processUpdateNote(word, targetFields, existing)
}

func (b *PlanBuilder) processNewNote(word string, fields map[string]string) error {
	var audios []anki.AudioInfo

	if audio, ok := b.createWordAudio(word); ok {
		audios = append(audios, audio)
	}

	sentence := fields[b.config.SentenceSourceCol]
	if audio, ok := b.createSentenceAudio(sentence); ok {
		audios = append(audios, audio)
	}

	b.plan.AddNewNote(fields, audios)
	return nil
}

func (b *PlanBuilder) processUpdateNote(word string, fields map[string]string, existing anki.NormalizedNote) error {
	if !b.config.ForceSync {
		return nil
	}

	var audios []anki.AudioInfo

	// Handle Word Audio - only if missing
	if !hasSound(existing.Fields[b.config.AudioField]) {
		if audio, ok := b.createWordAudio(word); ok {
			audios = append(audios, audio)
		}
	}

	// Handle Sentence Audio - if missing or sentence changed
	sentence := fields[b.config.SentenceSourceCol]
	sentenceChanged := existing.Fields[b.config.SentenceSourceCol] != sentence
	if sentence != "" && (!hasSound(existing.Fields[b.config.SentenceAudioField]) || sentenceChanged) {
		if audio, ok := b.createSentenceAudio(sentence); ok {
			audios = append(audios, audio)
		}
	}

	diff := Diff(fields, existing.Fields)
	if diff != nil || len(audios) > 0 {
		b.plan.AddUpdatedNote(existing.NoteID, diff, audios)
	}

	return nil
}

func (b *PlanBuilder) createWordAudio(word string) (info anki.AudioInfo, ok bool) {
	if b.config.AudioField == "" || word == "" {
		return anki.AudioInfo{}, false
	}
	info, _ = b.wordAudio.CreateAudioInfo(word, []string{b.config.AudioField})
	ok = true
	return
}

func (b *PlanBuilder) createSentenceAudio(sentence string) (info anki.AudioInfo, ok bool) {
	if b.config.SentenceAudioField == "" || b.config.SentenceSourceCol == "" {
		return anki.AudioInfo{}, false
	}
	if sentence == "" {
		return anki.AudioInfo{}, false
	}
	info, _ = b.sentenceAudio.CreateAudioInfo(sentence, []string{b.config.SentenceAudioField})
	ok = true
	return
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

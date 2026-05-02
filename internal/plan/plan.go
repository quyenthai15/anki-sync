package plan

import (
	"anki-sync-go/internal/anki"
	"anki-sync-go/internal/config"
	"anki-sync-go/internal/sheet"
	"fmt"
	"slices"
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

func NewPlan() Plan {
	return Plan{}
}

func sortedFields(fields map[string]string) []string {
	keys := make([]string, 0, len(fields))
	for k := range fields {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	return keys
}

func (n NoteToAdd) String() string {
	var lines []string
	for _, key := range sortedFields(n.Fields) {
		v := n.Fields[key]
		if v == "" {
			continue
		}
		if len(v) > 40 {
			v = v[:40] + "..."
		}
		lines = append(lines, fmt.Sprintf("   %s: %s ", key, v))
	}
	return strings.Join(lines, "|")
}

func (n NoteToUpdate) String() string {
	var lines []string
	for _, key := range sortedFields(n.Fields) {
		v := n.Fields[key]
		if len(v) > 40 {
			v = v[:40] + "..."
		}
		lines = append(lines, fmt.Sprintf("   %s: %s ", key, v))
	}
	if len(n.Audio) > 0 {
		lines = append(lines, fmt.Sprintf("   audio: %d file(s) ", len(n.Audio)))
	}
	return strings.Join(lines, "|")
}

func (p Plan) String() string {
	var parts []string
	sep := "   " + strings.Repeat("·", 30)
	if len(p.NewNotes) > 0 {
		section := fmt.Sprintf("Adding %d notes:\n", len(p.NewNotes))
		var notes []string
		for _, note := range p.NewNotes {
			notes = append(notes, note.String())
		}
		parts = append(parts, section+strings.Join(notes, "\n"+sep+"\n"))
	}
	if len(p.UpdatedNotes) > 0 {
		section := fmt.Sprintf("Updating %d notes:\n", len(p.UpdatedNotes))
		var notes []string
		for _, note := range p.UpdatedNotes {
			notes = append(notes, note.String())
		}
		parts = append(parts, section+strings.Join(notes, "\n"+sep+"\n"))
	}
	return strings.Join(parts, "\n\n")
}

func (p *Plan) AddNewNote(ankiFields map[string]string, audio []anki.AudioInfo) {
	p.NewNotes = append(p.NewNotes, NoteToAdd{Fields: ankiFields, Audio: audio})
}

func (p *Plan) AddUpdatedNote(id int, changedFields map[string]string, audio []anki.AudioInfo) {
	p.UpdatedNotes = append(p.UpdatedNotes, NoteToUpdate{NoteID: id, Fields: changedFields, Audio: audio})
}

type PlanBuilder struct {
	config        config.Config
	wordAudio     AudioProcessor
	sentenceAudio AudioProcessor
	existingNotes map[string]anki.NormalizedNote
	wordCol       string
	plan          Plan
}

func NewPlanBuilder(
	cfg config.Config,
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
	sentence := r[b.config.SentenceSourceCol]

	existing, found := b.existingNotes[word]
	if !found {
		return b.processNewNote(word, sentence, targetFields)
	}

	return b.processUpdateNote(word, sentence, targetFields, existing)
}

func (b *PlanBuilder) processNewNote(word, sentence string, fields map[string]string) error {
	var audios []anki.AudioInfo

	if audio, ok := b.createWordAudio(word); ok {
		audios = append(audios, audio)
	}

	if audio, ok := b.createSentenceAudio(sentence); ok {
		audios = append(audios, audio)
	}

	b.plan.AddNewNote(fields, audios)
	return nil
}

func (b *PlanBuilder) processUpdateNote(word, sentence string, fields map[string]string, existing anki.NormalizedNote) error {
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
	sentenceAnkiField := b.config.Mapping[b.config.SentenceSourceCol]
	sentenceChanged := existing.Fields[sentenceAnkiField] != sentence
	if sentence != "" && (!hasSound(existing.Fields[b.config.SentenceAudioField]) || sentenceChanged) {
		if audio, ok := b.createSentenceAudio(sentence); ok {
			audios = append(audios, audio)
		}
	}

	diff := getDiff(fields, existing.Fields)
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

func (b *PlanBuilder) GetPlan() Plan {
	return b.plan
}

func hasSound(val string) bool {
	return strings.Contains(val, "[sound:")
}

// Diff compares target fields with existing fields.
// It returns a map of fields that need updating, or nil if they are identical.
func getDiff(target, existing map[string]string) map[string]string {
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

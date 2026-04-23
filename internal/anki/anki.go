package anki

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const (
	AnkiConnectURL = "http://localhost:8765"
)

// Request represents an Anki Connect API request
type Request struct {
	Action  string         `json:"action"`
	Params  map[string]any `json:"params"`
	Version int            `json:"version"`
}

// Response represents an Anki Connect API response
type Response[T any] struct {
	Error  any `json:"error"`
	Result T   `json:"result"`
}

// RawFieldValue represents the structure of Anki field values
type RawFieldValue struct {
	Value string `json:"value"`
	Order int    `json:"order"`
}

// StringValue returns the string value of the field
func (fv RawFieldValue) StringValue() string {
	return fv.Value
}

// NormalizedNote represents the extracted note information with clean field values
type NormalizedNote struct {
	NoteID int               `json:"noteId"`
	Fields map[string]string `json:"fields"`
}

// NoteMap is a helper alias for notes indexed by their primary key (usually the word)
type NoteMap map[string]NormalizedNote

// AudioInfo represents the structured audio data for AnkiConnect
type AudioInfo struct {
	URL      string   `json:"url"`
	Filename string   `json:"filename"`
	Fields   []string `json:"fields"`
}

// NoteToAdd represents an Anki note with proper field structure
type NoteToAdd struct {
	DeckName  string            `json:"deckName"`
	ModelName string            `json:"modelName"`
	Fields    map[string]string `json:"fields"`
	Audio     []AudioInfo       `json:"audio,omitempty"`
	Tags      []string          `json:"tags"`
	Options   map[string]any    `json:"options"`
}

// RawNote represents raw Anki note data
type RawNote struct {
	NoteId    int                      `json:"noteId"`
	Fields    map[string]RawFieldValue `json:"fields"`
	ModelName string                   `json:"modelName"`
	Cards     []int                    `json:"cards"`
	Tags      []string                 `json:"tags"`
	Mod       int                      `json:"mod"`
}

type Client struct {
	url string
}

func invoke[R any](c *Client, action string, params map[string]any) (R, error) {
	var result R
	req := Request{
		Action:  action,
		Params:  params,
		Version: 6,
	}

	jsonReq, err := json.Marshal(req)
	if err != nil {
		return result, err
	}

	resp, err := http.Post(c.url, "application/json", bytes.NewBuffer(jsonReq))
	if err != nil {
		return result, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return result, err
	}
	var response Response[R]
	if err := json.Unmarshal(body, &response); err != nil {
		return result, err
	}

	if response.Error != nil {
		if errStr, ok := response.Error.(string); ok && errStr != "" {
			return result, fmt.Errorf("anki error: %s", errStr)
		}
	}

	return response.Result, nil
}

func NewClient() *Client {
	return &Client{
		url: AnkiConnectURL,
	}
}

func (c *Client) GetModels() ([]string, error) {
	return invoke[[]string](c, "modelNames", map[string]any{})
}

func (c *Client) GetDeckNames() ([]string, error) {
	return invoke[[]string](c, "deckNames", map[string]any{})
}

func (c *Client) GetModelFields(modelName string) ([]string, error) {
	params := map[string]any{"modelName": modelName}
	return invoke[[]string](c, "modelFieldNames", params)
}

func (c *Client) GetNotes(deckName string) ([]int, error) {
	params := map[string]any{"query": fmt.Sprintf("deck:%s", deckName)}
	return invoke[[]int](c, "findNotes", params)
}

func (c *Client) GetNotesInfo(noteIds []int) ([]RawNote, error) {
	params := map[string]any{"notes": noteIds}
	infos, err := invoke[[]RawNote](c, "notesInfo", params)
	if err != nil {
		return nil, err
	}
	if len(infos) == 0 {
		return nil, fmt.Errorf("no notes found: %d", len(noteIds))
	}
	return infos, nil
}

// ExtractFields converts map[string]FieldValue to map[string]string
func ExtractFields(fields map[string]RawFieldValue) map[string]string {
	result := make(map[string]string)
	for name, field := range fields {
		result[name] = field.StringValue()
	}
	return result
}

// GetFieldValue safely extracts a field value from a map of FieldValue
func GetFieldValue(fields map[string]RawFieldValue, fieldName string) (string, error) {
	field, exists := fields[fieldName]
	if !exists {
		return "", fmt.Errorf("field '%s' not found", fieldName)
	}
	return field.StringValue(), nil
}

// NormalizeAnkiNotes retrieves raw Anki note data and extracts field values
func NormalizeAnkiNotes(client *Client, deckName string, keyField string) (notesByWord NoteMap, err error) {
	notes, err := client.GetNotes(deckName)
	if err != nil {
		return nil, err
	}

	if len(notes) == 0 {
		return make(NoteMap), nil
	}

	notesInfo, err := client.GetNotesInfo(notes)
	if err != nil {
		return nil, err
	}

	notesByWord = make(NoteMap, len(notesInfo))
	for _, note := range notesInfo {
		// Extract the key field value
		keyValue, err := GetFieldValue(note.Fields, keyField)
		if err != nil {
			return nil, fmt.Errorf("failed to get key field '%s' for note %d: %w", keyField, note.NoteId, err)
		}

		// Trim and normalize the key value (equivalent to JS .trim().normalize('NFC'))
		keyValue = strings.TrimSpace(keyValue)

		// Extract all field values
		extractedFields := ExtractFields(note.Fields)

		notesByWord[keyValue] = NormalizedNote{
			NoteID: note.NoteId,
			Fields: extractedFields,
		}
	}
	return notesByWord, nil
}

func (c *Client) AddNotes(notes []NoteToAdd) ([]int, error) {
	params := map[string]any{"notes": notes}
	return invoke[[]int](c, "addNotes", params)
}

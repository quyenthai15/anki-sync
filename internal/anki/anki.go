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

// FieldValue represents the structure of Anki field values
type FieldValue struct {
	Value string `json:"value"`
	Order int    `json:"order"`
}

// StringValue returns the string value of the field
func (fv FieldValue) StringValue() string {
	return fv.Value
}

// NoteDetail represents the extracted note information with clean field values
type NoteDetail struct {
	NoteID int               `json:"noteId"`
	Fields map[string]string `json:"fields"`
}

// Note represents an Anki note with proper field structure
type Note struct {
	DeckName  string                 `json:"deckName"`
	ModelName string                 `json:"modelName"`
	Audio     string                 `json:"audio"`
	Fields    map[string]FieldValue  `json:"fields"`
	Tags      []string               `json:"tags"`
	Options   map[string]any         `json:"options"`
}

// NoteInfo represents detailed information about an Anki note
type NoteInfo struct {
	NoteId    int                     `json:"noteId"`
	Fields    map[string]FieldValue   `json:"fields"`
	ModelName string                  `json:"modelName"`
	Cards     []int                   `json:"cards"`
	Tags      []string                `json:"tags"`
	Mod       int                     `json:"mod"`
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

func (c *Client) GetNotesInfo(noteIds []int) ([]NoteInfo, error) {
	params := map[string]any{"notes": noteIds}
	infos, err := invoke[[]NoteInfo](c, "notesInfo", params)
	if err != nil {
		return nil, err
	}
	if len(infos) == 0 {
		return nil, fmt.Errorf("no notes found: %d", len(noteIds))
	}
	return infos, nil
}

// ExtractFields converts map[string]FieldValue to map[string]string
func ExtractFields(fields map[string]FieldValue) map[string]string {
	result := make(map[string]string)
	for name, field := range fields {
		result[name] = field.StringValue()
	}
	return result
}

// GetFieldValue safely extracts a field value from a map of FieldValue
func GetFieldValue(fields map[string]FieldValue, fieldName string) (string, error) {
	field, exists := fields[fieldName]
	if !exists {
		return "", fmt.Errorf("field '%s' not found", fieldName)
	}
	return field.StringValue(), nil
}

// GetAnkiNotesDetail retrieves detailed note information with extracted field values
func GetAnkiNotesDetail(client *Client, deckName string, keyField string) (map[string]NoteDetail, error) {
	notes, err := client.GetNotes(deckName)
	if err != nil {
		return nil, err
	}

	if len(notes) == 0 {
		return make(map[string]NoteDetail), nil
	}

	notesInfo, err := client.GetNotesInfo(notes)
	if err != nil {
		return nil, err
	}

	notesByWord := make(map[string]NoteDetail, len(notesInfo))
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
		
		notesByWord[keyValue] = NoteDetail{
			NoteID: note.NoteId,
			Fields: extractedFields,
		}
	}
	return notesByWord, nil
}
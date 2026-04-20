package anki

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

const (
	AnkiConnectURL = "http://localhost:8765"
)

type Action[T map[string]any] struct {
	Action  string `json:"action"`
	Params  T      `json:"params"`
	Version int    `json:"version"`
}

type Response[T any] struct {
	Error  any `json:"error"`
	Result T   `json:"result"`
}

type Note struct {
	DeckName  string            `json:"deckName"`
	ModelName string            `json:"modelName"`
	Audio     string            `json:"audio"`
	Fields    map[string]string `json:"fields"`
	Tags      []string          `json:"tags"`
	Options   map[string]any    `json:"options"`
}

type Client struct {
	url string
}

func invoke[T any](c *Client, action string, params map[string]any) (T, error) {
	var result T
	req := Action[map[string]any]{
		Action:  action,
		Params:  params,
		Version: 6,
	}

	jsonReq, err := json.Marshal(req)
	if err != nil {
		return result, err
	}
	log.Println("[anki.invoke] request:", string(jsonReq))

	resp, err := http.Post(c.url, "application/json", bytes.NewBuffer(jsonReq))
	if err != nil {
		return result, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return result, err
	}
	var response Response[T]
	log.Println("[anki.invoke] response:", string(body))
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
	return invoke[[]string](c, "modelFieldNames", map[string]any{"modelName": modelName})
}

func (c *Client) GetNotes(deckName string) ([]int, error) {
	return invoke[[]int](c, "findNotes", map[string]any{"query": fmt.Sprintf("deck:%s", deckName)})
}

func (c *Client) GetNoteInfo(noteId int) (map[string]any, error) {
	infos, err := invoke[[]map[string]any](c, "notesInfo", map[string]any{"notes": []int{noteId}})
	if err != nil {
		return nil, err
	}
	if len(infos) == 0 {
		return nil, fmt.Errorf("note not found: %d", noteId)
	}
	return infos[0], nil
}

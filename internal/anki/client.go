package anki

import (
	"encoding/json"
)

const (
	AnkiConnectURL = "http://localhost:8765"
)

type Action struct {
	Action  string `json:"action"`
	Params  any    `json:"params"`
	Version int    `json:"version"`
}

type Response struct {
	Error  string `json:"error"`
	Result any    `json:"result"`
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

func (c *Client) invoke(action string, params any) (*Response, error) {
	req := Action{
		Action:  action,
		Params:  params,
		Version: 6,
	}

	println(req)
	jsonReq, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	println(string(jsonReq))
	// TODO: implement
	// http.Post
	// json.Unmarshal response
	// handle anki error
	return nil, nil
}

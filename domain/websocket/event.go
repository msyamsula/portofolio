package websocket

import (
	"encoding/json"
	"fmt"
)

type Event struct {
	Type    string          `json:"type,omitempty"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

func createEvent(eventType string, payload json.RawMessage) []byte {
	b, err := json.Marshal(Event{
		Type:    eventType,
		Payload: payload,
	})
	if err != nil {
		fmt.Println(err)
	}
	return b
}

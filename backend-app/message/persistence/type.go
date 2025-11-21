package persistence

import (
	"encoding/json"
	"time"
)

type PostgresConfig struct {
	Username string
	Password string
	DbName   string
	Host     string
	Port     string
}

type SnsEvent struct {
	Event string `json:"event"`
}

type Message struct {
	Id             string    `json:"id,omitempty"`
	SenderId       int64     `json:"sender_id,omitempty"`
	ReceiverId     int64     `json:"receiver_id,omitempty"`
	ConversationId string    `json:"conversation_id,omitempty"`
	Data           string    `json:"data,omitempty"`
	CreateTime     time.Time `json:"create_time,omitempty"`

	SnsEvent
}

func (m *Message) UnmarshalJSON(b []byte) error {
	type Alias Message
	aux := &struct {
		*Alias
		SenderIDCamel       *int64  `json:"senderId"`
		ReceiverIDCamel     *int64  `json:"receiverId"`
		ConversationIDCamel *string `json:"conversationId"`
		CreateTimeCamel     *string `json:"createTime"`
	}{
		Alias: (*Alias)(m),
	}

	if err := json.Unmarshal(b, aux); err != nil {
		return err
	}

	// Override if snake_case exists
	if aux.SenderIDCamel != nil {
		m.SenderId = *aux.SenderIDCamel
	}
	if aux.ReceiverIDCamel != nil {
		m.ReceiverId = *aux.ReceiverIDCamel
	}
	if aux.ConversationIDCamel != nil {
		m.ConversationId = *aux.ConversationIDCamel
	}
	if aux.CreateTimeCamel != nil {
		t, err := time.Parse(time.RFC3339, *aux.CreateTimeCamel)
		if err != nil {
			return err
		}
		m.CreateTime = t
	}

	return nil
}

var (
	TableReadMessage   = "read_messages"
	TableUnreadMessage = "unread_messsages"
)

package persistence

import "time"

type PostgresConfig struct {
	Username string
	Password string
	DbName   string
	Host     string
	Port     string
}

type Message struct {
	Id             string    `json:"id,omitempty"`
	SenderId       int64     `json:"sender_id,omitempty"`
	ReceiverId     int64     `json:"receiver_id,omitempty"`
	ConversationId string    `json:"conversation_id,omitempty"`
	Data           string    `json:"data,omitempty"`
	CreateTime     time.Time `json:"create_time,omitempty"`
}

var (
	TableReadMessage   = "read_messages"
	TableUnreadMessage = "unread_messsages"
)

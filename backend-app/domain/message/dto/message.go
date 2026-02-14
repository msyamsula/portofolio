package dto

import "time"

// Message represents a message between users
type Message struct {
	ID             string    `json:"id,omitempty"`
	SenderID       int64     `json:"sender_id,omitempty"`
	ReceiverID     int64     `json:"receiver_id,omitempty"`
	ConversationID string    `json:"conversation_id,omitempty"`
	Data           string    `json:"data,omitempty"`
	CreateTime     time.Time `json:"create_time,omitempty"`
}

// InsertMessageRequest represents a request to insert a message
type InsertMessageRequest struct {
	SenderID       int64  `json:"sender_id"`
	ReceiverID     int64  `json:"receiver_id"`
	ConversationID string `json:"conversation_id"`
	Data           string `json:"data"`
}

// GetConversationRequest represents a request to get conversation messages
type GetConversationRequest struct {
	ConversationID string `json:"conversation_id,omitempty"`
}

// ConversationResponse represents the response from getting conversation messages
type ConversationResponse struct {
	Message string    `json:"message,omitempty"`
	Error   string    `json:"error,omitempty"`
	Data    []Message `json:"data,omitempty"`
}

// InsertMessageResponse represents the response from inserting a message
type InsertMessageResponse struct {
	Message string  `json:"message,omitempty"`
	Error   string  `json:"error,omitempty"`
	Data    Message `json:"data,omitempty"`
}

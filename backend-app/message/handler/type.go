package handler

import (
	"github.com/msyamsula/portofolio/backend-app/message/persistence"
	"github.com/msyamsula/portofolio/backend-app/message/service"
)

type HttpConfig struct {
	Svc service.Service
}

type conversationResponse struct {
	Error        string                `json:"error,omitempty"`
	Conversation []persistence.Message `json:"conversation,omitempty"`
}

type SqsConfig struct {
	QueueUrl string
	HttpConfig
}

type Config struct {
	SqsConfig
}

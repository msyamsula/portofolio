package handler

import (
	"context"
	"log"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

type Handler interface {
	GetConversation(http.ResponseWriter, *http.Request) // user open the app
	InsertMessage(http.ResponseWriter, *http.Request)   // listen from websocket event, and save message
	ReadMessage(http.ResponseWriter, *http.Request)     // listen from websocket event, and start read logic
}

func New(cfg Config) Handler {
	return &handler{
		svc: cfg.Svc,
	}
}

type SqsConsumer interface {
	Consume()
}

func NewSqsConsumer(c SqsConfig) SqsConsumer {
	ctx := context.Background()

	// Load AWS config (uses env vars, shared creds, IAM role, etc.)
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	return &sqsConsumer{
		client: sqs.NewFromConfig(cfg),
		url:    c.QueueUrl,
		svc:    c.Svc,
	}
}

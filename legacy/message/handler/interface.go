package handler

import (
	"net/http"
)

type Handler interface {
	HttpHandler // handle http request
	SqsConsumer // consume message from publisher
}

type HttpHandler interface {
	GetConversation(http.ResponseWriter, *http.Request)
}

type SqsConsumer interface {
	Consume()
}

func New(cfg Config) Handler {
	return &handler{
		httpHandler: newHttpHandler(cfg.HttpConfig),
		sqsConsumer: newSqsConsumer(cfg.SqsConfig),
	}
}

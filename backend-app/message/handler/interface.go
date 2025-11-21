package handler

import (
	"net/http"
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

package handler

import (
	"net/http"
)

type Handler interface {
	// http
	GoogleRedirectUrl(w http.ResponseWriter, req *http.Request)
	GetAppTokenForGoogle(w http.ResponseWriter, req *http.Request)
}

func New(cfg Config) Handler {
	return &httpHandler{
		randomizer: cfg.Randomizer,
		svc:        cfg.Svc,
		internal:   cfg.InternalToken,
	}
}

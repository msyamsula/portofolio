package handler

import (
	"net/http"
)

type Handler interface {
	// http
	GoogleRedirectUrl(w http.ResponseWriter, req *http.Request)
	GetAppTokenForGoogle(w http.ResponseWriter, req *http.Request)
}

type CombineHandler struct {
	*httpHandler
	*grpcHandler
}

func New(cfg Config) Handler {
	return &CombineHandler{
		httpHandler: &httpHandler{
			svc:        cfg.Svc,
			randomizer: cfg.Randomizer,
		},
		grpcHandler: &grpcHandler{},
	}
}

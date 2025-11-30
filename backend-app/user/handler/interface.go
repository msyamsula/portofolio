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

func New(cfg Config) *CombineHandler {
	return &CombineHandler{
		httpHandler: &httpHandler{
			randomizer: cfg.Randomizer,
			svc:        cfg.Svc,
		},
		grpcHandler: &grpcHandler{
			internalToken: cfg.InternalToken,
		},
	}
}

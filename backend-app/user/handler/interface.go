package handler

import (
	"net/http"
)

type Handler interface {
	GoogleRedirectUrl(w http.ResponseWriter, req *http.Request)
	GetAppTokenForGoogle(w http.ResponseWriter, req *http.Request)
}

func New(cfg Config) Handler {
	return &handler{
		svc: cfg.Svc,
	}
}

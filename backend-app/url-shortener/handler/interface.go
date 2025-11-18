package handler

import (
	"net/http"

	"github.com/msyamsula/portofolio/backend-app/url-shortener/services"
)

type Handler interface {
	Short(http.ResponseWriter, *http.Request)
	Redirect(http.ResponseWriter, *http.Request)
}

type Config struct {
	Svc services.Service
}

func New(cfg Config) Handler {
	return &handler{
		svc: cfg.Svc,
	}
}

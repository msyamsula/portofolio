package handler

import (
	"net/http"

	"github.com/msyamsula/portofolio/backend-app/user/service"
)

type Handler interface {
	InsertUser(http.ResponseWriter, *http.Request)
	GetUser(http.ResponseWriter, *http.Request)
}

type Config struct {
	Svc service.Service
}

func New(cfg Config) Handler {
	return &handler{
		svc: cfg.Svc,
	}
}

package handler

import (
	"net/http"

	"github.com/msyamsula/portofolio/backend-app/friend/service"
)

type Handler interface {
	AddFriend(http.ResponseWriter, *http.Request)
	GetFriends(http.ResponseWriter, *http.Request)
}

type Config struct {
	Svc service.Service
}

func New(cfg Config) Handler {
	return &handler{
		svc: cfg.Svc,
	}
}

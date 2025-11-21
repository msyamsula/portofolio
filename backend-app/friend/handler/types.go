package handler

import (
	repo "github.com/msyamsula/portofolio/backend-app/friend/persistent"
)

type Response struct {
	Message string    `json:"message,omitempty"`
	Error   string    `json:"error,omitempty"`
	Data    repo.User `json:"data,omitempty"`
}

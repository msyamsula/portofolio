package handler

import (
	"github.com/msyamsula/portofolio/backend-app/user/persistent"
)

type Response struct {
	Message string          `json:"message,omitempty"`
	Error   string          `json:"error,omitempty"`
	Data    persistent.User `json:"data,omitempty"`
}

package handler

import (
	"github.com/msyamsula/portofolio/backend-app/pkg/randomizer"
	"github.com/msyamsula/portofolio/backend-app/user/service"
	internaltoken "github.com/msyamsula/portofolio/backend-app/user/service/internal-token"
)

type Header struct {
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

type TokenResponse struct {
	Header
	Token string `json:"token"`
}

type Config struct {
	Svc           service.Service
	Randomizer    randomizer.Randomizer
	InternalToken internaltoken.InternalToken
}

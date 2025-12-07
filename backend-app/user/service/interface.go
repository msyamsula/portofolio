package service

import (
	"context"
)

type Service interface {
	GetRedirectUrlGoogle(c context.Context, browserCookies string) (string, error)
	GetAppTokenForGoogleUser(c context.Context, state, code string) (string, error)
}

func NewService(cfg ServiceConfig) Service {
	return &service{
		external:          cfg.External,
		internal:          cfg.Internal,
		sessionManagement: cfg.SessionManagement,
	}
}

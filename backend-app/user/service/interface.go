package service

import (
	"context"
)

type Service interface {
	GetRedirectUrlGoogle(c context.Context, browserCookies string) (string, error)
	GetAppTokenForGoogleUser(c context.Context, cookies, state, code string) (string, error)
}

func NewService(cfg ServiceConfig) Service {
	return &service{
		external: cfg.External,
		internal: cfg.Internal,
	}
}

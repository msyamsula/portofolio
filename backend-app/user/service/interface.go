package service

import (
	"context"

	"github.com/msyamsula/portofolio/backend-app/user/persistent"
)

type Service interface {
	SetUser(c context.Context, user persistent.User) (persistent.User, error)
	GetUser(c context.Context, username string) (persistent.User, error)
}

func New(cfg ServiceConfig) Service {
	return &service{
		persistence: cfg.Persistence,
	}
}

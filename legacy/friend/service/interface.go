package service

import (
	"context"

	repository "github.com/msyamsula/portofolio/backend-app/friend/persistent"
)

type Service interface {
	AddFriend(c context.Context, userA, userB repository.User) error
	GetFriends(c context.Context, user repository.User) ([]repository.User, error)
}

func New(config ServiceConfig) Service {
	return &service{
		persistent: config.Persistent,
	}
}

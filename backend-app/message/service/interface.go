package service

import (
	"context"

	"github.com/msyamsula/portofolio/backend-app/message/persistence"
)

type Service interface {
	InsertUnreadMessage(c context.Context, msg persistence.Message) (persistence.Message, error)
	GetConversation(c context.Context, conversationId string) ([]persistence.Message, error)
}

func New(config Config) Service {
	return &service{
		persistence: config.Persistence,
	}
}

package service

import (
	"context"

	repository "github.com/msyamsula/portofolio/backend-app/friend/persistent"
	"go.opentelemetry.io/otel"
)

type service struct {
	persistent repository.Repository
}

func (s *service) AddFriend(c context.Context, userA, userB repository.User) error {
	ctx, span := otel.Tracer("").Start(c, "service.AddFriend")
	defer span.End()

	var err error
	defer func() {
		if err != nil {
			span.RecordError(err)
		}
	}()

	err = s.persistent.AddFriend(ctx, userA, userB)
	if err != nil {
		return err
	}

	return nil
}

func (s *service) GetFriends(c context.Context, user repository.User) ([]repository.User, error) {
	ctx, span := otel.Tracer("").Start(c, "service.GetFriends")
	defer span.End()

	var err error
	defer func() {
		if err != nil {
			span.RecordError(err)
		}
	}()

	var users []repository.User
	users, err = s.persistent.GetFriends(ctx, user)
	if err != nil {
		return []repository.User{}, err
	}

	return users, nil
}

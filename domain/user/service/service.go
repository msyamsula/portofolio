package service

// mockgen -destination=test/mock_persistence.go -package=test . PersistenceLayer

import (
	"context"

	"github.com/msyamsula/portofolio/domain/user/repository"
	"go.opentelemetry.io/otel"
)

type Service struct {
	Persistence PersistenceLayer
	Cache       CacheLayer
}

type Dependencies struct {
	Persistence PersistenceLayer
	Cache       CacheLayer
}

func New(dep Dependencies) *Service {
	return &Service{
		Persistence: dep.Persistence,
		Cache:       dep.Cache,
	}
}

type PersistenceLayer interface {
	InsertUser(c context.Context, username string) (repository.User, error)
	GetUser(c context.Context, username string) (repository.User, error)
	AddFriend(c context.Context, userA, userB repository.User) error
	GetFriends(c context.Context, user repository.User) ([]repository.User, error)
}

type CacheLayer interface {
	SetUser(c context.Context, user repository.User) error
	GetUser(c context.Context, username string) (repository.User, error)
}

func (s *Service) SetUser(c context.Context, user repository.User) (repository.User, error) {
	ctx, span := otel.Tracer("").Start(c, "service.SetUser")
	defer span.End()

	var err error
	defer func() {
		if err != nil {
			span.RecordError(err)
		}
	}()

	user, err = s.Persistence.InsertUser(ctx, user.Username)
	if err != nil {
		return user, err
	}

	s.Cache.SetUser(ctx, user)
	return user, nil
}

func (s *Service) GetUser(c context.Context, username string) (repository.User, error) {
	ctx, span := otel.Tracer("").Start(c, "service.GetUser")
	defer span.End()

	var err error
	defer func() {
		if err != nil {
			span.RecordError(err)
		}
	}()

	var result repository.User
	result, err = s.Cache.GetUser(ctx, username)
	if err == nil && result.Id > 0 {
		// cache hit
		return result, nil
	}

	// cache miss
	result, err = s.Persistence.GetUser(ctx, username)
	if err != nil {
		return result, err
	}

	s.Cache.SetUser(ctx, result) // update cache
	return result, nil
}

func (s *Service) AddFriend(c context.Context, userA, userB repository.User) error {
	ctx, span := otel.Tracer("").Start(c, "service.AddFriend")
	defer span.End()

	var err error
	defer func() {
		if err != nil {
			span.RecordError(err)
		}
	}()

	err = s.Persistence.AddFriend(ctx, userA, userB)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) GetFriends(c context.Context, user repository.User) ([]repository.User, error) {
	ctx, span := otel.Tracer("").Start(c, "service.GetFriends")
	defer span.End()

	var err error
	defer func() {
		if err != nil {
			span.RecordError(err)
		}
	}()

	var users []repository.User
	users, err = s.Persistence.GetFriends(ctx, user)
	if err != nil {
		return []repository.User{}, err
	}

	return users, nil
}

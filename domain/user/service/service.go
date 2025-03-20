package service

// mockgen -destination=test/mock_persistence.go -package=test . PersistenceLayer

import (
	"context"

	"github.com/msyamsula/portofolio/domain/user/repository"
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
}

type CacheLayer interface {
	SetUser(c context.Context, user repository.User) error
	GetUser(c context.Context, username string) (repository.User, error)
}

func (s *Service) SetUser(c context.Context, user repository.User) (repository.User, error) {
	user, err := s.Persistence.InsertUser(c, user.Username)
	if err != nil {
		return user, err
	}

	s.Cache.SetUser(c, user)
	return user, nil
}

func (s *Service) GetUser(c context.Context, username string) (repository.User, error) {
	result, err := s.Cache.GetUser(c, username)
	if err == nil && result.Id > 0 {
		// cache hit
		return result, nil
	}

	// cache miss
	result, err = s.Persistence.GetUser(c, username)
	if err != nil {
		return result, err
	}

	s.Cache.SetUser(c, result) // update cache
	return result, nil
}

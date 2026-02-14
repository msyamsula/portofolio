package service

import (
	"context"

	"github.com/msyamsula/portofolio/backend-app/domain/friend/dto"
	"github.com/msyamsula/portofolio/backend-app/domain/friend/repository"
)

// Service defines the interface for friend business logic
type Service interface {
	// AddFriend adds a friendship relationship between two users
	AddFriend(ctx context.Context, userA, userB dto.User) error

	// GetFriends retrieves all friends for a given user
	GetFriends(ctx context.Context, user dto.User) ([]dto.User, error)
}

// friendService implements the Service interface
type friendService struct {
	repo repository.Repository
}

// New creates a new friend service
func New(repo repository.Repository) Service {
	return &friendService{
		repo: repo,
	}
}

// AddFriend adds a friendship relationship between two users
func (s *friendService) AddFriend(ctx context.Context, userA, userB dto.User) error {
	// Validation: userA and userB must be different
	if userA.ID == userB.ID {
		return repository.ErrIDMustBeDifferent
	}

	// Delegate to repository
	return s.repo.AddFriend(ctx, userA, userB)
}

// GetFriends retrieves all friends for a given user
func (s *friendService) GetFriends(ctx context.Context, user dto.User) ([]dto.User, error) {
	// Delegate to repository
	return s.repo.GetFriends(ctx, user)
}

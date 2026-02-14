package repository

import (
	"context"
	"errors"

	"github.com/msyamsula/portofolio/backend-app/domain/friend/dto"
	infraDB "github.com/msyamsula/portofolio/backend-app/infrastructure/database/postgres"
)

var (
	// ErrUserNotFound is returned when a user cannot be found
	ErrUserNotFound = errors.New("user not found")
	// ErrIDMustBeDifferent is returned when trying to add a user as friend of themselves
	ErrIDMustBeDifferent = errors.New("id must be different")
)

// Repository defines the interface for friend data access
type Repository interface {
	// AddFriend adds a friendship relationship between two users
	AddFriend(ctx context.Context, userA, userB dto.User) error

	// GetFriends retrieves all friends for a given user
	GetFriends(ctx context.Context, user dto.User) ([]dto.User, error)
}

// postgresRepository implements the Repository interface using PostgreSQL
type postgresRepository struct {
	db infraDB.Database
}

// NewPostgresRepository creates a new PostgreSQL-based repository
func NewPostgresRepository(db infraDB.Database) Repository {
	return &postgresRepository{
		db: db,
	}
}

// AddFriend adds a friendship relationship between two users
func (r *postgresRepository) AddFriend(ctx context.Context, userA, userB dto.User) error {
	// Validation: userA and userB must be different
	if userA.ID == userB.ID {
		return ErrIDMustBeDifferent
	}

	// Determine small_id and big_id (ensure consistent ordering)
	smallID := min(userA.ID, userB.ID)
	bigID := max(userA.ID, userB.ID)

	// Insert friendship relationship
	query := `
		INSERT INTO friendship (small_id, big_id)
		VALUES ($1, $2)
		ON CONFLICT (small_id, big_id) DO NOTHING
		RETURNING id
	`

	_, err := r.db.ExecContext(ctx, query, smallID, bigID)
	if err != nil {
		return err
	}

	return nil
}

// GetFriends retrieves all friends for a given user
func (r *postgresRepository) GetFriends(ctx context.Context, user dto.User) ([]dto.User, error) {
	query := `
		SELECT l.id, l.username, l.online, COALESCE(ur.unread, 0)
		FROM (
			-- Users where this user is the smaller ID
			SELECT u.id, u.username, u.online FROM
				(SELECT big_id FROM friendship WHERE small_id = $1) f
			JOIN users u ON u.id = f.big_id
			UNION ALL
			-- Users where this user is the bigger ID
			SELECT u.id, u.username, u.online FROM
				(SELECT small_id FROM friendship WHERE big_id = $1) f
			JOIN users u ON u.id = f.small_id
		) l
		LEFT JOIN unread ur ON l.id = ur.sender_id AND ur.receiver_id = $1
	`

	var users []dto.User
	err := r.db.SelectContext(ctx, &users, query, user.ID)
	if err != nil {
		return nil, err
	}

	return users, nil
}

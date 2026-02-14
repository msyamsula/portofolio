package repository

import (
	"context"
	"errors"

	"github.com/msyamsula/portofolio/backend-app/domain/message/dto"
	infraDB "github.com/msyamsula/portofolio/backend-app/infrastructure/database/postgres"
)

var (
	// ErrBadRequest is returned when the request has invalid parameters
	ErrBadRequest = errors.New("bad request in message body")
)

const (
	// TableMessages is the default table name for messages
	TableMessages = "messages"
)

// Repository defines the interface for message data access
type Repository interface {
	// InsertMessage inserts a new message into the database
	InsertMessage(ctx context.Context, msg dto.Message, table string) (dto.Message, error)

	// GetConversation retrieves all messages for a conversation
	GetConversation(ctx context.Context, conversationID string, table string) ([]dto.Message, error)
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

// InsertMessage inserts a new message into the database
func (r *postgresRepository) InsertMessage(ctx context.Context, msg dto.Message, table string) (dto.Message, error) {
	query := `
		INSERT INTO ` + table + ` (id, sender_id, receiver_id, conversation_id, data)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, sender_id, receiver_id, conversation_id, data, create_time
	`

	var result dto.Message
	err := r.db.GetContext(ctx, &result, query, msg.ID, msg.SenderID, msg.ReceiverID, msg.ConversationID, msg.Data)
	if err != nil {
		return dto.Message{}, err
	}

	return result, nil
}

// GetConversation retrieves all messages for a conversation
func (r *postgresRepository) GetConversation(ctx context.Context, conversationID string, table string) ([]dto.Message, error) {
	query := `
		SELECT id, sender_id, receiver_id, conversation_id, data, create_time
		FROM ` + table + `
		WHERE conversation_id = $1
		ORDER BY create_time
	`

	var messages []dto.Message
	err := r.db.SelectContext(ctx, &messages, query, conversationID)
	if err != nil {
		return nil, err
	}

	return messages, nil
}

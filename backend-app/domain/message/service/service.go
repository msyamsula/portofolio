package service

import (
	"context"
	"errors"

	"github.com/msyamsula/portofolio/backend-app/domain/message/dto"
	"github.com/msyamsula/portofolio/backend-app/domain/message/repository"
)

// Service defines the interface for message business logic
type Service interface {
	// InsertMessage inserts a new message
	InsertMessage(ctx context.Context, msg dto.Message) (dto.Message, error)

	// GetConversation retrieves all messages for a conversation
	GetConversation(ctx context.Context, conversationID string) ([]dto.Message, error)
}

// messageService implements the Service interface
type messageService struct {
	repo repository.Repository
}

// New creates a new message service
func New(repo repository.Repository) Service {
	return &messageService{
		repo: repo,
	}
}

// InsertMessage inserts a new message
func (s *messageService) InsertMessage(ctx context.Context, msg dto.Message) (dto.Message, error) {
	// Validation: sender_id, receiver_id, and data must be provided
	if msg.SenderID <= 0 ||
		msg.ReceiverID <= 0 ||
		msg.SenderID == msg.ReceiverID ||
		msg.Data == "" {
		return dto.Message{}, repository.ErrBadRequest
	}

	// Delegate to repository
	return s.repo.InsertMessage(ctx, msg, repository.TableMessages)
}

// GetConversation retrieves all messages for a conversation
func (s *messageService) GetConversation(ctx context.Context, conversationID string) ([]dto.Message, error) {
	// Validation: conversation_id must be provided
	if conversationID == "" {
		return nil, errors.New("conversation_id is required")
	}

	// Delegate to repository
	return s.repo.GetConversation(ctx, conversationID, repository.TableMessages)
}

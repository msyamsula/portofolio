package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"message/domain"
	"message/repository"
)

// ErrAlreadySent is returned when the event_id INSERT conflicts, meaning this
// message was already submitted by a previous call. Callers (HTTP handler)
// should treat this as a successful idempotent response (HTTP 200), not an error.
var ErrAlreadySent = errors.New("message already sent")

// SendRequest contains the fields supplied by the API caller for a new message.
type SendRequest struct {
	MessageID      uuid.UUID
	ConversationID uuid.UUID
	SenderID       uuid.UUID
	ReceiverID     uuid.UUID
	Payload        []byte
}

// Sender handles the a→b state machine transition.
// It atomically writes a single SENT event with version=0 to the event log.
type Sender struct {
	eventLog *repository.EventLogRepo
}

// NewSender constructs a Sender.
func NewSender(eventLog *repository.EventLogRepo) *Sender {
	return &Sender{eventLog: eventLog}
}

// SendMessage commits a SENT event for the given request.
//
// Idempotency: the event_id is hash(message_id, conversation_id, 0), so duplicate
// API calls with the same message_id silently resolve via ON CONFLICT DO NOTHING.
// The caller receives ErrAlreadySent, which the HTTP handler maps to HTTP 200.
//
// On commit success, the outbox relay picks up the unpublished event and forwards
// it to Kafka, which triggers the consumer to call ProcessSuccess (b→d).
func (s *Sender) SendMessage(ctx context.Context, req SendRequest) error {
	const version = 0

	event := domain.EventLog{
		EventID:        domain.HashEventID(req.MessageID, req.ConversationID, version),
		MessageID:      req.MessageID,
		ConversationID: req.ConversationID,
		SenderID:       req.SenderID,
		ReceiverID:     req.ReceiverID,
		Version:        version,
		EventName:      domain.EventSent,
		Payload:        req.Payload,
	}

	tx, err := s.eventLog.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("sender begin tx: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	inserted, err := s.eventLog.Append(ctx, tx, event)
	if err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("sender commit: %w", err)
	}

	if !inserted {
		// A previous call (or a concurrent duplicate) already wrote this SENT event.
		// Signal idempotency to the caller — not a true error.
		return ErrAlreadySent
	}
	return nil
}

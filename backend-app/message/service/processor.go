package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"message/domain"
	"message/repository"
)

// Processor handles the b→d state machine transition.
// It is called by the Kafka consumer for each SENT event and by the reconciler
// when it needs to re-drive a stuck message directly.
type Processor struct {
	eventLog     *repository.EventLogRepo
	conversation *repository.ConversationRepo
}

// NewProcessor constructs a Processor.
func NewProcessor(eventLog *repository.EventLogRepo, conversation *repository.ConversationRepo) *Processor {
	return &Processor{eventLog: eventLog, conversation: conversation}
}

// ProcessSuccess attempts to advance a message from SENT (b) to SUCCESS (d).
//
// Idempotency contract:
//   - If the conversation row already exists → return nil (ack Kafka, already done).
//   - If the event_log INSERT conflicts (peer won the version race) → check conversation
//     table and return nil either way. The peer either committed the full transaction
//     (conversation row exists, we're done) or it rolled back (reconciler will catch it).
//
// The SUCCESS event and conversation row are written in a single transaction,
// so they are always consistent: either both exist or neither does.
func (p *Processor) ProcessSuccess(ctx context.Context, e domain.EventLog) error {
	events, err := p.eventLog.GetByMessage(ctx, e.MessageID, e.ConversationID)
	if err != nil {
		return err
	}
	inConv, err := p.conversation.Exists(ctx, e.MessageID, e.ConversationID)
	if err != nil {
		return err
	}

	state := domain.BuildState(events, inConv)

	// Terminal states need no further action — ack the message.
	if state.Current == domain.StateSuccess || state.Current == domain.StateFailed {
		return nil
	}

	return p.commitSuccess(ctx, e.MessageID, e.ConversationID, state)
}

// commitSuccess writes the SUCCESS event and conversation row atomically.
//
// Race scenario: two consumers both read version=N and race to write version=N+1.
// Exactly one wins via the UNIQUE(message_id, conversation_id, version) constraint.
// The loser gets inserted=false, checks if the conversation row now exists,
// and returns nil (idempotent ack) regardless. The reconciler handles edge cases
// where the winner's transaction committed the event but not the conversation row.
func (p *Processor) commitSuccess(
	ctx context.Context,
	mid, cid uuid.UUID,
	state domain.MessageState,
) error {
	newVersion := state.Version + 1
	newEventID := domain.HashEventID(mid, cid, newVersion)

	successEvent := domain.EventLog{
		EventID:        newEventID,
		MessageID:      mid,
		ConversationID: cid,
		SenderID:       state.SenderID,
		ReceiverID:     state.ReceiverID,
		Version:        newVersion,
		EventName:      domain.EventSuccess,
		Payload:        state.Payload,
	}

	tx, err := p.eventLog.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("processor begin tx: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	inserted, err := p.eventLog.Append(ctx, tx, successEvent)
	if err != nil {
		return err
	}

	if !inserted {
		// A peer process already appended this version. Roll back our empty transaction
		// and determine whether the peer also committed the conversation row.
		// Either way we return nil — this message is handled (or the reconciler will retry).
		tx.Rollback(ctx) //nolint:errcheck
		_, err := p.conversation.Exists(ctx, mid, cid)
		if err != nil {
			return err
		}
		// Idempotent ack: peer either fully committed or will be caught by reconciler.
		return nil
	}

	conv := repository.Conversation{
		ID:             uuid.New(),
		ConversationID: cid,
		SenderID:       state.SenderID,
		ReceiverID:     state.ReceiverID,
		MessageID:      mid,
		EventID:        newEventID,
		Payload:        state.Payload,
	}
	if _, err := p.conversation.Insert(ctx, tx, conv); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("processor commit: %w", err)
	}
	return nil
}

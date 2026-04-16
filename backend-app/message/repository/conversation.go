package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Conversation is the data written to the conversation table when a message
// reaches SUCCESS state. A row here is the authoritative signal for delivery.
type Conversation struct {
	ID             uuid.UUID
	ConversationID uuid.UUID
	SenderID       uuid.UUID
	ReceiverID     uuid.UUID
	MessageID      uuid.UUID

	// EventID is the SUCCESS event_id that triggered this row.
	// Its UNIQUE constraint in the DB prevents duplicate conversation rows even if
	// two processes race — only the first INSERT wins; the second is silently ignored.
	EventID uuid.UUID

	Payload []byte
}

// ConversationRepo is the data-access layer for the conversation table.
type ConversationRepo struct {
	pool *pgxpool.Pool
}

// NewConversationRepo creates a repo backed by the given connection pool.
func NewConversationRepo(pool *pgxpool.Pool) *ConversationRepo {
	return &ConversationRepo{pool: pool}
}

// Insert writes a conversation row inside an existing transaction.
//
// ON CONFLICT DO NOTHING on (message_id, conversation_id) makes this safe to
// call multiple times for the same message — the second call is a silent no-op.
// Returns inserted=false when a row already existed (idempotent success).
func (r *ConversationRepo) Insert(ctx context.Context, tx pgx.Tx, c Conversation) (bool, error) {
	tag, err := tx.Exec(ctx, `
		INSERT INTO conversation
		    (id, conversation_id, sender_id, receiver_id, message_id, event_id, payload)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT DO NOTHING`,
		c.ID, c.ConversationID, c.SenderID, c.ReceiverID,
		c.MessageID, c.EventID, c.Payload,
	)
	if err != nil {
		return false, fmt.Errorf("conversation insert: %w", err)
	}
	return tag.RowsAffected() == 1, nil
}

// Exists returns true if a conversation row for (message_id, conversation_id) exists.
//
// This is the primary check for StateSuccess. It is called:
//   - Before every write in processor.ProcessSuccess to short-circuit re-processing.
//   - By the Kafka consumer after a version-conflict to decide whether to ack.
//   - By the reconciler to exclude already-delivered messages from GetStuck results.
func (r *ConversationRepo) Exists(ctx context.Context, mid, cid uuid.UUID) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx,
		`SELECT EXISTS(
		     SELECT 1 FROM conversation
		     WHERE  message_id = $1 AND conversation_id = $2
		 )`,
		mid, cid,
	).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("conversation exists: %w", err)
	}
	return exists, nil
}

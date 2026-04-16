package domain

import (
	"crypto/sha256"
	"strconv"
	"time"

	"github.com/google/uuid"
)

// EventName is the labeled edge in the state machine graph.
//
// State machine:
//
//	a ──SENT──► b ──SUCCESS──► d (terminal success)
//	            │
//	            ├──RECONCILE_STUCKED_MESSAGE──► b  (stuck retry loop)
//	            ├──RECONCILE_FAILED_MESSAGE──► b   (failed retry loop)
//	            └──FAILED──► c                     (terminal failure)
type EventName string

const (
	// EventSent is written by the API on the first message submission (a→b).
	// Always at version=0. The only event the API layer ever writes.
	EventSent EventName = "SENT"

	// EventSuccess is written by the worker when processing succeeds (b→d).
	// Always written atomically with a conversation table row.
	EventSuccess EventName = "SUCCESS"

	// EventFailed is terminal (→c). Written by the reconciler when retry
	// and attempt limits are exhausted. No processing occurs after this.
	EventFailed EventName = "FAILED"

	// EventReconcileStuck is written by the reconciler for messages stuck in
	// the SENT state. Counts toward MaxRetry.
	EventReconcileStuck EventName = "RECONCILE_STUCKED_MESSAGE"

	// EventReconcileFailed is written by the reconciler once stuck-retries are
	// exhausted and the message still hasn't succeeded. Counts toward MaxAttempt.
	EventReconcileFailed EventName = "RECONCILE_FAILED_MESSAGE"
)

// EventLog is a single immutable entry in the append-only event log.
// Every state change produces exactly one row. Rows are never updated or deleted.
type EventLog struct {
	EventID        uuid.UUID `json:"event_id"`
	MessageID      uuid.UUID `json:"message_id"`
	ConversationID uuid.UUID `json:"conversation_id"`
	SenderID       uuid.UUID `json:"sender_id"`
	ReceiverID     uuid.UUID `json:"receiver_id"`

	// Version is the optimistic concurrency counter for (MessageID, ConversationID).
	// Starts at 0 (SENT) and increments by exactly 1 on every append.
	// UNIQUE(message_id, conversation_id, version) in the DB ensures exactly one
	// writer wins when two processes race to write the same next version.
	Version int `json:"version"`

	EventName EventName `json:"event_name"`

	// Payload carries the original message body, copied from SENT into all
	// subsequent events so every event is self-describing.
	Payload []byte `json:"payload,omitempty"`

	Published bool      `json:"published"`
	CreatedAt time.Time `json:"created_at"`
}

// HashEventID produces a deterministic UUID from (messageID, conversationID, version).
//
// Determinism is the key property: two concurrent processes computing the same
// next event produce the same event_id. The DB PRIMARY KEY resolves the race —
// exactly one INSERT wins; the other sees a conflict and treats it as a no-op.
func HashEventID(messageID, conversationID uuid.UUID, version int) uuid.UUID {
	input := messageID.String() + ":" + conversationID.String() + ":" + strconv.Itoa(version)
	h := sha256.Sum256([]byte(input))
	var id uuid.UUID
	copy(id[:], h[:16])
	// RFC 4122 version-5 variant: signals deterministic / name-based origin.
	id[6] = (id[6] & 0x0f) | 0x50
	id[8] = (id[8] & 0x3f) | 0x80
	return id
}

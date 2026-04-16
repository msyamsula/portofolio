package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"message/domain"
)

// StuckRef identifies a (message_id, conversation_id) pair that is a candidate
// for reconciliation: not yet in the conversation table and not yet terminal.
type StuckRef struct {
	MessageID      uuid.UUID
	ConversationID uuid.UUID
}

// EventLogRepo is the data-access layer for the event_log table.
// It owns no business logic — callers decide what the conflict semantics mean.
type EventLogRepo struct {
	pool *pgxpool.Pool
}

// NewEventLogRepo creates a repo backed by the given connection pool.
func NewEventLogRepo(pool *pgxpool.Pool) *EventLogRepo {
	return &EventLogRepo{pool: pool}
}

// BeginTx starts a new transaction on the pool.
// The caller is responsible for calling Commit or Rollback.
func (r *EventLogRepo) BeginTx(ctx context.Context) (pgx.Tx, error) {
	return r.pool.Begin(ctx)
}

// Append inserts an EventLog row inside an existing transaction.
//
// ON CONFLICT DO NOTHING means a duplicate (event_id or version) is silently
// ignored. The caller inspects the returned `inserted` flag to decide semantics:
//   - inserted=true:  the event was new; normal flow continues.
//   - inserted=false: a peer process already wrote this event; treat as idempotent.
func (r *EventLogRepo) Append(ctx context.Context, tx pgx.Tx, e domain.EventLog) (bool, error) {
	tag, err := tx.Exec(ctx, `
		INSERT INTO event_log
		    (event_id, message_id, conversation_id, sender_id, receiver_id,
		     version, event_name, payload, published)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, false)
		ON CONFLICT DO NOTHING`,
		e.EventID, e.MessageID, e.ConversationID,
		e.SenderID, e.ReceiverID,
		e.Version, string(e.EventName), e.Payload,
	)
	if err != nil {
		return false, fmt.Errorf("event_log append: %w", err)
	}
	return tag.RowsAffected() == 1, nil
}

// GetByMessage returns all events for a (message_id, conversation_id) pair,
// ordered by version ASC. Callers pass this slice to domain.BuildState to
// reconstruct the current message state without reading any other table.
func (r *EventLogRepo) GetByMessage(ctx context.Context, mid, cid uuid.UUID) ([]domain.EventLog, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT event_id, message_id, conversation_id, sender_id, receiver_id,
		       version, event_name, payload, published, created_at
		FROM   event_log
		WHERE  message_id = $1 AND conversation_id = $2
		ORDER  BY version ASC`,
		mid, cid,
	)
	if err != nil {
		return nil, fmt.Errorf("event_log get_by_message: %w", err)
	}
	defer rows.Close()
	return scanEvents(rows)
}

// GetUnpublished returns up to `limit` events that have not yet been forwarded
// to Kafka, ordered by creation time. The outbox relay calls this on each tick.
func (r *EventLogRepo) GetUnpublished(ctx context.Context, limit int) ([]domain.EventLog, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT event_id, message_id, conversation_id, sender_id, receiver_id,
		       version, event_name, payload, published, created_at
		FROM   event_log
		WHERE  published = false
		ORDER  BY created_at ASC
		LIMIT  $1`,
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf("event_log get_unpublished: %w", err)
	}
	defer rows.Close()
	return scanEvents(rows)
}

// MarkPublished sets published=true for all given event IDs in a single batch
// UPDATE, reducing per-event round-trips in the outbox relay hot loop.
func (r *EventLogRepo) MarkPublished(ctx context.Context, ids []uuid.UUID) error {
	if len(ids) == 0 {
		return nil
	}
	// pgx/v5 encodes []uuid.UUID ([16]byte slice) as a PostgreSQL UUID array natively.
	_, err := r.pool.Exec(ctx,
		`UPDATE event_log SET published = true WHERE event_id = ANY($1::uuid[])`,
		uuidSliceToStrings(ids),
	)
	if err != nil {
		return fmt.Errorf("event_log mark_published: %w", err)
	}
	return nil
}

// GetStuck returns all (message_id, conversation_id) pairs that need reconciliation:
//   - NOT present in the conversation table (message not yet delivered), AND
//   - No FAILED event in the log (message not yet terminal).
//
// The reconciler sweeps this set every ReconcileInterval.
func (r *EventLogRepo) GetStuck(ctx context.Context) ([]StuckRef, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT DISTINCT el.message_id, el.conversation_id
		FROM   event_log el
		WHERE  NOT EXISTS (
		    SELECT 1 FROM conversation c
		    WHERE  c.message_id      = el.message_id
		    AND    c.conversation_id = el.conversation_id
		)
		AND NOT EXISTS (
		    SELECT 1 FROM event_log el2
		    WHERE  el2.message_id      = el.message_id
		    AND    el2.conversation_id = el.conversation_id
		    AND    el2.event_name      = 'FAILED'
		)`,
	)
	if err != nil {
		return nil, fmt.Errorf("event_log get_stuck: %w", err)
	}
	defer rows.Close()

	var refs []StuckRef
	for rows.Next() {
		var ref StuckRef
		if err := rows.Scan(&ref.MessageID, &ref.ConversationID); err != nil {
			return nil, fmt.Errorf("event_log get_stuck scan: %w", err)
		}
		refs = append(refs, ref)
	}
	return refs, rows.Err()
}

// scanEvents maps raw pgx rows to a []domain.EventLog slice.
// Kept as a package-private helper to avoid duplicating scan logic.
func scanEvents(rows pgx.Rows) ([]domain.EventLog, error) {
	var events []domain.EventLog
	for rows.Next() {
		var e domain.EventLog
		var eventName string
		var createdAt time.Time

		if err := rows.Scan(
			&e.EventID, &e.MessageID, &e.ConversationID,
			&e.SenderID, &e.ReceiverID,
			&e.Version, &eventName, &e.Payload,
			&e.Published, &createdAt,
		); err != nil {
			return nil, fmt.Errorf("event_log scan: %w", err)
		}
		e.EventName = domain.EventName(eventName)
		e.CreatedAt = createdAt
		events = append(events, e)
	}
	return events, rows.Err()
}

// uuidSliceToStrings converts []uuid.UUID to []string for use with pgx ANY($1::uuid[]).
// pgx/v5 handles []string → text[] → uuid[] cast cleanly via the ::uuid[] explicit cast.
func uuidSliceToStrings(ids []uuid.UUID) []string {
	out := make([]string, len(ids))
	for i, id := range ids {
		out[i] = id.String()
	}
	return out
}

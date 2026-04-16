package outbox

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	msgkafka "message/infrastructure/kafka"
	"message/repository"

	"github.com/google/uuid"
)

// Relay polls the event_log for unpublished events and forwards them to Kafka.
//
// Write path:
//   event_log INSERT (published=false) → Relay polls → Kafka publish → mark published
//
// At-least-once guarantee: if the process crashes between the Kafka write and
// MarkPublished, the event will be re-published on the next poll cycle.
// Consumers handle duplicates via the idempotent state machine.
type Relay struct {
	eventLog *repository.EventLogRepo
	producer *msgkafka.Producer
	batch    int
}

// NewRelay creates a Relay that publishes unpublished events to Kafka.
// batch controls how many events are fetched and published per poll cycle.
func NewRelay(eventLog *repository.EventLogRepo, producer *msgkafka.Producer, batch int) *Relay {
	return &Relay{eventLog: eventLog, producer: producer, batch: batch}
}

// Run starts the polling loop. It ticks every 2 seconds and publishes
// any unpublished events in batches. Blocks until ctx is cancelled.
func (r *Relay) Run(ctx context.Context) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := r.relayOnce(ctx); err != nil {
				slog.Error("outbox relay error", "err", err)
			}
		}
	}
}

// relayOnce fetches one batch of unpublished events, publishes them to Kafka,
// then marks them published. On Kafka failure the batch is not marked, so it
// will be retried on the next tick — preserving at-least-once delivery.
func (r *Relay) relayOnce(ctx context.Context) error {
	events, err := r.eventLog.GetUnpublished(ctx, r.batch)
	if err != nil || len(events) == 0 {
		return err
	}

	published := make([]uuid.UUID, 0, len(events))
	for _, e := range events {
		body, err := json.Marshal(e)
		if err != nil {
			slog.Error("outbox marshal failed", "event_id", e.EventID, "err", err)
			continue
		}
		// Partition key = conversationID: all events for the same conversation
		// land on the same Kafka partition, preserving per-conversation ordering.
		if err := r.producer.Publish(ctx, e.ConversationID.String(), body); err != nil {
			// Stop the batch on the first publish failure.
			// Already-published events in this batch will be re-delivered on the
			// next tick, but consumers handle duplicates via idempotent state.
			slog.Error("outbox publish failed", "event_id", e.EventID, "err", err)
			break
		}
		published = append(published, e.EventID)
	}

	if len(published) == 0 {
		return nil
	}
	return r.eventLog.MarkPublished(ctx, published)
}

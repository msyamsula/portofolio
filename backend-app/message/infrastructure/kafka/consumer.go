package kafka

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
)

// Consumer wraps kafka-go's Reader for consume-and-commit use.
// One Consumer instance maps to one consumer group + topic pair.
type Consumer struct {
	reader *kafka.Reader
}

// NewConsumer creates a Consumer for the given brokers, topic, and group.
// The group ID ensures that multiple worker replicas share the partition load
// without double-processing: Kafka delivers each message to exactly one
// member of the consumer group.
func NewConsumer(brokers []string, topic, groupID string) *Consumer {
	return &Consumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers:        brokers,
			Topic:          topic,
			GroupID:        groupID,
			MinBytes:       1,
			MaxBytes:       10e6,
			CommitInterval: time.Second, // async commit every second
			StartOffset:    kafka.FirstOffset,
		}),
	}
}

// Consume reads messages in a blocking loop until ctx is cancelled.
//
// Each message is passed to `handler`. If the handler returns an error, the
// error is logged and the message is committed anyway — this is the
// poison-pill protection strategy: a bad message is acknowledged and does
// not block the consumer from making progress on subsequent messages.
//
// If persistent failures are a concern, route errored messages to a DLQ
// topic inside the handler before returning nil.
func (c *Consumer) Consume(ctx context.Context, handler func([]byte) error) error {
	for {
		msg, err := c.reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return nil // clean shutdown
			}
			return fmt.Errorf("fetch message: %w", err)
		}

		if err := handler(msg.Value); err != nil {
			// Log and move on — duplicate work is acceptable; stalling is not.
			log.Printf("[consumer] handler error key=%s: %v", msg.Key, err)
		}

		// Commit unconditionally: we prefer at-least-once delivery over
		// blocking on a bad message. Handlers must be idempotent.
		if err := c.reader.CommitMessages(ctx, msg); err != nil {
			log.Printf("[consumer] commit error: %v", err)
		}
	}
}

// Close releases the underlying Kafka connection.
func (c *Consumer) Close() error {
	return c.reader.Close()
}

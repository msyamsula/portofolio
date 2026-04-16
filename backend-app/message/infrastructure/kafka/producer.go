package kafka

import (
	"context"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
)

// Producer wraps kafka-go's Writer for publish-only use.
// It is safe for concurrent use.
type Producer struct {
	writer *kafka.Writer
}

// NewProducer creates a Producer that sends to the given brokers.
// RequiredAcks=1 (leader acknowledgment) balances durability and throughput.
func NewProducer(brokers []string, topic string) *Producer {
	return &Producer{
		writer: &kafka.Writer{
			Addr:         kafka.TCP(brokers...),
			Topic:        topic,
			Balancer:     &kafka.Hash{},       // route by key for ordered per-message delivery
			RequiredAcks: kafka.RequireOne,
			WriteTimeout: 10 * time.Second,
		},
	}
}

// Publish sends a single message to Kafka.
// The key is used for partition routing — pass the event_id so all events
// for the same message land on the same partition (ordering guarantee).
func (p *Producer) Publish(ctx context.Context, key string, value []byte) error {
	err := p.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(key),
		Value: value,
	})
	if err != nil {
		return fmt.Errorf("kafka publish: %w", err)
	}
	return nil
}

// Close flushes pending messages and closes the underlying writer.
// Call this on process shutdown.
func (p *Producer) Close() error {
	return p.writer.Close()
}

package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/jackc/pgx/v5/pgxpool"

	"message/config"
	"message/domain"
	msgkafka "message/infrastructure/kafka"
	"message/repository"
	"message/service"
)

// main wires the Worker process: Kafka consumer only.
// Reconciliation runs in its own dedicated container (cmd/reconciler).
// The API process (cmd/api) handles HTTP and the outbox relay.
func main() {
	cfg := config.Load()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool := mustConnectDB(ctx, cfg.DBDSN)
	defer pool.Close()

	eventLogRepo := repository.NewEventLogRepo(pool)
	convRepo := repository.NewConversationRepo(pool)
	processor := service.NewProcessor(eventLogRepo, convRepo)

	// Kafka consumer: reads events published by the outbox relay and drives
	// the b→d state machine transition by calling ProcessSuccess.
	// Uses a consumer group so multiple worker replicas share partition load
	// without double-processing (Kafka delivers each partition to one member).
	consumer := msgkafka.NewConsumer(
		strings.Split(cfg.KafkaBrokers, ","),
		cfg.KafkaTopic,
		"cg-message-worker",
	)
	defer consumer.Close()

	slog.Info("worker consuming", "topic", cfg.KafkaTopic)
	if err := consumer.Consume(ctx, makeHandler(processor)); err != nil {
		slog.Error("consumer stopped", "err", err)
	}
	slog.Info("worker shut down")
}

// makeHandler returns a Kafka message handler that routes events to the processor.
//
// Only SENT and RECONCILE_* events trigger processing — they advance the message
// toward SUCCESS or FAILED. Terminal events (SUCCESS, FAILED) are acked immediately.
//
// The consumer always commits regardless of handler error (poison-pill safe).
// The reconciler container catches messages whose processing failed before ack.
func makeHandler(processor *service.Processor) func([]byte) error {
	return func(data []byte) error {
		var event domain.EventLog
		if err := json.Unmarshal(data, &event); err != nil {
			slog.Error("worker unmarshal failed", "err", err)
			return nil // commit to unblock the consumer
		}

		switch event.EventName {
		case domain.EventSent, domain.EventReconcileStuck, domain.EventReconcileFailed:
			if err := processor.ProcessSuccess(context.Background(), event); err != nil {
				// Log and return nil: consumer still commits.
				// Reconciler catches this message within ReconcileInterval.
				slog.Error("worker process failed",
					"message_id", event.MessageID,
					"event_name", event.EventName,
					"err", err,
				)
			}

		case domain.EventSuccess, domain.EventFailed:
			// Terminal state already written to DB — ack without side effects.

		default:
			slog.Warn("worker unknown event", "event_name", event.EventName)
		}

		return nil
	}
}

func mustConnectDB(ctx context.Context, dsn string) *pgxpool.Pool {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		slog.Error("db connect failed", "err", err)
		os.Exit(1)
	}
	if err := pool.Ping(ctx); err != nil {
		slog.Error("db ping failed", "err", err)
		os.Exit(1)
	}
	return pool
}

package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	msgapi "message/api"
	"message/config"
	msgkafka "message/infrastructure/kafka"
	"message/outbox"
	"message/repository"
	"message/service"
)

// main wires the API process:
//   - HTTP server: POST /message/send, GET /health
//   - Outbox relay goroutine: polls event_log for unpublished events → Kafka
//
// The worker process (cmd/worker) handles Kafka consumption and reconciliation.
func main() {
	cfg := config.Load()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool := mustConnectDB(ctx, cfg.DBDSN)
	defer pool.Close()

	// The API process only needs the event_log repo — the sender writes SENT events only.
	// The conversation repo is used by the worker's processor for the b→d transition.
	eventLogRepo := repository.NewEventLogRepo(pool)

	// Sender service: handles the a→b SENT transition.
	sender := service.NewSender(eventLogRepo)

	// Kafka producer: used by the outbox relay to publish events.
	// The relay uses conversationID as the partition key for per-conversation ordering.
	producer := msgkafka.NewProducer(strings.Split(cfg.KafkaBrokers, ","), cfg.KafkaTopic)
	defer producer.Close()

	// Outbox relay: polls unpublished events on a 2-second ticker and forwards
	// them to Kafka. Runs alongside the HTTP server in this process.
	relay := outbox.NewRelay(eventLogRepo, producer, cfg.OutboxBatch)
	go relay.Run(ctx)

	// HTTP server.
	handler := msgapi.NewHandler(sender)
	mux := http.NewServeMux()
	mux.HandleFunc("POST /message/send", handler.SendMessage)
	mux.HandleFunc("GET /health", handler.Health)

	srv := &http.Server{Addr: cfg.Port, Handler: mux}
	go func() {
		slog.Info("api listening", "addr", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "err", err)
		}
	}()

	<-ctx.Done()
	slog.Info("api shutting down")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	srv.Shutdown(shutdownCtx) //nolint:errcheck
}

func mustConnectDB(ctx context.Context, dsn string) *pgxpool.Pool {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		slog.Error("db connect failed", "err", err)
		os.Exit(1)
	}
	// Verify connectivity on startup rather than at first request.
	if err := pool.Ping(ctx); err != nil {
		slog.Error("db ping failed", "err", err)
		os.Exit(1)
	}
	return pool
}

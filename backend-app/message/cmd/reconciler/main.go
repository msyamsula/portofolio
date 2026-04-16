package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/jackc/pgx/v5/pgxpool"

	"message/config"
	"message/reconciler"
	"message/repository"
	"message/service"
)

// main wires the Reconciler process.
//
// The reconciler runs as its own container so it can be scaled, restarted,
// or paused independently from the API and worker processes.
// Multiple reconciler replicas are safe: the version guard on the event_log
// ensures that concurrent reconcilers racing on the same message produce
// at most one winner per version slot — all others get ON CONFLICT DO NOTHING.
func main() {
	cfg := config.Load()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool := mustConnectDB(ctx, cfg.DBDSN)
	defer pool.Close()

	eventLogRepo := repository.NewEventLogRepo(pool)
	convRepo := repository.NewConversationRepo(pool)
	processor := service.NewProcessor(eventLogRepo, convRepo)

	rec := reconciler.New(eventLogRepo, convRepo, processor, cfg)

	slog.Info("reconciler starting", "interval", cfg.ReconcileInterval)
	rec.Run(ctx) // blocks until ctx is cancelled
	slog.Info("reconciler shut down")
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

package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"auto-pr-review/internal/config"
	"auto-pr-review/internal/kafkaproducer"
	"auto-pr-review/internal/orchestrator"
	"auto-pr-review/internal/redisstore"
	"auto-pr-review/internal/service"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	cfg, err := config.Load()
	if err != nil {
		logger.Error("load config", slog.String("error", err.Error()))
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	jobStore, err := redisstore.NewRepository(ctx, cfg.Redis)
	if err != nil {
		logger.Error("connect redis", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer func() {
		if err := jobStore.Close(); err != nil {
			logger.Error("close redis", slog.String("error", err.Error()))
		}
	}()

	kafkaProducer, err := kafkaproducer.NewProducer(cfg.Kafka)
	if err != nil {
		logger.Error("connect kafka", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer func() {
		if err := kafkaProducer.Close(); err != nil {
			logger.Error("close kafka", slog.String("error", err.Error()))
		}
	}()

	jobOrchestrator := orchestrator.NewPeriodicJobOrchestrator(
		jobStore.Client().Eval,
		redisstore.PeriodicJobNextTimestampKey,
		cfg.JobOrchestrator.Interval,
		logger,
		func(callCtx context.Context, job service.PeriodicJob) error {
			published, err := kafkaProducer.PublishPeriodicJob(callCtx, job)
			if err != nil {
				return err
			}
			logFn := logger.Info
			msg := "periodic job published"
			if !published {
				logFn = logger.Warn
				msg = "periodic job dead-lettered"
			}
			logFn(
				msg,
				slog.String("github_link", job.GitHubLink),
				slog.String("period", job.Period),
				slog.Int64("next_timestamp", job.NextTimestamp),
			)
			return nil
		},
	)

	logger.Info("job orchestrator started", slog.String("interval", cfg.JobOrchestrator.Interval.String()))
	jobOrchestrator.Run(ctx)
}

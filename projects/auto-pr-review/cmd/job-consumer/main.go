package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"auto-pr-review/internal/config"
	"auto-pr-review/internal/kafkaconsumer"
	"auto-pr-review/internal/kafkaproducer"

	"github.com/jackc/pgx/v5/pgxpool"
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

	pgPool, err := newPostgresPool(ctx, cfg.Database)
	if err != nil {
		logger.Error("connect postgres", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer pgPool.Close()

	dlqPublisher, err := kafkaproducer.NewProducer(cfg.Kafka)
	if err != nil {
		logger.Error("connect kafka publisher", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer func() {
		if err := dlqPublisher.Close(); err != nil {
			logger.Error("close kafka publisher", slog.String("error", err.Error()))
		}
	}()

	consumer, err := kafkaconsumer.NewConsumer(cfg.Kafka, cfg.KafkaConsumer, pgPool, dlqPublisher, logger)
	if err != nil {
		logger.Error("create kafka consumer", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer func() {
		if err := consumer.Close(); err != nil {
			logger.Error("close kafka consumer", slog.String("error", err.Error()))
		}
	}()

	logger.Info("job consumer started", slog.String("group_id", cfg.KafkaConsumer.GroupID), slog.String("topic", cfg.Kafka.Topic))
	if err := consumer.Run(ctx); err != nil && err != context.Canceled {
		logger.Error("run consumer", slog.String("error", err.Error()))
		os.Exit(1)
	}
}

func newPostgresPool(ctx context.Context, cfg config.DatabaseConfig) (*pgxpool.Pool, error) {
	poolConfig, err := pgxpool.ParseConfig(cfg.URL)
	if err != nil {
		return nil, err
	}

	poolConfig.MaxConns = cfg.MaxConns
	poolConfig.MinConns = cfg.MinConns
	poolConfig.MaxConnLifetime = cfg.MaxConnLifetime
	poolConfig.MaxConnIdleTime = cfg.MaxConnIdleTime

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, err
	}

	pingCtx, cancel := context.WithTimeout(ctx, cfg.ExecTimeout)
	defer cancel()
	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, err
	}

	return pool, nil
}

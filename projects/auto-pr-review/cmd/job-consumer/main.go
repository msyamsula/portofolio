package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"auto-pr-review/internal/config"
	"auto-pr-review/internal/kafkaconsumer"
	"auto-pr-review/internal/kafkaproducer"
	"auto-pr-review/internal/lockmanager"
	"auto-pr-review/internal/postgres"
	"auto-pr-review/internal/redisstore"
	"auto-pr-review/internal/scheduler"
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

	store, err := postgres.NewRepository(ctx, cfg.Database)
	if err != nil {
		logger.Error("connect postgres", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer store.Close()

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

	scheduler := scheduler.New(store, jobStore, logger)

	lockRedis, err := lockmanager.NewRedisBackend(lockmanager.RedisBackendConfig{
		Addr:         cfg.Redis.Addr,
		Password:     cfg.Redis.Password,
		DB:           cfg.Redis.DB,
		PoolSize:     cfg.Redis.PoolSize,
		MinIdleConns: cfg.Redis.MinIdleConns,
		ExecTimeout:  cfg.Redis.ExecTimeout,
	}, "lock:pr:")
	if err != nil {
		logger.Error("connect lock redis", slog.String("error", err.Error()))
		os.Exit(1)
	}
	lockManager := lockmanager.New(lockRedis, logger, 5*time.Minute)

	consumer, err := kafkaconsumer.NewPrReviewAgentConsumer(cfg, dlqPublisher, logger, scheduler, lockManager)
	if err != nil {
		logger.Error("create PR review agent consumer", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer func() {
		if err := consumer.Close(); err != nil {
			logger.Error("close kafka consumer", slog.String("error", err.Error()))
		}
	}()
	defer func() {
		if err := lockRedis.Close(); err != nil {
			logger.Error("close lock redis", slog.String("error", err.Error()))
		}
	}()

	logger.Info("job consumer started (pr-review agent)", slog.String("group_id", cfg.KafkaConsumer.GroupID), slog.String("topic", cfg.Kafka.Topic))
	if err := consumer.Run(ctx); err != nil && err != context.Canceled {
		logger.Error("run consumer", slog.String("error", err.Error()))
		os.Exit(1)
	}
}

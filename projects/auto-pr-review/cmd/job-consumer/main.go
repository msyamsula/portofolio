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

	consumer, err := kafkaconsumer.NewPrReviewConsumer(cfg, dlqPublisher, logger)
	if err != nil {
		logger.Error("create PR review consumer", slog.String("error", err.Error()))
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

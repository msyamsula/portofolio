package kafkaconsumer

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand"
	"time"

	"auto-pr-review/internal/config"
	"auto-pr-review/internal/kafkaproducer"
	"auto-pr-review/internal/schemaregistry"

	"github.com/IBM/sarama"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Consumer struct {
	group          sarama.ConsumerGroup
	topic          string
	dlqTopic       string
	dlqPublisher   *kafkaproducer.Producer
	pgPool         *pgxpool.Pool
	logger         *slog.Logger
	schemaRegistry *schemaregistry.Registry
	maxRetries     int
	retryBackoff   time.Duration
	attemptTimeout time.Duration
}

type consumerGroupHandler struct {
	consumer *Consumer
}

func NewConsumer(cfg config.KafkaConfig, consumerCfg config.KafkaConsumerConfig, pgPool *pgxpool.Pool, dlqPublisher *kafkaproducer.Producer, logger *slog.Logger) (*Consumer, error) {
	saramaConfig := sarama.NewConfig()
	saramaConfig.Version = sarama.V2_8_0_0
	saramaConfig.ClientID = cfg.ClientID
	saramaConfig.ChannelBufferSize = consumerCfg.MaxInFlight
	saramaConfig.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
	saramaConfig.Consumer.Offsets.Initial = sarama.OffsetOldest
	saramaConfig.Consumer.Return.Errors = true
	saramaConfig.Consumer.Group.Session.Timeout = 10 * time.Second
	saramaConfig.Consumer.Group.Heartbeat.Interval = 3 * time.Second

	group, err := sarama.NewConsumerGroup(cfg.Brokers, consumerCfg.GroupID, saramaConfig)
	if err != nil {
		return nil, fmt.Errorf("create kafka consumer group: %w", err)
	}

	registry, err := schemaregistry.New()
	if err != nil {
		if closeErr := group.Close(); closeErr != nil {
			return nil, fmt.Errorf("load schema registry: %w; close consumer group: %w", err, closeErr)
		}
		return nil, fmt.Errorf("load schema registry: %w", err)
	}

	return &Consumer{
		group:          group,
		topic:          cfg.Topic,
		dlqTopic:       cfg.DLQTopic,
		dlqPublisher:   dlqPublisher,
		pgPool:         pgPool,
		logger:         logger,
		schemaRegistry: registry,
		maxRetries:     consumerCfg.MaxRetries,
		retryBackoff:   consumerCfg.RetryBackoff,
		attemptTimeout: consumerCfg.AttemptTimeout,
	}, nil
}

func (c *Consumer) Close() error {
	return c.group.Close()
}

func (c *Consumer) Run(ctx context.Context) error {
	handler := &consumerGroupHandler{consumer: c}

	for {
		if err := c.group.Consume(ctx, []string{c.topic}, handler); err != nil {
			c.logger.Warn("consume kafka topic failed", slog.String("error", err.Error()))
		}
		if err := ctx.Err(); err != nil {
			return err
		}
	}
}

func (h *consumerGroupHandler) Setup(sarama.ConsumerGroupSession) error { return nil }

func (h *consumerGroupHandler) Cleanup(sarama.ConsumerGroupSession) error { return nil }

func (h *consumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		if err := h.consumer.handleMessage(session.Context(), msg); err != nil {
			h.consumer.logger.Warn(
				"kafka message processing failed",
				slog.String("error", err.Error()),
				slog.String("topic", msg.Topic),
				slog.Int("partition", int(msg.Partition)),
				slog.Int64("offset", msg.Offset),
			)
			return err
		}

		session.MarkMessage(msg, "")
	}

	return nil
}

func (c *Consumer) handleMessage(ctx context.Context, msg *sarama.ConsumerMessage) error {
	version := schemaregistry.PeriodicJobEnvelopeV2Version
	event, err := c.schemaRegistry.UnmarshalPeriodicJobEnvelope(msg.Value, version)
	if err != nil {
		event = schemaregistry.PeriodicJobEnvelope{
			SchemaVersion: int32(version),
			GitHubLink:    string(msg.Key),
			Period:        "",
			NextTimestamp: msg.Timestamp.UnixMilli(),
			Error:         err.Error(),
		}
		c.logger.Warn(
			"kafka message decode failed",
			slog.String("error", err.Error()),
			slog.String("topic", msg.Topic),
			slog.Int("partition", int(msg.Partition)),
			slog.Int64("offset", msg.Offset),
		)
	}

	var lastErr error
	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		if attempt > 0 {
			if err := sleepWithExponentialJitter(ctx, c.retryBackoff, attempt); err != nil {
				return err
			}
		}

		attemptCtx, cancel := context.WithTimeout(ctx, c.attemptTimeout)
		err = c.pgPool.Ping(attemptCtx)
		cancel()
		if err == nil {
			c.logger.Info(
				"postgres ping success",
				slog.String("github_link", event.GitHubLink),
				slog.String("idempotency_key", event.IdempotencyKey),
				slog.String("topic", msg.Topic),
				slog.Int("partition", int(msg.Partition)),
				slog.Int64("offset", msg.Offset),
				slog.Int("attempt", attempt+1),
			)
			return nil
		}

		lastErr = err
		c.logger.Warn(
			"postgres ping failed",
			slog.String("error", err.Error()),
			slog.String("github_link", event.GitHubLink),
			slog.String("idempotency_key", event.IdempotencyKey),
			slog.Int("attempt", attempt+1),
			slog.Int("max_retries", c.maxRetries),
			slog.String("topic", msg.Topic),
			slog.Int("partition", int(msg.Partition)),
			slog.Int64("offset", msg.Offset),
		)
	}

	event.Error = lastErr.Error()
	if err := c.dlqPublisher.PublishEvent(ctx, c.dlqTopic, event.GitHubLink, event, version); err != nil {
		return fmt.Errorf("publish dlq message: %w", err)
	}

	c.logger.Warn(
		"kafka message dead-lettered",
		slog.String("github_link", event.GitHubLink),
		slog.String("idempotency_key", event.IdempotencyKey),
		slog.String("topic", msg.Topic),
		slog.Int("partition", int(msg.Partition)),
		slog.Int64("offset", msg.Offset),
	)

	return nil
}

func sleepWithExponentialJitter(ctx context.Context, base time.Duration, attempt int) error {
	if base <= 0 {
		return nil
	}
	if attempt < 1 {
		attempt = 1
	}

	// Exponential backoff with a 5% additive jitter.
	delay := base << (attempt - 1)
	jitterMax := delay / 20
	if jitterMax <= 0 {
		jitterMax = base / 20
	}
	if jitterMax <= 0 {
		jitterMax = time.Millisecond
	}

	jitter := time.Duration(rand.Int63n(int64(jitterMax) + 1))
	wait := delay + jitter

	timer := time.NewTimer(wait)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

package kafkaconsumer

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math/rand"
	"time"

	"auto-pr-review/internal/config"
	"auto-pr-review/internal/kafkaproducer"
	"auto-pr-review/internal/lockmanager"
	"auto-pr-review/internal/schemaregistry"
	"auto-pr-review/internal/service"

	"github.com/IBM/sarama"
)

var ErrSkipMessage = errors.New("skip message")

type MessageProcessor interface {
	Process(ctx context.Context, event schemaregistry.PeriodicJobEnvelope) error
}

type MessageFinalizer interface {
	Finalize(ctx context.Context, event schemaregistry.PeriodicJobEnvelope) error
}

type MessageProcessorFunc func(context.Context, schemaregistry.PeriodicJobEnvelope) error

func (f MessageProcessorFunc) Process(ctx context.Context, event schemaregistry.PeriodicJobEnvelope) error {
	return f(ctx, event)
}

type MessageFinalizerFunc func(context.Context, schemaregistry.PeriodicJobEnvelope) error

func (f MessageFinalizerFunc) Finalize(ctx context.Context, event schemaregistry.PeriodicJobEnvelope) error {
	return f(ctx, event)
}

type PeriodicJobScheduleManager interface {
	SetSchedule(ctx context.Context, job service.PeriodicJob) error
}

type Consumer struct {
	group          sarama.ConsumerGroup
	topic          string
	dlqTopic       string
	dlqPublisher   *kafkaproducer.Producer
	logger         *slog.Logger
	schemaRegistry *schemaregistry.Registry
	processor      MessageProcessor
	finalizer      MessageFinalizer
	preProcess     func(ctx context.Context, key string) (release func(), skip bool, err error)
	maxRetries     int
	retryBackoff   time.Duration
	attemptTimeout time.Duration
}

type PrReviewConsumer struct {
	consumer  *Consumer
	reviewer  *service.PRReviewer
	scheduler PeriodicJobScheduleManager
	logger    *slog.Logger
	now       func() time.Time
}

type consumerGroupHandler struct {
	consumer *Consumer
}

func NewConsumer(cfg config.KafkaConfig, consumerCfg config.KafkaConsumerConfig, dlqPublisher *kafkaproducer.Producer, logger *slog.Logger, processor MessageProcessor) (*Consumer, error) {
	if processor == nil {
		return nil, fmt.Errorf("message processor is required")
	}

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
		logger:         logger,
		schemaRegistry: registry,
		processor:      processor,
		maxRetries:     consumerCfg.MaxRetries,
		retryBackoff:   consumerCfg.RetryBackoff,
		attemptTimeout: consumerCfg.AttemptTimeout,
	}, nil
}

func NewPrReviewConsumer(cfg config.Config, dlqPublisher *kafkaproducer.Producer, logger *slog.Logger) (*PrReviewConsumer, error) {
	return NewPrReviewConsumerWithScheduler(cfg, dlqPublisher, logger, nil, nil)
}

func NewPrReviewConsumerWithScheduler(cfg config.Config, dlqPublisher *kafkaproducer.Producer, logger *slog.Logger, scheduler PeriodicJobScheduleManager, lockManager *lockmanager.Manager) (*PrReviewConsumer, error) {
	reviewer, err := service.NewPRReviewer(service.PRDiscussionReviewConfig{
		LLM:             cfg.LLM,
		GitHubTokenPath: cfg.GitHub.TokenPath,
		OutputDir:       cfg.PRReview.OutputDir,
	})
	if err != nil {
		return nil, fmt.Errorf("create PR reviewer: %w", err)
	}

	prReviewConsumer := &PrReviewConsumer{
		reviewer:  reviewer,
		scheduler: scheduler,
		logger:    logger,
		now:       time.Now,
	}

	consumer, err := NewConsumer(cfg.Kafka, cfg.KafkaConsumer, dlqPublisher, logger, MessageProcessorFunc(prReviewConsumer.process))
	if err != nil {
		return nil, err
	}
	consumer.finalizer = MessageFinalizerFunc(prReviewConsumer.reschedule)

	if lockManager != nil {
		lockTTL := cfg.KafkaConsumer.LockTTL
		consumer.preProcess = func(ctx context.Context, key string) (func(), bool, error) {
			token, err := lockManager.AcquireWithTTL(ctx, key, lockTTL)
			if err != nil {
				if errors.Is(err, lockmanager.ErrLockHeld) {
					return nil, true, nil
				}
				return nil, false, err
			}
			release := func() {
				if err := lockManager.Release(context.Background(), key, token); err != nil {
					logger.Error("release lock failed",
						slog.String("error", err.Error()))
				}
			}
			return release, false, nil
		}
	}

	prReviewConsumer.consumer = consumer

	return prReviewConsumer, nil
}

func (c *Consumer) Close() error {
	return c.group.Close()
}

func (c *PrReviewConsumer) Close() error {
	return c.consumer.Close()
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

func (c *PrReviewConsumer) Run(ctx context.Context) error {
	return c.consumer.Run(ctx)
}

func (h *consumerGroupHandler) Setup(sarama.ConsumerGroupSession) error { return nil }

func (h *consumerGroupHandler) Cleanup(sarama.ConsumerGroupSession) error { return nil }

func (h *consumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		err := h.consumer.handleMessage(session.Context(), msg)
		if errors.Is(err, ErrSkipMessage) {
			continue
		}
		if err != nil {
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

	if c.preProcess != nil {
		release, skip, err := c.preProcess(ctx, event.GitHubLink)
		if skip {
			c.logger.Info("message skipped, lock held by another consumer",
				slog.String("github_link", event.GitHubLink))
			return ErrSkipMessage
		}
		if err != nil {
			return fmt.Errorf("pre-process: %w", err)
		}
		defer release()
	}

	var lastErr error
	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		if attempt > 0 {
			if err := sleepWithExponentialJitter(ctx, c.retryBackoff, attempt); err != nil {
				return err
			}
		}

		attemptCtx, cancel := context.WithTimeout(ctx, c.attemptTimeout)
		err = c.processor.Process(attemptCtx, event)
		cancel()
		if err == nil {
			c.logger.Info(
				"kafka message processed",
				slog.String("github_link", event.GitHubLink),
				slog.String("idempotency_key", event.IdempotencyKey),
				slog.String("topic", msg.Topic),
				slog.Int("partition", int(msg.Partition)),
				slog.Int64("offset", msg.Offset),
				slog.Int("attempt", attempt+1),
			)
			return c.finalizeMessage(ctx, event)
		}

		lastErr = err
		if shouldStopRetrying(err) {
			c.logger.Warn(
				"kafka message processing failed permanently",
				slog.String("error", err.Error()),
				slog.String("github_link", event.GitHubLink),
				slog.String("idempotency_key", event.IdempotencyKey),
				slog.Int("attempt", attempt+1),
				slog.String("topic", msg.Topic),
				slog.Int("partition", int(msg.Partition)),
				slog.Int64("offset", msg.Offset),
			)
			break
		}
		c.logger.Warn(
			"kafka message processing attempt failed",
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

	return c.finalizeMessage(ctx, event)
}

func (c *Consumer) finalizeMessage(ctx context.Context, event schemaregistry.PeriodicJobEnvelope) error {
	if c.finalizer == nil {
		return nil
	}
	if err := c.finalizer.Finalize(ctx, event); err != nil {
		return fmt.Errorf("finalize kafka message: %w", err)
	}
	return nil
}

func shouldStopRetrying(err error) bool {
	var githubErr *service.GitHubAPIError
	if errors.As(err, &githubErr) {
		return githubErr.StatusCode == 404
	}
	return false
}

func (c *PrReviewConsumer) process(ctx context.Context, event schemaregistry.PeriodicJobEnvelope) error {
	if event.GitHubLink == "" {
		return fmt.Errorf("github link is required")
	}

	outputPath, err := c.reviewer.WriteReviewPlanMarkdownToDefaultFile(ctx, event.GitHubLink)
	if err != nil {
		return err
	}

	return c.consumer.logReviewOutput(event, outputPath)
}

func (c *PrReviewConsumer) reschedule(ctx context.Context, event schemaregistry.PeriodicJobEnvelope) error {
	if c.scheduler == nil {
		return nil
	}
	if event.GitHubLink == "" || event.Period == "" {
		c.logger.Warn(
			"periodic job reschedule skipped",
			slog.String("github_link", event.GitHubLink),
			slog.String("period", event.Period),
			slog.String("idempotency_key", event.IdempotencyKey),
		)
		return nil
	}

	period, err := time.ParseDuration(event.Period)
	if err != nil || period < time.Second {
		return fmt.Errorf("invalid periodic job period %q", event.Period)
	}

	job := service.PeriodicJob{
		GitHubLink:     event.GitHubLink,
		Period:         event.Period,
		NextTimestamp:  c.now().Add(period).Unix(),
		IdempotencyKey: event.IdempotencyKey,
	}
	if err := c.scheduler.SetSchedule(ctx, job); err != nil {
		return err
	}

	c.logger.Info(
		"periodic job rescheduled",
		slog.String("github_link", job.GitHubLink),
		slog.String("period", job.Period),
		slog.Int64("next_timestamp", job.NextTimestamp),
		slog.String("idempotency_key", job.IdempotencyKey),
	)

	return nil
}

func (c *Consumer) logReviewOutput(event schemaregistry.PeriodicJobEnvelope, outputPath string) error {
	c.logger.Info(
		"PR review written",
		slog.String("github_link", event.GitHubLink),
		slog.String("idempotency_key", event.IdempotencyKey),
		slog.String("output_path", outputPath),
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

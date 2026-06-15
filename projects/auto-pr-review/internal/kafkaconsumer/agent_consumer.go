package kafkaconsumer

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"auto-pr-review/internal/config"
	"auto-pr-review/internal/kafkaproducer"
	"auto-pr-review/internal/lockmanager"
	"auto-pr-review/internal/prreviewagent"
	"auto-pr-review/internal/schemaregistry"
	"auto-pr-review/internal/service"
)

type PrReviewAgentConsumer struct {
	consumer  *Consumer
	executor  *prreviewagent.Executor
	scheduler PeriodicJobScheduleManager
	logger    *slog.Logger
	now       func() time.Time
}

func NewPrReviewAgentConsumer(
	cfg config.Config,
	dlqPublisher *kafkaproducer.Producer,
	logger *slog.Logger,
	scheduler PeriodicJobScheduleManager,
	lockManager *lockmanager.Manager,
) (*PrReviewAgentConsumer, error) {
	executor, err := prreviewagent.New(prreviewagent.Config{
		GitHubTokenPath: cfg.GitHub.TokenPath,
		LLM:             cfg.LLM,
		WorkspaceRoot:   cfg.PRReviewAgent.WorkspaceRoot,
		DevelopBranch:   cfg.PRReviewAgent.DevelopBranch,
		OutputDir:       cfg.PRReviewAgent.OutputDir,
		GitUser:         cfg.PRReviewAgent.GitUser,
		GitEmail:        cfg.PRReviewAgent.GitEmail,
	})
	if err != nil {
		return nil, err
	}

	prReviewAgentConsumer := &PrReviewAgentConsumer{
		executor:  executor,
		scheduler: scheduler,
		logger:    logger,
		now:       time.Now,
	}

	consumer, err := NewConsumer(cfg.Kafka, cfg.KafkaConsumer, dlqPublisher, logger, MessageProcessorFunc(prReviewAgentConsumer.process))
	if err != nil {
		return nil, err
	}
	consumer.finalizer = MessageFinalizerFunc(prReviewAgentConsumer.reschedule)

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
					logger.Error("release lock failed", slog.String("error", err.Error()))
				}
			}
			return release, false, nil
		}
	}

	prReviewAgentConsumer.consumer = consumer

	return prReviewAgentConsumer, nil
}

func (c *PrReviewAgentConsumer) Close() error {
	return c.consumer.Close()
}

func (c *PrReviewAgentConsumer) Run(ctx context.Context) error {
	return c.consumer.Run(ctx)
}

func (c *PrReviewAgentConsumer) process(ctx context.Context, event schemaregistry.PeriodicJobEnvelope) error {
	if event.GitHubLink == "" {
		return nil
	}

	_, err := c.executor.Execute(ctx, event.GitHubLink)
	return err
}

func (c *PrReviewAgentConsumer) reschedule(ctx context.Context, event schemaregistry.PeriodicJobEnvelope) error {
	if c.scheduler == nil {
		return nil
	}
	if event.GitHubLink == "" {
		return nil
	}

	job := service.PeriodicJob{GitHubLink: event.GitHubLink}
	if err := c.scheduler.SetSchedule(ctx, job); err != nil {
		return err
	}

	c.logger.Info("periodic job rescheduled",
		slog.String("github_link", job.GitHubLink),
	)

	return nil
}

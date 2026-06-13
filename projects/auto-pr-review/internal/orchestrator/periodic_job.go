package orchestrator

import (
	"context"
	"log/slog"
	"time"

	"auto-pr-review/internal/service"
)

type PeriodicJobClaimer interface {
	ClaimDuePeriodicJob(ctx context.Context, now time.Time) (service.PeriodicJob, bool, error)
}

type PeriodicJobDoneFunc func(ctx context.Context, job service.PeriodicJob) error

type PeriodicJobOrchestrator struct {
	claimer  PeriodicJobClaimer
	interval time.Duration
	logger   *slog.Logger
	done     PeriodicJobDoneFunc
}

func NewPeriodicJobOrchestrator(
	claimer PeriodicJobClaimer,
	interval time.Duration,
	logger *slog.Logger,
	done PeriodicJobDoneFunc,
) *PeriodicJobOrchestrator {
	return &PeriodicJobOrchestrator{
		claimer:  claimer,
		interval: interval,
		logger:   logger,
		done:     done,
	}
}

func (o *PeriodicJobOrchestrator) Run(ctx context.Context) {
	ticker := time.NewTicker(o.interval)
	defer ticker.Stop()

	for {
		o.runOnce(ctx)

		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func (o *PeriodicJobOrchestrator) runOnce(ctx context.Context) {
	job, ok, err := o.claimer.ClaimDuePeriodicJob(ctx, time.Now())
	if err != nil {
		o.logger.Warn("claim periodic job failed", slog.String("error", err.Error()))
		return
	}
	if !ok {
		return
	}

	if err := o.done(ctx, job); err != nil {
		o.logger.Warn("periodic job dispatch failed", slog.String("error", err.Error()))
		return
	}
}

package orchestrator

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"log/slog"
	"time"

	"auto-pr-review/internal/service"
)

type PeriodicJobClaimer interface {
	ClaimDuePeriodicJob(ctx context.Context, now time.Time) (service.PeriodicJob, bool, error)
}

type PeriodicJobDoneFunc func(ctx context.Context, job service.PeriodicJob) error

type IdempotencyKeyFunc func() (string, error)

type PeriodicJobOrchestrator struct {
	claimer        PeriodicJobClaimer
	interval       time.Duration
	logger         *slog.Logger
	done           PeriodicJobDoneFunc
	idempotencyKey IdempotencyKeyFunc
}

func NewPeriodicJobOrchestrator(
	claimer PeriodicJobClaimer,
	interval time.Duration,
	logger *slog.Logger,
	done PeriodicJobDoneFunc,
) *PeriodicJobOrchestrator {
	return &PeriodicJobOrchestrator{
		claimer:        claimer,
		interval:       interval,
		logger:         logger,
		done:           done,
		idempotencyKey: NewIdempotencyKey,
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

	idempotencyKey, err := o.idempotencyKey()
	if err != nil {
		o.logger.Warn("generate periodic job idempotency key failed", slog.String("error", err.Error()))
		return
	}
	job.IdempotencyKey = idempotencyKey

	if err := o.done(ctx, job); err != nil {
		o.logger.Warn("periodic job dispatch failed", slog.String("error", err.Error()))
		return
	}
}

func NewIdempotencyKey() (string, error) {
	var key [16]byte
	if _, err := io.ReadFull(rand.Reader, key[:]); err != nil {
		return "", fmt.Errorf("generate idempotency key: %w", err)
	}

	return hex.EncodeToString(key[:]), nil
}

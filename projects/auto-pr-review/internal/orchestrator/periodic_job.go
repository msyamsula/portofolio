package orchestrator

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"log/slog"
	"strconv"
	"time"

	"auto-pr-review/internal/service"

	"github.com/redis/go-redis/v9"
)

const claimPeriodicJobScript = `
local items = redis.call("ZRANGE", KEYS[1], 0, 0, "WITHSCORES")
if #items == 0 then
	return ""
end

local member = items[1]
local score = tonumber(items[2])

if score > tonumber(ARGV[1]) then
	return ""
end

local removed = redis.call("ZREM", KEYS[1], member)
if removed == 0 then
	return ""
end

return member
`

type PeriodicJobDoneFunc func(ctx context.Context, job service.PeriodicJob) error

type IdempotencyKeyFunc func() (string, error)

type PeriodicJobOrchestrator struct {
	sortedSetKey   string
	interval       time.Duration
	logger         *slog.Logger
	done           PeriodicJobDoneFunc
	idempotencyKey IdempotencyKeyFunc
	eval           func(ctx context.Context, script string, keys []string, args ...interface{}) *redis.Cmd
}

func NewPeriodicJobOrchestrator(
	eval func(ctx context.Context, script string, keys []string, args ...interface{}) *redis.Cmd,
	sortedSetKey string,
	interval time.Duration,
	logger *slog.Logger,
	done PeriodicJobDoneFunc,
) *PeriodicJobOrchestrator {
	return &PeriodicJobOrchestrator{
		eval:           eval,
		sortedSetKey:   sortedSetKey,
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
	job, ok, err := o.claimDuePeriodicJob(ctx, time.Now())
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

func (o *PeriodicJobOrchestrator) claimDuePeriodicJob(ctx context.Context, now time.Time) (service.PeriodicJob, bool, error) {
	member, err := o.eval(
		ctx,
		claimPeriodicJobScript,
		[]string{o.sortedSetKey},
		strconv.FormatInt(now.Unix(), 10),
	).Text()
	if err != nil {
		return service.PeriodicJob{}, false, fmt.Errorf("eval claim script: %w", err)
	}
	if member == "" {
		return service.PeriodicJob{}, false, nil
	}

	return service.PeriodicJob{GitHubLink: member}, true, nil
}

func NewIdempotencyKey() (string, error) {
	var key [16]byte
	if _, err := io.ReadFull(rand.Reader, key[:]); err != nil {
		return "", fmt.Errorf("generate idempotency key: %w", err)
	}

	return hex.EncodeToString(key[:]), nil
}

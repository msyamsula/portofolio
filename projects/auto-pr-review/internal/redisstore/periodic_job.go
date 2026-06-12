package redisstore

import (
	"context"
	"fmt"
	"time"

	"auto-pr-review/internal/config"
	"auto-pr-review/internal/service"

	"github.com/redis/go-redis/v9"
)

const PeriodicJobNextTimestampKey = "link:periodic:nexttimestamp"

type Repository struct {
	client      *redis.Client
	execTimeout time.Duration
}

func NewRepository(ctx context.Context, cfg config.RedisConfig) (*Repository, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         cfg.Addr,
		Password:     cfg.Password,
		DB:           cfg.DB,
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdleConns,
	})

	pingCtx, cancel := context.WithTimeout(ctx, cfg.ExecTimeout)
	defer cancel()
	if err := client.Ping(pingCtx).Err(); err != nil {
		if closeErr := client.Close(); closeErr != nil {
			return nil, fmt.Errorf("ping redis: %w; close redis: %w", err, closeErr)
		}
		return nil, fmt.Errorf("ping redis: %w", err)
	}

	return &Repository{
		client:      client,
		execTimeout: cfg.ExecTimeout,
	}, nil
}

func (r *Repository) Close() error {
	return r.client.Close()
}

func (r *Repository) RegisterPeriodicJob(ctx context.Context, job service.PeriodicJob) error {
	queryCtx, cancel := context.WithTimeout(ctx, r.execTimeout)
	defer cancel()

	err := r.client.ZAdd(queryCtx, PeriodicJobNextTimestampKey, redis.Z{
		Score:  float64(job.NextTimestamp),
		Member: periodicJobMember(job),
	}).Err()
	if err != nil {
		return fmt.Errorf("register periodic job: %w", err)
	}

	return nil
}

func periodicJobMember(job service.PeriodicJob) string {
	return fmt.Sprintf("%s:%s", job.GitHubLink, job.Period)
}

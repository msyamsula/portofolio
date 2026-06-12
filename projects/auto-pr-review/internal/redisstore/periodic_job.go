package redisstore

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"auto-pr-review/internal/config"
	"auto-pr-review/internal/service"

	"github.com/redis/go-redis/v9"
)

const PeriodicJobNextTimestampKey = "link:periodic:nexttimestamp"

const advancePeriodicJobScript = `
local current_score = redis.call("ZSCORE", KEYS[1], ARGV[1])
if not current_score then
	return 0
end
if tonumber(current_score) ~= tonumber(ARGV[2]) then
	return 0
end
return redis.call("ZADD", KEYS[1], "XX", "CH", ARGV[3], ARGV[1])
`

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

func (r *Repository) ListPeriodicJobs(ctx context.Context) ([]service.PeriodicJob, error) {
	queryCtx, cancel := context.WithTimeout(ctx, r.execTimeout)
	defer cancel()

	items, err := r.client.ZRangeWithScores(queryCtx, PeriodicJobNextTimestampKey, 0, -1).Result()
	if err != nil {
		return nil, fmt.Errorf("list periodic jobs: %w", err)
	}

	jobs := make([]service.PeriodicJob, 0, len(items))
	for _, item := range items {
		member, ok := item.Member.(string)
		if !ok {
			return nil, fmt.Errorf("periodic job member is not a string")
		}

		job, _, err := parsePeriodicJobMember(member)
		if err != nil {
			return nil, err
		}
		job.NextTimestamp = int64(item.Score)
		jobs = append(jobs, job)
	}

	return jobs, nil
}

func (r *Repository) DeregisterPeriodicJob(ctx context.Context, job service.PeriodicJob) error {
	queryCtx, cancel := context.WithTimeout(ctx, r.execTimeout)
	defer cancel()

	if err := r.client.ZRem(queryCtx, PeriodicJobNextTimestampKey, periodicJobMember(job)).Err(); err != nil {
		return fmt.Errorf("deregister periodic job: %w", err)
	}

	return nil
}

func (r *Repository) ClaimDuePeriodicJob(ctx context.Context, now time.Time) (service.PeriodicJob, bool, error) {
	queryCtx, cancel := context.WithTimeout(ctx, r.execTimeout)
	defer cancel()

	items, err := r.client.ZRangeWithScores(queryCtx, PeriodicJobNextTimestampKey, 0, 0).Result()
	if err != nil {
		return service.PeriodicJob{}, false, fmt.Errorf("get next periodic job: %w", err)
	}
	if len(items) == 0 {
		return service.PeriodicJob{}, false, nil
	}

	member, ok := items[0].Member.(string)
	if !ok {
		return service.PeriodicJob{}, false, fmt.Errorf("periodic job member is not a string")
	}

	nextTimestamp := int64(items[0].Score)
	if now.Unix() < nextTimestamp {
		return service.PeriodicJob{}, false, nil
	}

	job, period, err := parsePeriodicJobMember(member)
	if err != nil {
		return service.PeriodicJob{}, false, err
	}

	updatedTimestamp := now.Add(period).Unix()
	result, err := r.client.Eval(
		queryCtx,
		advancePeriodicJobScript,
		[]string{PeriodicJobNextTimestampKey},
		member,
		strconv.FormatInt(nextTimestamp, 10),
		strconv.FormatInt(updatedTimestamp, 10),
	).Int64()
	if err != nil {
		return service.PeriodicJob{}, false, fmt.Errorf("advance periodic job: %w", err)
	}
	if result != 1 {
		return service.PeriodicJob{}, false, nil
	}

	job.NextTimestamp = updatedTimestamp
	return job, true, nil
}

func periodicJobMember(job service.PeriodicJob) string {
	return fmt.Sprintf("%s:%s", job.GitHubLink, job.Period)
}

func parsePeriodicJobMember(member string) (service.PeriodicJob, time.Duration, error) {
	separator := strings.LastIndex(member, ":")
	if separator < 1 || separator == len(member)-1 {
		return service.PeriodicJob{}, 0, fmt.Errorf("invalid periodic job member")
	}

	period, err := time.ParseDuration(member[separator+1:])
	if err != nil || period < time.Second {
		return service.PeriodicJob{}, 0, fmt.Errorf("invalid periodic job period")
	}

	return service.PeriodicJob{
		GitHubLink: member[:separator],
		Period:     member[separator+1:],
	}, period, nil
}

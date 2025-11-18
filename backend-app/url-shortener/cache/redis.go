package cache

import (
	"context"
	"time"

	redisPkg "github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel"
)

type redis struct {
	db  *redisPkg.Client
	ttl time.Duration
}

func (r *redis) Get(ctx context.Context, key string) (string, error) {
	ctx, span := otel.Tracer("").Start(ctx, "redisPkg.getCache")
	defer span.End()

	cmd := r.db.Get(ctx, key)

	return cmd.Result()
}

func (r *redis) Set(ctx context.Context, key string, value string) error {
	ctx, redisSpan := otel.Tracer("").Start(ctx, "redis.setCache")
	defer redisSpan.End()

	cmd := r.db.Set(ctx, key, value, r.ttl)
	_, err := cmd.Result()
	return err
}

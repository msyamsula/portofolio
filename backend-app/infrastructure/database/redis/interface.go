package redis

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// Cache defines the interface for cache operations
type Cache interface {
	Get(ctx context.Context, key string) *redis.StringCmd
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd
	Del(ctx context.Context, keys ...string) *redis.IntCmd
}

// Ensure *redis.Client implements the Cache interface
var _ Cache = (*redis.Client)(nil)

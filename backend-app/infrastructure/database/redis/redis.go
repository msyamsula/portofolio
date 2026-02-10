package redis

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

// RedisConfig holds the Redis configuration
type RedisConfig struct {
	Host     string
	Password string
	DB       int
}

// RedisConnectFunc is a function that creates a Redis connection
type RedisConnectFunc func(ctx context.Context, opts *redis.Options) *redis.Client

// redisConnectFunc is the default Redis connection function
var redisConnectFunc RedisConnectFunc = func(ctx context.Context, opts *redis.Options) *redis.Client {
	return redis.NewClient(opts)
}

// configureRedisPoolFunc configures the Redis pool settings
type configureRedisPoolFunc func(client *redis.Client)

var configureRedisPool configureRedisPoolFunc = func(client *redis.Client) {
	// Pool size defaults are handled by go-redis, but we can customize if needed
}

// NewRedisClient creates a new Redis client connection
func NewRedisClient(ctx context.Context, cfg RedisConfig) *redis.Client {
	opts := &redis.Options{
		Addr:     cfg.Host,
		Password: cfg.Password,
		DB:       cfg.DB,
	}

	client := redisConnectFunc(ctx, opts)
	configureRedisPool(client)

	return client
}

// PingRedis tests the Redis connection
func PingRedis(ctx context.Context, client *redis.Client) error {
	_, err := client.Ping(ctx).Result()
	if err != nil {
		return fmt.Errorf("failed to ping redis: %w", err)
	}
	return nil
}

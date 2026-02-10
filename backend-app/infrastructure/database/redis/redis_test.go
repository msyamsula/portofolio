package redis

import (
	"context"
	"testing"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRedisClientSuccess(t *testing.T) {
	originalConnect := redisConnectFunc
	originalPool := configureRedisPool
	defer func() {
		redisConnectFunc = originalConnect
		configureRedisPool = originalPool
	}()

	ctx := context.Background()
	mockClient := &redis.Client{}
	redisConnectFunc = func(ctx context.Context, opts *redis.Options) *redis.Client {
		assert.Equal(t, "localhost:6379", opts.Addr)
		assert.Equal(t, "", opts.Password)
		assert.Equal(t, 0, opts.DB)
		return mockClient
	}
	configureRedisPool = func(client *redis.Client) {} // Skip pool config in test

	cfg := RedisConfig{
		Host:     "localhost:6379",
		Password: "",
		DB:       0,
	}

	client := NewRedisClient(ctx, cfg)
	assert.Same(t, mockClient, client)
}

func TestNewRedisClientWithPassword(t *testing.T) {
	originalConnect := redisConnectFunc
	originalPool := configureRedisPool
	defer func() {
		redisConnectFunc = originalConnect
		configureRedisPool = originalPool
	}()

	ctx := context.Background()
	redisConnectFunc = func(ctx context.Context, opts *redis.Options) *redis.Client {
		assert.Equal(t, "redis.example.com:6379", opts.Addr)
		assert.Equal(t, "secretpass", opts.Password)
		assert.Equal(t, 1, opts.DB)
		return &redis.Client{}
	}
	configureRedisPool = func(client *redis.Client) {} // Skip pool config in test

	cfg := RedisConfig{
		Host:     "redis.example.com:6379",
		Password: "secretpass",
		DB:       1,
	}

	client := NewRedisClient(ctx, cfg)
	assert.NotNil(t, client)
}

func TestNewRedisClientDifferentDB(t *testing.T) {
	originalConnect := redisConnectFunc
	originalPool := configureRedisPool
	defer func() {
		redisConnectFunc = originalConnect
		configureRedisPool = originalPool
	}()

	ctx := context.Background()
	var capturedDB int
	redisConnectFunc = func(ctx context.Context, opts *redis.Options) *redis.Client {
		capturedDB = opts.DB
		return &redis.Client{}
	}
	configureRedisPool = func(client *redis.Client) {} // Skip pool config in test

	tests := []struct {
		name string
		db   int
	}{
		{"DB 0", 0},
		{"DB 1", 1},
		{"DB 2", 2},
		{"DB 15", 15},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := RedisConfig{
				Host:     "localhost:6379",
				Password: "",
				DB:       tt.db,
			}

			client := NewRedisClient(ctx, cfg)
			assert.NotNil(t, client)
			assert.Equal(t, tt.db, capturedDB)
		})
	}
}

func TestNewRedisClientCustomPort(t *testing.T) {
	originalConnect := redisConnectFunc
	originalPool := configureRedisPool
	defer func() {
		redisConnectFunc = originalConnect
		configureRedisPool = originalPool
	}()

	ctx := context.Background()
	redisConnectFunc = func(ctx context.Context, opts *redis.Options) *redis.Client {
		assert.Equal(t, "localhost:6380", opts.Addr)
		return &redis.Client{}
	}
	configureRedisPool = func(client *redis.Client) {} // Skip pool config in test

	cfg := RedisConfig{
		Host:     "localhost:6380",
		Password: "",
		DB:       0,
	}

	client := NewRedisClient(ctx, cfg)
	assert.NotNil(t, client)
}

func TestNewRedisClientEmptyConfig(t *testing.T) {
	originalConnect := redisConnectFunc
	originalPool := configureRedisPool
	defer func() {
		redisConnectFunc = originalConnect
		configureRedisPool = originalPool
	}()

	ctx := context.Background()
	redisConnectFunc = func(ctx context.Context, opts *redis.Options) *redis.Client {
		assert.Equal(t, "", opts.Addr)
		assert.Equal(t, "", opts.Password)
		assert.Equal(t, 0, opts.DB)
		return &redis.Client{}
	}
	configureRedisPool = func(client *redis.Client) {} // Skip pool config in test

	cfg := RedisConfig{
		Host:     "",
		Password: "",
		DB:       0,
	}

	client := NewRedisClient(ctx, cfg)
	assert.NotNil(t, client)
}

func TestNewRedisClientWithContext(t *testing.T) {
	originalConnect := redisConnectFunc
	originalPool := configureRedisPool
	defer func() {
		redisConnectFunc = originalConnect
		configureRedisPool = originalPool
	}()

	redisConnectFunc = func(ctx context.Context, opts *redis.Options) *redis.Client {
		return &redis.Client{}
	}
	configureRedisPool = func(client *redis.Client) {} // Skip pool config in test

	cfg := RedisConfig{
		Host:     "localhost:6379",
		Password: "",
		DB:       0,
	}

	ctx := context.Background()
	client := NewRedisClient(ctx, cfg)
	assert.NotNil(t, client)
}

func TestRedisConfigOptions(t *testing.T) {
	originalConnect := redisConnectFunc
	originalPool := configureRedisPool
	defer func() {
		redisConnectFunc = originalConnect
		configureRedisPool = originalPool
	}()

	ctx := context.Background()
	var capturedOpts *redis.Options
	redisConnectFunc = func(ctx context.Context, opts *redis.Options) *redis.Client {
		capturedOpts = opts
		return &redis.Client{}
	}
	configureRedisPool = func(client *redis.Client) {} // Skip pool config in test

	cfg := RedisConfig{
		Host:     "redis.prod.example.com:6379",
		Password: "prodpassword",
		DB:       5,
	}

	client := NewRedisClient(ctx, cfg)
	require.NotNil(t, client)

	assert.NotNil(t, capturedOpts)
	assert.Equal(t, "redis.prod.example.com:6379", capturedOpts.Addr)
	assert.Equal(t, "prodpassword", capturedOpts.Password)
	assert.Equal(t, 5, capturedOpts.DB)
}

func TestPingRedis(t *testing.T) {
	// This test would require a more sophisticated mock
	// For now, we just verify the function signature
	t.Skip("requires mock redis client with Ping implementation")
}

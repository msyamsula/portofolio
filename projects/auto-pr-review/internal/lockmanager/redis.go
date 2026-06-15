package lockmanager

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const releaseLockScript = `
if redis.call("GET", KEYS[1]) == ARGV[1] then
	return redis.call("DEL", KEYS[1])
end
return 0
`

type RedisBackendConfig struct {
	Addr         string
	Password     string
	DB           int
	PoolSize     int
	MinIdleConns int
	ExecTimeout  time.Duration
}

type RedisBackend struct {
	client      *redis.Client
	keyPrefix   string
	execTimeout time.Duration
}

func NewRedisBackend(cfg RedisBackendConfig, keyPrefix string) (*RedisBackend, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         cfg.Addr,
		Password:     cfg.Password,
		DB:           cfg.DB,
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdleConns,
	})

	pingCtx, cancel := context.WithTimeout(context.Background(), cfg.ExecTimeout)
	defer cancel()
	if err := client.Ping(pingCtx).Err(); err != nil {
		client.Close()
		return nil, fmt.Errorf("ping redis: %w", err)
	}

	return &RedisBackend{
		client:      client,
		keyPrefix:   keyPrefix,
		execTimeout: cfg.ExecTimeout,
	}, nil
}

func (b *RedisBackend) Close() error {
	return b.client.Close()
}

func (b *RedisBackend) Acquire(ctx context.Context, key string, ttl time.Duration) (string, bool, error) {
	var raw [16]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return "", false, fmt.Errorf("generate lock token: %w", err)
	}
	token := hex.EncodeToString(raw[:])

	queryCtx, cancel := context.WithTimeout(ctx, b.execTimeout)
	defer cancel()

	ok, err := b.client.SetNX(queryCtx, b.keyPrefix+key, token, ttl).Result()
	if err != nil {
		return "", false, fmt.Errorf("acquire lock: %w", err)
	}
	if !ok {
		return "", false, nil
	}

	return token, true, nil
}

func (b *RedisBackend) Release(ctx context.Context, key, token string) (bool, error) {
	queryCtx, cancel := context.WithTimeout(ctx, b.execTimeout)
	defer cancel()

	count, err := b.client.Eval(
		queryCtx,
		releaseLockScript,
		[]string{b.keyPrefix + key},
		token,
	).Int64()
	if err != nil {
		return false, fmt.Errorf("release lock: %w", err)
	}

	return count > 0, nil
}

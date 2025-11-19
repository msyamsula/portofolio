package cache

import (
	"context"
	"crypto/tls"
	"fmt"
	"time"

	redisPkg "github.com/redis/go-redis/v9"
)

type Repository interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value string) error
}

type Config struct {
	Host     string
	Port     string
	Password string
	Ttl      time.Duration
}

func New(cfg Config) Repository {
	client := redisPkg.NewClient(&redisPkg.Options{
		Addr:           fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Password:       cfg.Password, // No password set
		DB:             0,            // Use default DB
		Protocol:       2,            // Connection protocol
		PoolSize:       10,
		TLSConfig:      &tls.Config{},
		DialTimeout:    2 * time.Second,
		PoolTimeout:    1 * time.Second,
		MaxActiveConns: 7,
		MinIdleConns:   3,
		MaxIdleConns:   3,
		ReadTimeout:    0,
		WriteTimeout:   0,
	})
	return &redis{
		db:  client,
		ttl: cfg.Ttl,
	}
}

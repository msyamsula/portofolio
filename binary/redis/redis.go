package redis

import (
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type Redis struct {
	*redis.Client
	Ttl time.Duration
}

type Config struct {
	Host     string
	Port     string
	Password string
	Ttl      time.Duration
}

func New(cfg Config) *Redis {
	client := redis.NewClient(&redis.Options{
		Addr:           fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Password:       cfg.Password, // No password set
		DB:             0,            // Use default DB
		Protocol:       2,            // Connection protocol
		PoolSize:       10,
		DialTimeout:    2 * time.Second,
		PoolTimeout:    1 * time.Second,
		MaxActiveConns: 7,
		MinIdleConns:   3,
		MaxIdleConns:   3,
		ReadTimeout:    0,
		WriteTimeout:   0,
	})
	return &Redis{
		Client: client,
		Ttl:    cfg.Ttl,
	}
}

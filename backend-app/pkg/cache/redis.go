package cache

import (
	"context"
	"crypto/tls"
	"fmt"
	"time"

	"github.com/msyamsula/portofolio/backend-app/observability/logger"
	"github.com/redis/go-redis/extra/redisotel/v9"
	redisPkg "github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
)

type redis struct {
	db  *redisPkg.Client
	ttl time.Duration
}

func NewRedis(cfg RedisConfig, options *redisPkg.Options) *redisPkg.Client {
	var tlsConfig *tls.Config
	if cfg.Env == "production" {
		tlsConfig = &tls.Config{
			InsecureSkipVerify: true,
		}
	}

	options.Addr = fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)

	client := redisPkg.NewClient(&redisPkg.Options{
		Addr:           fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		DB:             0, // Use default DB
		Protocol:       2, // Connection protocol
		PoolSize:       10,
		TLSConfig:      tlsConfig,
		DialTimeout:    2 * time.Second,
		PoolTimeout:    1 * time.Second,
		MaxActiveConns: 7,
		MinIdleConns:   5,
		MaxIdleConns:   20,
		ReadTimeout:    0,
		WriteTimeout:   0,
		MaxRetries:     5,
	})
	ping := client.Ping(context.Background())
	_, err := ping.Result()
	if err != nil {
		logger.Logger.Fatalf("ping %s failed, %v", options.Addr, err.Error())
	}
	redisotel.InstrumentTracing(client)
	return client
}

func (r *redis) Get(c context.Context, key string) (string, error) {
	var err error
	ctx, span := otel.Tracer("cache").Start(c, "Redis Get")
	defer func() {
		if err != nil {
			logger.Logger.Info(err.Error())
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}
		span.End()
	}()

	cmd := r.db.Get(ctx, key)

	var value string
	value, err = cmd.Result()
	return value, err
}

func (r *redis) Set(c context.Context, key string, value string) error {
	var err error
	ctx, span := otel.Tracer("redis").Start(c, "Redis Set")
	defer func() {
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}

		span.End()
	}()

	cmd := r.db.Set(ctx, key, value, r.ttl)
	_, err = cmd.Result()
	return err
}

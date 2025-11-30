package cache

import (
	"context"
	"time"

	"github.com/msyamsula/portofolio/backend-app/observability/logger"
	redisPkg "github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
)

type redis struct {
	db *redisPkg.Client
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

func (r *redis) Set(c context.Context, key string, value string, ttl time.Duration) error {
	var err error
	ctx, span := otel.Tracer("redis").Start(c, "Redis Set")
	defer func() {
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}

		span.End()
	}()

	cmd := r.db.Set(ctx, key, value, ttl)
	_, err = cmd.Result()
	return err
}

func (r *redis) Del(c context.Context, key string) error {
	var err error
	ctx, span := otel.Tracer("redis").Start(c, "redis del")
	defer func() {
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}

		span.End()
	}()

	cmd := r.db.Del(ctx, key)
	_, err = cmd.Result()
	return err
}

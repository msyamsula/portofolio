package cache

import (
	"context"
	"fmt"
	"time"

	redisPkg "github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
)

type redis struct {
	db  *redisPkg.Client
	ttl time.Duration
}

func (r *redis) Get(c context.Context, key string) (string, error) {
	var err error
	ctx, span := otel.Tracer("cache").Start(c, "Redis Get")
	defer func() {
		fmt.Println(err, "err")
		if err != nil {
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

package repository

import (
	"context"
	"fmt"
	"strconv"

	"github.com/msyamsula/portofolio/binary/redis"
	"go.opentelemetry.io/otel"
)

type Cache struct {
	*redis.Redis
}

func (s *Cache) SetUser(c context.Context, user User) error {
	ctx, span := otel.Tracer("").Start(c, "repository.cache.SetUser")
	defer span.End()

	var err error
	defer func() {
		if err != nil {
			span.RecordError(err)
		}
	}()

	value := map[string]interface{}{
		"id":       fmt.Sprintf("%d", user.Id),
		"username": user.Username,
		"online":   user.Online,
	}
	cmd := s.HSet(ctx, user.Username, value)
	_, err = cmd.Result()
	if err != nil {
		return err
	}

	ttlCmd := s.Expire(ctx, user.Username, s.Ttl)
	_, err = ttlCmd.Result()

	return err
}

func (s *Cache) GetUser(c context.Context, username string) (User, error) {
	ctx, span := otel.Tracer("").Start(c, "respository.cache.GetUser")
	defer span.End()

	var err error
	defer func() {
		if err != nil {
			span.RecordError(err)
		}
	}()

	cmd := s.HGetAll(ctx, username)
	result, err := cmd.Result()
	if err != nil {
		return User{}, err
	}

	var id int64
	id, err = strconv.ParseInt(result["id"], 10, 64)
	if err != nil {
		return User{}, err
	}

	var isOnline bool
	isOnline, err = strconv.ParseBool(result["online"])
	if err != nil {
		return User{}, err
	}

	user := User{
		Username: result["username"],
		Id:       id,
		Online:   isOnline,
	}
	return user, nil
}

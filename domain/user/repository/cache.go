package repository

import (
	"context"
	"fmt"
	"strconv"

	"github.com/msyamsula/portofolio/tech-stack/redis"
)

type Cache struct {
	*redis.Redis
}

func (s *Cache) SetUser(c context.Context, user User) error {
	value := map[string]interface{}{
		"id":       fmt.Sprintf("%d", user.Id),
		"username": user.Username,
	}
	cmd := s.HSet(c, user.Username, value)
	_, err := cmd.Result()
	if err != nil {
		return err
	}

	ttlCmd := s.Expire(c, user.Username, s.Ttl)
	_, err = ttlCmd.Result()

	return err
}

func (s *Cache) GetUser(c context.Context, username string) (User, error) {
	cmd := s.HGetAll(c, username)
	result, err := cmd.Result()
	if err != nil {
		return User{}, err
	}

	id, err := strconv.ParseInt(result["id"], 10, 64)
	if err != nil {
		return User{}, err
	}
	user := User{
		Username: result["username"],
		Id:       id,
	}
	return user, nil
}

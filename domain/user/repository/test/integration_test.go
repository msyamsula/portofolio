//go:build integration

package test

import (
	"context"
	"testing"
	"time"

	"github.com/msyamsula/portofolio/domain/user/repository"
	"github.com/msyamsula/portofolio/domain/utils"
	"github.com/msyamsula/portofolio/tech-stack/postgres"
	"github.com/msyamsula/portofolio/tech-stack/redis"
	"github.com/stretchr/testify/assert"
)

var (
	config = postgres.Config{
		Username: "admin",
		Password: "admin",
		DbName:   "postgres",
		Host:     "0.0.0.0",
		Port:     "5432",
	}

	pg          = postgres.New(config)
	persistence = repository.Persistence{
		Postgres: pg,
	}
	c = context.Background()

	redisConfig = redis.Config{
		Host:     "0.0.0.0",
		Port:     "6379",
		Password: "admin",
		Ttl:      300 * time.Second,
	}

	cache = repository.Cache{
		Redis: redis.New(redisConfig),
	}
)

func TestIntegration(t *testing.T) {
	username := utils.RandomName(20) // make sure you enter new cacheUser
	user, err := persistence.InsertUser(c, username)
	assert.Nil(t, err)
	u, err := persistence.GetUser(c, username)
	assert.Nil(t, err)
	assert.Equal(t, user, u)

	cacheUser := repository.User{
		Username: username,
		Id:       100,
	}
	err = cache.SetUser(c, cacheUser)
	assert.Nil(t, err)
	result, err := cache.GetUser(c, cacheUser.Username)
	assert.Nil(t, err)
	assert.Equal(t, cacheUser, result)
}

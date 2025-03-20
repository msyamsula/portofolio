//go:build integration

package test

import (
	"context"
	"testing"
	"time"

	"github.com/msyamsula/portofolio/domain/user/repository"
	"github.com/msyamsula/portofolio/domain/user/service"
	"github.com/msyamsula/portofolio/domain/utils"
	"github.com/msyamsula/portofolio/tech-stack/postgres"
	"github.com/msyamsula/portofolio/tech-stack/redis"
	"github.com/stretchr/testify/assert"
)

var (
	pgConfig = postgres.Config{
		Username: "admin",
		Password: "admin",
		DbName:   "postgres",
		Host:     "0.0.0.0",
		Port:     "5432",
	}
	rdConfig = redis.Config{
		Host:     "0.0.0.0",
		Port:     "6379",
		Password: "admin",
		Ttl:      500 * time.Second,
	}

	persistence = &repository.Persistence{
		Postgres: postgres.New(pgConfig),
	}
	cache = &repository.Cache{
		Redis: redis.New(rdConfig),
	}
	svc = service.New(service.Dependencies{
		Persistence: persistence,
		Cache:       cache,
	})

	ctx = context.Background()
)

func TestIntegrationUser(t *testing.T) {

	user := repository.User{
		Username: utils.RandomName(20), //use new unique name
	}

	var err error
	user, err = svc.SetUser(ctx, user)
	assert.Nil(t, err)
	var u repository.User
	u, err = svc.GetUser(ctx, user.Username)
	assert.Equal(t, user, u)
}

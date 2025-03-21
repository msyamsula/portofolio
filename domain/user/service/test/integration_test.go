//go:build integration

package test

import (
	"context"
	"fmt"
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

func TestIntegrationFriend(t *testing.T) {

	// check both user to db
	userA := repository.User{
		Username: "",
		Id:       35,
	}
	userB := repository.User{
		Username: "",
		Id:       37,
	}

	err := svc.AddFriend(ctx, userA, userB)
	assert.Nil(t, err)

	users, err := svc.GetFriends(ctx, repository.User{
		Username: "admin",
		Id:       21,
	})
	assert.Nil(t, err)
	assert.NotZero(t, users)
	for _, u := range users {
		fmt.Println(u)
	}
}

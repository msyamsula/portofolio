//go:build integration

package test

import (
	"context"
	"fmt"
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

func TestIntegrationAddFriend(t *testing.T) {

	existA := 17
	existB := 16
	type testCase struct {
		IdA, idB int64
		isErr    bool
	}
	testCases := []testCase{
		{
			// same case
			IdA:   0,
			idB:   0,
			isErr: true,
		},
		{
			// user does not exist
			IdA:   999999999999,
			idB:   -12312,
			isErr: true,
		},
		{
			// success, check the db first before execution
			IdA:   int64(existA),
			idB:   int64(existB),
			isErr: false,
		},
		{
			// inverse add, existing
			IdA:   int64(existB),
			idB:   int64(existA),
			isErr: true,
		},
	}

	for _, tt := range testCases {
		userA := repository.User{
			Id: tt.IdA,
		}
		userB := repository.User{
			Id: tt.idB,
		}
		err := persistence.AddFriend(c, userA, userB)
		fmt.Println(err)
		assert.Equal(t, tt.isErr, err != nil)
	}
}

func TestIntegrationGetFriends(t *testing.T) {

	user := repository.User{
		Id: 21,
	}

	c := context.Background()
	users, err := persistence.GetFriends(c, user)
	assert.Nil(t, err)
	assert.NotZero(t, users)
	for _, u := range users {
		fmt.Println(u.Username)
	}
}

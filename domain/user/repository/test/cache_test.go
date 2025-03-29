package test

import (
	"context"
	"time"

	"github.com/go-redis/redismock/v9"
	"github.com/msyamsula/portofolio/domain/user/repository"
	"github.com/msyamsula/portofolio/tech-stack/redis"
)

func (s *RepositoryTestSuite) TestCacheSetUser() {
	db, mock := redismock.NewClientMock()
	ttl := 300 * time.Second
	cache := &repository.Cache{
		Redis: &redis.Redis{
			Client: db,
			Ttl:    ttl,
		},
	}

	type (
		args struct {
			c    context.Context
			user repository.User
		}
		want struct {
			user repository.User
			err  error
		}
		testCase struct {
			name     string
			args     args
			want     want
			mockFunc func()
		}
	)

	testCases := []testCase{
		{
			name: "succes",
			args: args{
				c: context.Background(),
				user: repository.User{
					Username: "admin",
					Id:       10,
					Online:   true,
				},
			},
			want: want{},
			mockFunc: func() {
				value := map[string]interface{}{
					"id":       "10",
					"username": "admin",
					"online":   true,
				}
				mock.ExpectHSet("admin", value).SetVal(int64(1))
				mock.ExpectExpire("admin", ttl).SetVal(true)
			},
		},
		{
			name: "error",
			args: args{
				c: context.Background(),
				user: repository.User{
					Username: "admin",
					Id:       10,
				},
			},
			want: want{
				err: s.mockErr,
			},
			mockFunc: func() {
				value := map[string]interface{}{
					"id":       "10",
					"username": "admin",
					"online":   false,
				}
				mock.ExpectHSet("admin", value).SetErr(s.mockErr)
				mock.ExpectHSet("admin", value).SetVal(int64(1))

			},
		},
	}

	for _, tt := range testCases {
		s.Run(tt.name, func() {
			tt.mockFunc()
			err := cache.SetUser(tt.args.c, tt.args.user)
			s.Equal(tt.want.err, err)
		})
	}
}

func (s *RepositoryTestSuite) TestCacheGetUser() {

	type (
		args struct {
			c        context.Context
			username string
		}
		want struct {
			user repository.User
			err  error
		}
		testCase struct {
			name     string
			args     args
			want     want
			mockFunc func()
		}
	)

	db, mock := redismock.NewClientMock()
	cache := &repository.Cache{
		Redis: &redis.Redis{
			Client: db,
			Ttl:    300 * time.Second,
		},
	}
	testCases := []testCase{
		{
			name: "succes",
			args: args{
				c:        context.Background(),
				username: "admin",
			},
			want: want{
				user: repository.User{
					Username: "admin",
					Id:       10,
					Online:   true,
				},
			},
			mockFunc: func() {
				value := map[string]string{
					"id":       "10",
					"username": "admin",
					"online":   "true",
				}
				mock.ExpectHGetAll("admin").SetVal(value)
			},
		},
		{
			name: "error",
			args: args{
				c:        context.Background(),
				username: "admin",
			},
			want: want{
				err: s.mockErr,
			},
			mockFunc: func() {
				value := map[string]string{
					"id":       "10",
					"username": "admin",
				}
				mock.ExpectHGetAll("admin").SetErr(s.mockErr)
				mock.ExpectHGetAll("admin").SetVal(value)
			},
		},
	}

	for _, tt := range testCases {
		s.Run(tt.name, func() {
			tt.mockFunc()
			user, err := cache.GetUser(tt.args.c, tt.args.username)
			s.Equal(tt.want.user, user)
			s.Equal(tt.want.err, err)
		})
	}
}

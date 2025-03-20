package test

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/msyamsula/portofolio/domain/user/repository"
	"github.com/msyamsula/portofolio/domain/user/service"
	"github.com/stretchr/testify/assert"
)

func TestGetUser(t *testing.T) {

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCache := NewMockCacheLayer(ctrl)
	mockPersistence := NewMockPersistenceLayer(ctrl)

	svc := &service.Service{
		Persistence: mockPersistence,
		Cache:       mockCache,
	}

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

	mockErr := errors.New("ops")
	testCases := []testCase{
		{
			name: "cache hit",
			args: args{
				c:        context.Background(),
				username: "admin",
			},
			want: want{
				user: repository.User{
					Username: "admin",
					Id:       10,
				},
				err: nil,
			},
			mockFunc: func() {
				mockCache.EXPECT().
					GetUser(gomock.Any(), "admin").
					Return(repository.User{
						Username: "admin",
						Id:       10,
					}, nil)
			},
		},
		{
			name: "failed to get db",
			args: args{
				c:        context.Background(),
				username: "admin",
			},
			want: want{
				user: repository.User{},
				err:  mockErr,
			},
			mockFunc: func() {
				mockCache.EXPECT().
					GetUser(gomock.Any(), "admin").
					Return(repository.User{
						Username: "admin",
						Id:       -1,
					}, nil)

				mockPersistence.EXPECT().GetUser(gomock.Any(), "admin").Return(repository.User{}, mockErr)
			},
		},
		{
			name: "success",
			args: args{
				c:        context.Background(),
				username: "admin",
			},
			want: want{
				user: repository.User{
					Username: "admin",
					Id:       10,
				},
				err: nil,
			},
			mockFunc: func() {
				mockCache.EXPECT().
					GetUser(gomock.Any(), "admin").
					Return(repository.User{
						Username: "admin",
						Id:       -1,
					}, nil)

				mockPersistence.EXPECT().
					GetUser(gomock.Any(), "admin").
					Return(repository.User{
						Username: "admin",
						Id:       10,
					}, nil)

				mockCache.EXPECT().SetUser(gomock.Any(), gomock.Any())
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockFunc()
			user, err := svc.GetUser(tt.args.c, tt.args.username)
			assert.Equal(t, tt.want.err, err)
			assert.Equal(t, tt.want.user, user)
		})
	}
}

func TestSetUser(t *testing.T) {

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCache := NewMockCacheLayer(ctrl)
	mockPersistence := NewMockPersistenceLayer(ctrl)

	svc := &service.Service{
		Persistence: mockPersistence,
		Cache:       mockCache,
	}

	type (
		args struct {
			c    context.Context
			user repository.User
		}
		want struct {
			err  error
			user repository.User
		}
		testCase struct {
			name     string
			args     args
			want     want
			mockFunc func()
		}
	)

	mockErr := errors.New("ops")
	testCases := []testCase{
		{
			name: "persistence failed",
			args: args{
				c: context.Background(),
				user: repository.User{
					Username: "admin",
				},
			},
			want: want{
				user: repository.User{
					Username: "admin",
					Id:       10,
				},
				err: mockErr,
			},
			mockFunc: func() {
				mockPersistence.EXPECT().
					InsertUser(gomock.Any(), "admin").
					Return(repository.User{
						Username: "admin",
						Id:       10,
					}, mockErr)
			},
		},
		{
			name: "success",
			args: args{
				c: context.Background(),
				user: repository.User{
					Username: "admin",
					Id:       10,
				},
			},
			want: want{
				user: repository.User{
					Username: "admin",
					Id:       10,
				},
				err: nil,
			},
			mockFunc: func() {
				mockPersistence.EXPECT().
					InsertUser(gomock.Any(), "admin").
					Return(repository.User{
						Username: "admin",
						Id:       10,
					}, nil)

				mockCache.EXPECT().SetUser(gomock.Any(), gomock.Any())
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockFunc()
			user, err := svc.SetUser(tt.args.c, tt.args.user)
			assert.Equal(t, tt.want.err, err)
			assert.Equal(t, tt.want.user, user)
		})
	}
}

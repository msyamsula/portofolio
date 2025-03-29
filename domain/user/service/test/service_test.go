package test

import (
	"context"
	"errors"

	"github.com/golang/mock/gomock"
	"github.com/msyamsula/portofolio/domain/user/repository"
	"github.com/msyamsula/portofolio/domain/user/service"
)

func (s *ServiceTestSuite) TestGetUser() {

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
				s.mockCache.EXPECT().
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
				s.mockCache.EXPECT().
					GetUser(gomock.Any(), "admin").
					Return(repository.User{
						Username: "admin",
						Id:       -1,
					}, nil)

				s.mockPersistence.EXPECT().GetUser(gomock.Any(), "admin").Return(repository.User{}, mockErr)
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
				s.mockCache.EXPECT().
					GetUser(gomock.Any(), "admin").
					Return(repository.User{
						Username: "admin",
						Id:       -1,
					}, nil)

				s.mockPersistence.EXPECT().
					GetUser(gomock.Any(), "admin").
					Return(repository.User{
						Username: "admin",
						Id:       10,
					}, nil)

				s.mockCache.EXPECT().SetUser(gomock.Any(), gomock.Any())
			},
		},
	}

	svc := &service.Service{
		Persistence: s.mockPersistence,
		Cache:       s.mockCache,
	}
	for _, tt := range testCases {
		s.Run(tt.name, func() {
			tt.mockFunc()
			user, err := svc.GetUser(tt.args.c, tt.args.username)
			s.Equal(tt.want.err, err)
			s.Equal(tt.want.user, user)
		})
	}
}

func (s *ServiceTestSuite) TestSetUser() {

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

	svc := &service.Service{
		Persistence: s.mockPersistence,
		Cache:       s.mockCache,
	}
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
				err: s.mockErr,
			},
			mockFunc: func() {
				s.mockPersistence.EXPECT().
					InsertUser(gomock.Any(), repository.User{
						Username: "admin",
						Online:   false,
					}).
					Return(repository.User{
						Username: "admin",
						Id:       10,
					}, s.mockErr)
			},
		},
		{
			name: "success",
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
					Online:   true,
				},
				err: nil,
			},
			mockFunc: func() {
				s.mockPersistence.EXPECT().
					InsertUser(gomock.Any(), repository.User{
						Username: "admin",
					}).
					Return(repository.User{
						Username: "admin",
						Id:       10,
						Online:   true,
					}, nil)

				s.mockCache.EXPECT().SetUser(gomock.Any(), gomock.Any())
			},
		},
	}

	for _, tt := range testCases {
		s.Run(tt.name, func() {
			tt.mockFunc()
			user, err := svc.SetUser(tt.args.c, tt.args.user)
			s.Equal(tt.want.err, err)
			s.Equal(tt.want.user, user)
		})
	}
}

func (s *ServiceTestSuite) TestAddFriend() {

	type (
		args struct {
			c            context.Context
			userA, userB repository.User
		}
		want struct {
			err error
		}
		testCase struct {
			name     string
			args     args
			want     want
			mockFunc func()
		}
	)

	svc := &service.Service{
		Persistence: s.mockPersistence,
		Cache:       s.mockCache,
	}
	testCases := []testCase{
		{
			name: "service error",
			args: args{
				c: context.Background(),
				userA: repository.User{
					Username: "",
					Id:       1,
				},
				userB: repository.User{
					Username: "",
					Id:       2,
				},
			},
			want: want{
				err: s.mockErr,
			},
			mockFunc: func() {
				s.mockPersistence.EXPECT().
					AddFriend(gomock.Any(),
						repository.User{
							Username: "",
							Id:       1,
						},
						repository.User{
							Username: "",
							Id:       2,
						}).
					Return(s.mockErr)
			},
		},
		{
			name: "success",
			args: args{
				c: context.Background(),
				userA: repository.User{
					Username: "",
					Id:       1,
				},
				userB: repository.User{
					Username: "",
					Id:       2,
				},
			},
			want: want{
				err: nil,
			},
			mockFunc: func() {
				s.mockPersistence.EXPECT().
					AddFriend(gomock.Any(),
						repository.User{
							Username: "",
							Id:       1,
						},
						repository.User{
							Username: "",
							Id:       2,
						}).
					Return(nil)
			},
		},
	}

	for _, tt := range testCases {
		s.Run(tt.name, func() {
			tt.mockFunc()
			err := svc.AddFriend(tt.args.c, tt.args.userA, tt.args.userB)
			s.Equal(tt.want.err, err)
		})
	}
}

func (s *ServiceTestSuite) TestGetFriends() {

	type (
		args struct {
			c    context.Context
			user repository.User
		}
		want struct {
			err   error
			users []repository.User
		}
		testCase struct {
			name     string
			args     args
			want     want
			mockFunc func()
		}
	)

	svc := &service.Service{
		Persistence: s.mockPersistence,
		Cache:       s.mockCache,
	}

	testCases := []testCase{
		{
			name: "service error",
			args: args{
				c:    context.Background(),
				user: repository.User{},
			},
			want: want{
				err:   s.mockErr,
				users: []repository.User{},
			},
			mockFunc: func() {
				s.mockPersistence.EXPECT().
					GetFriends(gomock.Any(), gomock.Any()).
					Return([]repository.User{}, s.mockErr)
			},
		},
		{
			name: "success",
			args: args{
				c: context.Background(),
				user: repository.User{
					Username: "admin",
					Id:       1,
				},
			},
			want: want{
				err: nil,
				users: []repository.User{
					{
						Username: "2",
						Id:       2,
					},
					{
						Username: "3",
						Id:       3,
					},
				},
			},
			mockFunc: func() {
				s.mockPersistence.EXPECT().
					GetFriends(gomock.Any(), repository.User{
						Username: "admin",
						Id:       1,
					}).
					Return([]repository.User{
						{
							Username: "2",
							Id:       2,
						},
						{
							Username: "3",
							Id:       3,
						},
					}, nil)
			},
		},
	}

	for _, tt := range testCases {
		s.Run(tt.name, func() {
			tt.mockFunc()
			users, err := svc.GetFriends(tt.args.c, tt.args.user)
			s.Equal(tt.want.err, err)
			s.Equal(tt.want.users, users)
		})
	}
}

package test

import (
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/msyamsula/portofolio/domain/user/repository"
	"github.com/msyamsula/portofolio/domain/utils"
	"github.com/msyamsula/portofolio/tech-stack/postgres"
	"github.com/stretchr/testify/assert"
)

func (s *RepositoryTestSuite) TestInsertUser() {
	mockDb, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	s.Nil(err)
	defer mockDb.Close()

	sqlxDb := sqlx.NewDb(mockDb, "sqlmock")
	db := &postgres.Postgres{
		DB: sqlxDb,
	}
	persistence := &repository.Persistence{
		Postgres: db,
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

	testCases := []testCase{
		{
			name: "success",
			args: args{
				c: context.Background(),
				user: repository.User{
					Username: "admin",
					Id:       0,
					Online:   false,
				},
			},
			want: want{
				err: nil,
				user: repository.User{
					Username: "admin",
					Id:       1,
				},
			},
			mockFunc: func() {
				rows := sqlmock.NewRows([]string{"id"})
				rows.AddRow(1)

				mock.ExpectBegin()
				mock.ExpectPrepare(utils.CreatePrepareQuery(repository.QueryInsertUser)).
					ExpectQuery().
					WithArgs("admin", false).
					WillReturnRows(rows)
				mock.ExpectCommit()
			},
		},
		{
			name: "error in prepare context",
			args: args{
				c: context.Background(),
				user: repository.User{
					Username: "admin",
					Id:       0,
					Online:   false,
				},
			},
			want: want{
				err: s.mockErr,
			},
			mockFunc: func() {
				mock.ExpectBegin().WillReturnError(nil)
				mock.ExpectPrepare(utils.CreatePrepareQuery(repository.QueryInsertUser)).
					ExpectQuery().
					WithArgs("admin", false).
					WillReturnError(s.mockErr)
				mock.ExpectRollback()
			},
		},
		{
			name: "error in commit",
			args: args{
				c: context.Background(),
				user: repository.User{
					Username: "admin",
					Id:       0,
					Online:   false,
				},
			},
			want: want{
				err: s.mockErr,
				user: repository.User{
					Username: "admin",
					Id:       1,
				},
			},
			mockFunc: func() {
				rows := sqlmock.NewRows([]string{"id"})
				rows.AddRow(1)

				mock.ExpectBegin()
				mock.ExpectPrepare(utils.CreatePrepareQuery(repository.QueryInsertUser)).
					ExpectQuery().
					WithArgs("admin", false).
					WillReturnRows(rows)
				mock.ExpectCommit().WillReturnError(s.mockErr)
				mock.ExpectRollback()
			},
		},
	}

	for _, tt := range testCases {
		s.Run(tt.name, func() {
			tt.mockFunc()
			user, err := persistence.InsertUser(tt.args.c, tt.args.user)
			s.Equal(tt.want.err, err)
			s.Equal(tt.want.user, user)

		})
	}

}

func (s *RepositoryTestSuite) TestGetUser() {
	mockDb, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	s.Nil(err)
	defer mockDb.Close()

	sqlxDb := sqlx.NewDb(mockDb, "sqlmock")
	db := &postgres.Postgres{
		DB: sqlxDb,
	}
	persistence := &repository.Persistence{
		Postgres: db,
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

	testCases := []testCase{
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
					Online:   true,
				},
				err: nil,
			},
			mockFunc: func() {
				rows := sqlmock.NewRows([]string{"id", "username", "online"})
				rows.AddRow(10, "admin", true)

				mock.ExpectBegin()
				mock.ExpectPrepare(utils.CreatePrepareQuery(repository.QueryGetUser)).
					ExpectQuery().
					WithArgs("admin").
					WillReturnRows(rows)
				mock.ExpectCommit()
			},
		},
		{
			name: "user not found",
			args: args{
				c:        context.Background(),
				username: "admin",
			},
			want: want{
				user: repository.User{},
				err:  repository.ErrUserNotFound,
			},
			mockFunc: func() {
				rows := sqlmock.NewRows([]string{"id", "username", "online"})

				mock.ExpectBegin()
				mock.ExpectPrepare(utils.CreatePrepareQuery(repository.QueryGetUser)).
					ExpectQuery().
					WithArgs("admin").
					WillReturnRows(rows)
				mock.ExpectRollback()
			},
		},
		{
			name: "commit failed",
			args: args{
				c:        context.Background(),
				username: "admin",
			},
			want: want{
				user: repository.User{},
				err:  s.mockErr,
			},
			mockFunc: func() {
				rows := sqlmock.NewRows([]string{"id", "username", "online"})
				rows.AddRow(1, "admin", false)

				mock.ExpectBegin()
				mock.ExpectPrepare(utils.CreatePrepareQuery(repository.QueryGetUser)).
					ExpectQuery().
					WithArgs("admin").
					WillReturnRows(rows)
				mock.ExpectCommit().WillReturnError(s.mockErr)
				mock.ExpectRollback()
			},
		},
	}

	for _, tt := range testCases {
		s.Run(tt.name, func() {
			tt.mockFunc()

			user, err := persistence.GetUser(tt.args.c, tt.args.username)
			s.Equal(tt.want.user, user)
			s.Equal(tt.want.err, err)
		})
	}
}

func TestAddFriend(t *testing.T) {
	mockDb, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	assert.Nil(t, err)
	assert.NotNil(t, mock)
	defer mockDb.Close()

	sqlxDb := sqlx.NewDb(mockDb, "sqlmock")
	db := &postgres.Postgres{
		DB: sqlxDb,
	}
	persistence := &repository.Persistence{
		Postgres: db,
	}

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

	mockErr := errors.New("ops")
	testCases := []testCase{
		{
			name: "same id",
			args: args{
				c:     context.Background(),
				userA: repository.User{},
				userB: repository.User{},
			},
			want: want{
				err: repository.ErrIdMustBeDifferent,
			},
			mockFunc: func() {
				mock.ExpectBegin().WillReturnError(nil)
			},
		},
		{
			name: "error in prepare context",
			args: args{
				c: context.Background(),
				userA: repository.User{
					Username: "1",
					Id:       1,
				},
				userB: repository.User{
					Username: "2",
					Id:       2,
				},
			},
			want: want{
				err: mockErr,
			},
			mockFunc: func() {
				mock.ExpectBegin().WillReturnError(nil)
				mock.ExpectPrepare("INSERT INTO friendship (small_id, big_id) VALUES (?, ?) RETURNING id").
					WillReturnError(mockErr)

			},
		},
		{
			name: "error in query context",
			args: args{
				c: context.Background(),
				userA: repository.User{
					Username: "1",
					Id:       1,
				},
				userB: repository.User{
					Username: "2",
					Id:       2,
				},
			},
			want: want{
				err: mockErr,
			},
			mockFunc: func() {
				mock.ExpectBegin().WillReturnError(nil)
				mock.ExpectPrepare("INSERT INTO friendship (small_id, big_id) VALUES (?, ?) RETURNING id").
					ExpectQuery().
					WithArgs(1, 2).
					WillReturnError(mockErr)

			},
		},
		{
			name: "error in commit",
			args: args{
				c: context.Background(),
				userA: repository.User{
					Username: "1",
					Id:       1,
				},
				userB: repository.User{
					Username: "2",
					Id:       2,
				},
			},
			want: want{
				err: mockErr,
			},
			mockFunc: func() {
				rows := sqlmock.NewRows([]string{"id"})
				rows.AddRow(int64(11))

				mock.ExpectBegin().WillReturnError(nil)
				mock.ExpectPrepare("INSERT INTO friendship (small_id, big_id) VALUES (?, ?) RETURNING id").
					ExpectQuery().
					WithArgs(int64(1), int64(2)).
					WillReturnRows(rows).
					WillReturnError(nil)

				mock.ExpectCommit().WillReturnError(mockErr)
			},
		},
		{
			name: "success",
			args: args{
				c: context.Background(),
				userA: repository.User{
					Username: "1",
					Id:       1,
				},
				userB: repository.User{
					Username: "2",
					Id:       2,
				},
			},
			want: want{
				err: nil,
			},
			mockFunc: func() {
				rows := sqlmock.NewRows([]string{"id"})
				rows.AddRow(int64(11))

				mock.ExpectBegin().WillReturnError(nil)
				mock.ExpectPrepare("INSERT INTO friendship (small_id, big_id) VALUES (?, ?) RETURNING id").
					ExpectQuery().
					WithArgs(int64(1), int64(2)).
					WillReturnRows(rows).
					WillReturnError(nil)

				mock.ExpectCommit().WillReturnError(nil)
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockFunc()
			err := persistence.AddFriend(tt.args.c, tt.args.userA, tt.args.userB)
			assert.Equal(t, tt.want.err, err)
		})
	}
}

func (s *RepositoryTestSuite) TestGetFriends() {
	mockDb, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	s.Nil(err)
	defer mockDb.Close()

	sqlxDb := sqlx.NewDb(mockDb, "sqlmock")
	db := &postgres.Postgres{
		DB: sqlxDb,
	}
	persistence := &repository.Persistence{
		Postgres: db,
	}

	type (
		args struct {
			c    context.Context
			user repository.User
		}
		want struct {
			users []repository.User
			err   error
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
			name: "error in prepare",
			args: args{
				c:    context.Background(),
				user: repository.User{},
			},
			want: want{
				users: []repository.User{},
				err:   s.mockErr,
			},
			mockFunc: func() {
				mock.ExpectBegin().WillReturnError(nil)
				mock.ExpectPrepare(
					utils.CreatePrepareQuery(repository.QueryGetFriends)).
					WillReturnError(s.mockErr)
			},
		},
		{
			name: "error in query",
			args: args{
				c:    context.Background(),
				user: repository.User{},
			},
			want: want{
				users: []repository.User{},
				err:   s.mockErr,
			},
			mockFunc: func() {
				mock.ExpectBegin().WillReturnError(nil)
				mock.ExpectPrepare(
					utils.CreatePrepareQuery(repository.QueryGetFriends)).
					ExpectQuery().
					WithArgs(0, 0, 0).
					WillReturnError(s.mockErr)
			},
		},
		{
			name: "error in commit",
			args: args{
				c:    context.Background(),
				user: repository.User{},
			},
			want: want{
				users: []repository.User{},
				err:   s.mockErr,
			},
			mockFunc: func() {
				rows := sqlmock.NewRows([]string{"id", "username"})
				rows.AddRow(1, "admin")

				mock.ExpectBegin().WillReturnError(nil)
				mock.ExpectPrepare(
					utils.CreatePrepareQuery(repository.QueryGetFriends)).
					ExpectQuery().
					WithArgs(0, 0, 0).
					WillReturnRows(rows)

				mock.ExpectCommit().WillReturnError(s.mockErr)

			},
		},
		{
			name: "success",
			args: args{
				c: context.Background(),
				user: repository.User{
					Id: 10,
				},
			},
			want: want{
				users: []repository.User{
					{
						Username: "admin",
						Id:       1,
						Online:   true,
						Unread:   10,
					},
					{
						Username: "testing",
						Id:       19,
						Online:   false,
						Unread:   7,
					},
				},
				err: nil,
			},
			mockFunc: func() {
				rows := sqlmock.NewRows([]string{"id", "username", "online", "unread"})
				rows.AddRow(1, "admin", true, 10)
				rows.AddRow(19, "testing", false, 7)

				mock.ExpectBegin().WillReturnError(nil)
				mock.ExpectPrepare(
					utils.CreatePrepareQuery(repository.QueryGetFriends)).
					ExpectQuery().
					WithArgs(10, 10, 10).
					WillReturnRows(rows)

				mock.ExpectCommit().WillReturnError(nil)

			},
		},
	}
	for _, tt := range testCases {
		s.Run(tt.name, func() {
			tt.mockFunc()
			users, err := persistence.GetFriends(tt.args.c, tt.args.user)
			s.Equal(tt.want.err, err)
			s.Equal(tt.want.users, users)
		})
	}
}

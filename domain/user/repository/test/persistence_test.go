package test

import (
	"context"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/msyamsula/portofolio/domain/user/repository"
	"github.com/msyamsula/portofolio/tech-stack/postgres"
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
			c        context.Context
			username string
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
				c:        context.Background(),
				username: "admin",
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
				mock.ExpectPrepare("INSERT INTO users (username) VALUES (?) RETURNING id").
					ExpectQuery().
					WithArgs("admin").
					WillReturnRows(rows)
				mock.ExpectCommit()
			},
		},
		{
			name: "error in prepare context",
			args: args{
				c:        context.Background(),
				username: "admin",
			},
			want: want{
				err: s.mockErr,
			},
			mockFunc: func() {
				mock.ExpectBegin().WillReturnError(nil)
				mock.ExpectPrepare("INSERT INTO users (username) VALUES (?) RETURNING id").
					ExpectQuery().
					WithArgs("admin").
					WillReturnError(s.mockErr)
				mock.ExpectRollback()
			},
		},
		{
			name: "error in commit",
			args: args{
				c:        context.Background(),
				username: "admin",
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
				mock.ExpectPrepare("INSERT INTO users (username) VALUES (?) RETURNING id").
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
			user, err := persistence.InsertUser(tt.args.c, tt.args.username)
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
				},
				err: nil,
			},
			mockFunc: func() {
				rows := sqlmock.NewRows([]string{"id", "username"})
				rows.AddRow(10, "admin")

				mock.ExpectBegin()
				mock.ExpectPrepare("SELECT id, username FROM users WHERE username = ?").
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
				rows := sqlmock.NewRows([]string{"id", "username"})

				mock.ExpectBegin()
				mock.ExpectPrepare("SELECT id, username FROM users WHERE username = ?").
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
				rows := sqlmock.NewRows([]string{"id", "username"})
				rows.AddRow(1, "admin")

				mock.ExpectBegin()
				mock.ExpectPrepare("SELECT id, username FROM users WHERE username = ?").
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

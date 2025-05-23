package test

import (
	"context"
	"fmt"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/msyamsula/portofolio/binary/postgres"
	"github.com/msyamsula/portofolio/domain/message/repository"
	"github.com/msyamsula/portofolio/domain/utils"
)

func (s *RepositoryTestSuite) TestAddMessage() {

	type (
		args struct {
			c   context.Context
			msg repository.Message
		}
		want struct {
			msg repository.Message
			err error
		}
		testCase struct {
			name     string
			args     args
			want     want
			mockFunc func()
		}
	)

	persistence := &repository.Persistence{
		Postgres: &postgres.Postgres{
			DB: s.sqlxDb,
		},
	}
	testCases := []testCase{
		{
			name: "prepare error",
			args: args{
				c:   context.Background(),
				msg: repository.Message{},
			},
			want: want{
				msg: repository.Message{},
				err: s.mockErr,
			},
			mockFunc: func() {
				s.mock.ExpectBegin().WillReturnError(nil)
				s.mock.
					ExpectPrepare(
						utils.CreatePrepareQuery(repository.QueryInsertMessage),
					).WillReturnError(s.mockErr)
			},
		},
		{
			name: "query error",
			args: args{
				c: context.Background(),
				msg: repository.Message{
					SenderId:   1,
					ReceiverId: 2,
					Text:       "mantap",
				},
			},
			want: want{
				msg: repository.Message{},
				err: s.mockErr,
			},
			mockFunc: func() {
				s.mock.ExpectBegin().WillReturnError(nil)
				s.mock.
					ExpectPrepare(
						utils.CreatePrepareQuery(repository.QueryInsertMessage),
					).
					ExpectQuery().
					WithArgs(int64(1), int64(2), "mantap").
					WillReturnError(s.mockErr)
			},
		},
		{
			name: "commit error",
			args: args{
				c: context.Background(),
				msg: repository.Message{
					SenderId:   1,
					ReceiverId: 2,
					Text:       "mantap",
				},
			},
			want: want{
				msg: repository.Message{},
				err: s.mockErr,
			},
			mockFunc: func() {
				rows := sqlmock.NewRows([]string{"id"})
				rows.AddRow(4)

				s.mock.ExpectBegin().WillReturnError(nil)
				s.mock.
					ExpectPrepare(
						utils.CreatePrepareQuery(repository.QueryInsertMessage),
					).
					ExpectQuery().
					WithArgs(int64(1), int64(2), "mantap").
					WillReturnRows(rows)
				s.mock.ExpectCommit().WillReturnError(s.mockErr)
			},
		},
		{
			name: "success",
			args: args{
				c: context.Background(),
				msg: repository.Message{
					SenderId:   1,
					ReceiverId: 2,
					Text:       "mantap",
				},
			},
			want: want{
				msg: repository.Message{
					Id:         4,
					SenderId:   1,
					ReceiverId: 2,
					Text:       "mantap",
				},
				err: nil,
			},
			mockFunc: func() {
				rows := sqlmock.NewRows([]string{"id"})
				rows.AddRow(4)

				s.mock.ExpectBegin().WillReturnError(nil)
				s.mock.
					ExpectPrepare(
						utils.CreatePrepareQuery(repository.QueryInsertMessage),
					).
					ExpectQuery().
					WithArgs(int64(1), int64(2), "mantap").
					WillReturnRows(rows)
				s.mock.ExpectCommit().WillReturnError(nil)
			},
		},
	}

	for _, tt := range testCases {
		s.Run(tt.name, func() {
			tt.mockFunc()
			msg, err := persistence.AddMessage(tt.args.c, tt.args.msg)
			s.Equal(tt.want.err, err)
			s.Equal(tt.want.msg, msg)
		})
	}
}

func (s *RepositoryTestSuite) TestGetConversation() {

	type (
		args struct {
			c                    context.Context
			senderId, receiverId int64
		}
		want struct {
			msgs []repository.Message
			err  error
		}
		testCase struct {
			name     string
			args     args
			want     want
			mockFunc func()
		}
	)

	persistence := &repository.Persistence{
		Postgres: &postgres.Postgres{
			DB: s.sqlxDb,
		},
	}

	timeA := time.Now()
	timeB := timeA.Add(1 * time.Hour)
	timeC := timeB.Add(1 * time.Hour)
	testCases := []testCase{
		{
			name: "prepare error",
			args: args{
				c:          context.Background(),
				senderId:   1,
				receiverId: 2,
			},
			want: want{
				msgs: []repository.Message{},
				err:  s.mockErr,
			},
			mockFunc: func() {
				s.mock.ExpectBegin().WillReturnError(nil)
				s.mock.
					ExpectPrepare(
						utils.CreatePrepareQuery(repository.QueryGetConversation),
					).
					WillReturnError(s.mockErr)

			},
		},
		{
			name: "query error",
			args: args{
				c:          context.Background(),
				senderId:   1,
				receiverId: 2,
			},
			want: want{
				msgs: []repository.Message{},
				err:  s.mockErr,
			},
			mockFunc: func() {
				s.mock.ExpectBegin().WillReturnError(nil)
				s.mock.
					ExpectPrepare(
						utils.CreatePrepareQuery(repository.QueryGetConversation),
					).
					ExpectQuery().
					WithArgs(int64(1), int64(2)).
					WillReturnError(s.mockErr)

			},
		},
		{
			name: "commit error",
			args: args{
				c:          context.Background(),
				senderId:   1,
				receiverId: 2,
			},
			want: want{
				msgs: []repository.Message{},
				err:  s.mockErr,
			},
			mockFunc: func() {
				rows := sqlmock.NewRows([]string{"id", "sender_id", "receiver_id", "text", "create_time", "is_read"})
				rows.AddRow(3, 1, 2, "three", time.Now().Add(3*time.Hour), false)
				rows.AddRow(1, 1, 2, "one", time.Now().Add(1*time.Hour), false)
				rows.AddRow(2, 1, 2, "two", time.Now().Add(2*time.Hour), false)

				s.mock.ExpectBegin().WillReturnError(nil)
				s.mock.
					ExpectPrepare(
						utils.CreatePrepareQuery(repository.QueryGetConversation),
					).
					ExpectQuery().
					WithArgs(int64(1), int64(2)).
					WillReturnRows(rows)
				s.mock.ExpectCommit().WillReturnError(s.mockErr)

			},
		},
		{
			name: "success",
			args: args{
				c:          context.Background(),
				senderId:   1,
				receiverId: 2,
			},
			want: want{
				msgs: []repository.Message{
					{
						Id:         3,
						SenderId:   1,
						ReceiverId: 2,
						Text:       "three",
						CreateTime: timeC,
						IsRead:     false,
					},
					{
						Id:         1,
						SenderId:   1,
						ReceiverId: 2,
						Text:       "one",
						CreateTime: timeA,
						IsRead:     false,
					},
					{
						Id:         2,
						SenderId:   1,
						ReceiverId: 2,
						Text:       "two",
						CreateTime: timeB,
						IsRead:     true,
					},
				},
				err: nil,
			},
			mockFunc: func() {
				rows := sqlmock.NewRows([]string{"id", "sender_id", "receiver_id", "text", "create_time", "is_read"})
				rows.AddRow(3, 1, 2, "three", timeC, false)
				rows.AddRow(1, 1, 2, "one", timeA, false)
				rows.AddRow(2, 1, 2, "two", timeB, true)

				s.mock.ExpectBegin().WillReturnError(nil)
				s.mock.
					ExpectPrepare(
						utils.CreatePrepareQuery(repository.QueryGetConversation),
					).
					ExpectQuery().
					WithArgs(int64(1), int64(2)).
					WillReturnRows(rows)
				s.mock.ExpectCommit().WillReturnError(nil)

			},
		},
	}

	for _, tt := range testCases {
		s.Run(tt.name, func() {
			tt.mockFunc()
			msgs, err := persistence.GetConversation(tt.args.c, tt.args.senderId, tt.args.receiverId)
			s.Equal(tt.want.err, err)
			s.Equal(tt.want.msgs, msgs)
		})
	}
}

func (s *RepositoryTestSuite) TestReadMessage() {

	type (
		args struct {
			c                    context.Context
			senderId, receiverId int64
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

	persistence := &repository.Persistence{
		Postgres: &postgres.Postgres{
			DB: s.sqlxDb,
		},
	}

	timeA := time.Now()
	timeB := timeA.Add(1 * time.Hour)
	timeC := timeB.Add(1 * time.Hour)
	fmt.Println(timeC)
	testCases := []testCase{
		{
			name: "prepare error",
			args: args{
				c:          context.Background(),
				senderId:   1,
				receiverId: 2,
			},
			want: want{
				err: s.mockErr,
			},
			mockFunc: func() {
				s.mock.ExpectBegin().WillReturnError(nil)
				s.mock.
					ExpectPrepare(
						utils.CreatePrepareQuery(repository.QueryReadMessage),
					).
					WillReturnError(s.mockErr)

			},
		},
		{
			name: "query error",
			args: args{
				c:          context.Background(),
				senderId:   1,
				receiverId: 2,
			},
			want: want{
				err: s.mockErr,
			},
			mockFunc: func() {
				s.mock.ExpectBegin().WillReturnError(nil)
				s.mock.
					ExpectPrepare(
						utils.CreatePrepareQuery(repository.QueryReadMessage),
					).
					ExpectQuery().
					WithArgs(int64(1), int64(2)).
					WillReturnError(s.mockErr)

			},
		},
		{
			name: "commit error",
			args: args{
				c:          context.Background(),
				senderId:   1,
				receiverId: 2,
			},
			want: want{
				err: s.mockErr,
			},
			mockFunc: func() {
				rows := sqlmock.NewRows([]string{"id", "sender_id", "receiver_id", "text", "create_time", "is_read"})

				s.mock.ExpectBegin().WillReturnError(nil)
				s.mock.
					ExpectPrepare(
						utils.CreatePrepareQuery(repository.QueryReadMessage),
					).
					ExpectQuery().
					WithArgs(int64(1), int64(2)).
					WillReturnRows(rows)
				s.mock.ExpectCommit().WillReturnError(s.mockErr)

			},
		},
		{
			name: "success",
			args: args{
				c:          context.Background(),
				senderId:   1,
				receiverId: 2,
			},
			want: want{
				err: nil,
			},
			mockFunc: func() {
				rows := sqlmock.NewRows([]string{"id", "sender_id", "receiver_id", "text", "create_time", "is_read"})

				s.mock.ExpectBegin().WillReturnError(nil)
				s.mock.
					ExpectPrepare(
						utils.CreatePrepareQuery(repository.QueryReadMessage),
					).
					ExpectQuery().
					WithArgs(int64(1), int64(2)).
					WillReturnRows(rows)
				s.mock.ExpectCommit().WillReturnError(nil)

			},
		},
	}

	for _, tt := range testCases {
		s.Run(tt.name, func() {
			tt.mockFunc()
			err := persistence.ReadMessage(tt.args.c, tt.args.senderId, tt.args.receiverId)
			s.Equal(tt.want.err, err)
		})
	}
}

func (s *RepositoryTestSuite) TestUpdateUnread() {

	type (
		args struct {
			c                            context.Context
			senderId, receiverId, unread int64
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

	persistence := &repository.Persistence{
		Postgres: &postgres.Postgres{
			DB: s.sqlxDb,
		},
	}

	timeA := time.Now()
	timeB := timeA.Add(1 * time.Hour)
	timeC := timeB.Add(1 * time.Hour)
	fmt.Println(timeC)
	testCases := []testCase{
		{
			name: "prepare error",
			args: args{
				c:          context.Background(),
				senderId:   1,
				receiverId: 2,
				unread:     10,
			},
			want: want{
				err: s.mockErr,
			},
			mockFunc: func() {
				s.mock.ExpectBegin().WillReturnError(nil)
				s.mock.
					ExpectPrepare(
						utils.CreatePrepareQuery(repository.QueryUpdateUnread),
					).
					WillReturnError(s.mockErr)

			},
		},
		{
			name: "query error",
			args: args{
				c:          context.Background(),
				senderId:   1,
				receiverId: 2,
				unread:     10,
			},
			want: want{
				err: s.mockErr,
			},
			mockFunc: func() {
				s.mock.ExpectBegin().WillReturnError(nil)
				s.mock.
					ExpectPrepare(
						utils.CreatePrepareQuery(repository.QueryUpdateUnread),
					).
					ExpectQuery().
					WithArgs(int64(1), int64(2), int64(10)).
					WillReturnError(s.mockErr)

			},
		},
		{
			name: "commit error",
			args: args{
				c:          context.Background(),
				senderId:   1,
				receiverId: 2,
				unread:     10,
			},
			want: want{
				err: s.mockErr,
			},
			mockFunc: func() {
				rows := sqlmock.NewRows([]string{"id"})
				rows.AddRow(1)

				s.mock.ExpectBegin().WillReturnError(nil)
				s.mock.
					ExpectPrepare(
						utils.CreatePrepareQuery(repository.QueryUpdateUnread),
					).
					ExpectQuery().
					WithArgs(int64(1), int64(2), int64(10)).
					WillReturnRows(rows)
				s.mock.ExpectCommit().WillReturnError(s.mockErr)

			},
		},
		{
			name: "success",
			args: args{
				c:          context.Background(),
				senderId:   1,
				receiverId: 2,
				unread:     10,
			},
			want: want{
				err: nil,
			},
			mockFunc: func() {
				rows := sqlmock.NewRows([]string{"id"})
				rows.AddRow(1)

				s.mock.ExpectBegin().WillReturnError(nil)
				s.mock.
					ExpectPrepare(
						utils.CreatePrepareQuery(repository.QueryUpdateUnread),
					).
					ExpectQuery().
					WithArgs(int64(1), int64(2), int64(10)).
					WillReturnRows(rows)
				s.mock.ExpectCommit().WillReturnError(nil)

			},
		},
	}

	for _, tt := range testCases {
		s.Run(tt.name, func() {
			tt.mockFunc()
			err := persistence.UpdateUnread(tt.args.c, tt.args.senderId, tt.args.receiverId, tt.args.unread)
			s.Equal(tt.want.err, err)
		})
	}
}

package test

import (
	context "context"
	"time"

	"github.com/golang/mock/gomock"
	repository "github.com/msyamsula/portofolio/domain/message/repository"
	"github.com/msyamsula/portofolio/domain/message/service"
)

func (s *ServiceTestSuite) TestAddMessage() {
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

	svc := service.Service{
		Persistence: s.mockPersistence,
	}
	testCases := []testCase{
		{
			name: "invalid sender id",
			args: args{
				c: context.Background(),
				msg: repository.Message{
					SenderId: 0,
				},
			},
			want: want{
				msg: repository.Message{},
				err: service.ErrBadRequest,
			},
			mockFunc: func() {
			},
		},
		{
			name: "invalid receiver id",
			args: args{
				c: context.Background(),
				msg: repository.Message{
					ReceiverId: 0,
				},
			},
			want: want{
				msg: repository.Message{},
				err: service.ErrBadRequest,
			},
			mockFunc: func() {
			},
		},
		{
			name: "same id",
			args: args{
				c: context.Background(),
				msg: repository.Message{
					ReceiverId: 12,
					SenderId:   12,
				},
			},
			want: want{
				msg: repository.Message{},
				err: service.ErrBadRequest,
			},
			mockFunc: func() {
			},
		},
		{
			name: "empty text",
			args: args{
				c: context.Background(),
				msg: repository.Message{
					SenderId:   12,
					ReceiverId: 13,
					Text:       "",
				},
			},
			want: want{
				msg: repository.Message{},
				err: service.ErrBadRequest,
			},
			mockFunc: func() {
			},
		},
		{
			name: "db error",
			args: args{
				c: context.Background(),
				msg: repository.Message{
					SenderId:   12,
					ReceiverId: 13,
					Text:       "mantap",
				},
			},
			want: want{
				msg: repository.Message{},
				err: s.mockErr,
			},
			mockFunc: func() {
				s.mockPersistence.EXPECT().
					AddMessage(gomock.Any(), repository.Message{
						SenderId:   12,
						ReceiverId: 13,
						Text:       "mantap",
					}).
					Return(repository.Message{}, s.mockErr)
			},
		},
		{
			name: "success",
			args: args{
				c: context.Background(),
				msg: repository.Message{
					SenderId:   12,
					ReceiverId: 13,
					Text:       "mantap",
				},
			},
			want: want{
				msg: repository.Message{
					Id:         10,
					SenderId:   12,
					ReceiverId: 13,
					Text:       "mantap",
				},
				err: nil,
			},
			mockFunc: func() {
				s.mockPersistence.EXPECT().
					AddMessage(gomock.Any(), repository.Message{
						SenderId:   12,
						ReceiverId: 13,
						Text:       "mantap",
					}).
					Return(repository.Message{
						Id:         10,
						SenderId:   12,
						ReceiverId: 13,
						Text:       "mantap",
					}, nil)
			},
		},
	}

	for _, tt := range testCases {
		s.Run(tt.name, func() {
			tt.mockFunc()
			msg, err := svc.AddMessage(tt.args.c, tt.args.msg)
			s.Equal(tt.want.err, err)
			s.Equal(tt.want.msg, msg)
		})
	}

}

func (s *ServiceTestSuite) TestGetConversation() {
	type (
		args struct {
			c        context.Context
			idA, idB int64
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

	svc := service.Service{
		Persistence: s.mockPersistence,
	}
	timeA := time.Now()
	timeB := timeA.Add(1 * time.Hour)
	timeC := timeB.Add(1 * time.Hour)
	testCases := []testCase{
		{
			name: "same id",
			args: args{
				c:   context.Background(),
				idA: 12,
				idB: 12,
			},
			want: want{
				msgs: []repository.Message{},
				err:  service.ErrBadRequest,
			},
			mockFunc: func() {
			},
		},
		{
			name: "first half failed",
			args: args{
				c:   context.Background(),
				idA: 11,
				idB: 12,
			},
			want: want{
				msgs: []repository.Message{},
				err:  s.mockErr,
			},
			mockFunc: func() {
				s.mockPersistence.EXPECT().
					GetConversation(gomock.Any(), int64(11), int64(12)).
					Return([]repository.Message{}, s.mockErr)
			},
		},
		{
			name: "second half failed",
			args: args{
				c:   context.Background(),
				idA: 11,
				idB: 12,
			},
			want: want{
				msgs: []repository.Message{},
				err:  s.mockErr,
			},
			mockFunc: func() {
				s.mockPersistence.EXPECT().
					GetConversation(gomock.Any(), int64(11), int64(12)).
					Return([]repository.Message{
						{
							Id:         3,
							SenderId:   11,
							ReceiverId: 12,
							Text:       "three",
							CreateTime: timeC,
						},
					}, nil)

				s.mockPersistence.EXPECT().
					GetConversation(gomock.Any(), int64(12), int64(11)).
					Return([]repository.Message{
						{
							Id:         1,
							SenderId:   12,
							ReceiverId: 11,
							Text:       "one",
							CreateTime: timeA,
						},
						{
							Id:         2,
							SenderId:   12,
							ReceiverId: 11,
							Text:       "two",
							CreateTime: timeB,
						},
					}, s.mockErr)
			},
		},
		{
			name: "success",
			args: args{
				c:   context.Background(),
				idA: 11,
				idB: 12,
			},
			want: want{
				msgs: []repository.Message{
					{
						Id:         1,
						SenderId:   12,
						ReceiverId: 11,
						Text:       "one",
						CreateTime: timeA,
						IsRead:     true,
					},
					{
						Id:         2,
						SenderId:   12,
						ReceiverId: 11,
						Text:       "two",
						CreateTime: timeB,
						IsRead:     false,
					},
					{
						Id:         3,
						SenderId:   11,
						ReceiverId: 12,
						Text:       "three",
						CreateTime: timeC,
						IsRead:     false,
					},
				},
				err: nil,
			},
			mockFunc: func() {
				s.mockPersistence.EXPECT().
					GetConversation(gomock.Any(), int64(11), int64(12)).
					Return([]repository.Message{
						{
							Id:         3,
							SenderId:   11,
							ReceiverId: 12,
							Text:       "three",
							CreateTime: timeC,
							IsRead:     false,
						},
					}, nil)

				s.mockPersistence.EXPECT().
					GetConversation(gomock.Any(), int64(12), int64(11)).
					Return([]repository.Message{
						{
							Id:         2,
							SenderId:   12,
							ReceiverId: 11,
							Text:       "two",
							CreateTime: timeB,
							IsRead:     false,
						},
						{
							Id:         1,
							SenderId:   12,
							ReceiverId: 11,
							Text:       "one",
							CreateTime: timeA,
							IsRead:     true,
						},
					}, nil)
			},
		},
	}

	for _, tt := range testCases {
		s.Run(tt.name, func() {
			tt.mockFunc()
			msgs, err := svc.GetConversation(tt.args.c, tt.args.idA, tt.args.idB)
			s.Equal(tt.want.err, err)
			s.Equal(tt.want.msgs, msgs)
		})
	}
}

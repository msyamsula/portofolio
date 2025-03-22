package test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"

	"github.com/golang/mock/gomock"
	messagehttp "github.com/msyamsula/portofolio/domain/message/http"
	"github.com/msyamsula/portofolio/domain/message/repository"
	"github.com/msyamsula/portofolio/domain/message/service"
)

func (s *HandlerTestSuite) TestAddMessage() {

	type (
		args struct {
			c      context.Context
			msg    repository.Message
			method string
		}
		want struct {
			data repository.Message
			err  string
		}
		testCase struct {
			name     string
			args     args
			want     want
			mockFunc func()
		}
	)

	// offsetSeconds := (7 * 3600) + (7 * 60)
	// Create a fixed timezone with the offset
	// loc := time.FixedZone("+0707", offsetSeconds)

	// timeA := time.Date(1, 1, 1, 0, 0, 0, 0, time.Local)
	testCases := []testCase{
		{
			name: "invalid method",
			args: args{
				c:      context.Background(),
				msg:    repository.Message{},
				method: http.MethodDelete,
			},
			want: want{
				data: repository.Message{},
				err:  "",
			},
			mockFunc: func() {
			},
		},
		{
			name: "service error",
			args: args{
				c: context.Background(),
				msg: repository.Message{
					SenderId:   1,
					ReceiverId: 2,
					Text:       "mantap",
					// CreateTime: timeA,
				},
				method: http.MethodPost,
			},
			want: want{
				data: repository.Message{},
				err:  s.mockErr.Error(),
			},
			mockFunc: func() {
				s.mockService.EXPECT().
					AddMessage(
						gomock.Any(),
						repository.Message{
							SenderId:   1,
							ReceiverId: 2,
							Text:       "mantap",
						},
					).
					Return(repository.Message{}, s.mockErr)
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
					// CreateTime: timeA,
				},
				method: http.MethodPost,
			},
			want: want{
				data: repository.Message{
					Id:         10,
					SenderId:   1,
					ReceiverId: 2,
					Text:       "mantap",
					// CreateTime: timeA,
				},
				err: "",
			},
			mockFunc: func() {
				s.mockService.EXPECT().
					AddMessage(
						gomock.Any(),
						repository.Message{
							SenderId:   1,
							ReceiverId: 2,
							Text:       "mantap",
						},
					).
					Return(repository.Message{
						Id:         10,
						SenderId:   1,
						ReceiverId: 2,
						Text:       "mantap",
						// CreateTime: timeA,
					}, nil)
			},
		},
	}

	h := messagehttp.Handler{
		Service: s.mockService,
	}
	for _, tt := range testCases {
		s.Run(tt.name, func() {
			server := httptest.NewServer(http.HandlerFunc(h.ManageMesage))
			defer server.Close()

			bBody, err := json.Marshal(tt.args.msg)
			s.Nil(err)
			body := bytes.NewBuffer(bBody)
			request, err := http.NewRequest(tt.args.method, server.URL, body)
			s.Nil(err)

			tt.mockFunc()

			c := &http.Client{}
			resp, err := c.Do(request)
			s.Nil(err)
			defer resp.Body.Close()

			type r struct {
				Error   string             `json:"error,omitempty"`
				Message string             `json:"message,omitempty"`
				Data    repository.Message `json:"data,omitempty"`
			}
			httpResponse := r{}
			bResp, _ := io.ReadAll(resp.Body)
			json.Unmarshal(bResp, &httpResponse)

			s.Equal(tt.want.err, httpResponse.Error)
			s.Equal(tt.want.data, httpResponse.Data)
		})
	}
}

func (s *HandlerTestSuite) TestGetconversation() {

	type (
		args struct {
			c              context.Context
			pairId, userId string
			method         string
		}
		want struct {
			data []repository.Message
			err  string
		}
		testCase struct {
			name     string
			args     args
			want     want
			mockFunc func()
		}
	)

	// offsetSeconds := (7 * 3600) + (7 * 60)
	// Create a fixed timezone with the offset
	// loc := time.FixedZone("+0707", offsetSeconds)

	// timeA := time.Date(1, 1, 1, 0, 0, 0, 0, time.Local)
	testCases := []testCase{
		{
			name: "invalid method",
			args: args{
				c:      context.Background(),
				method: http.MethodDelete,
			},
			want: want{
				data: nil,
				err:  "",
			},
			mockFunc: func() {
			},
		},
		{
			name: "userId bad",
			args: args{
				c:      context.Background(),
				pairId: "12",
				userId: "abc",
				method: http.MethodGet,
			},
			want: want{
				data: nil,
				err:  service.ErrBadRequest.Error(),
			},
			mockFunc: func() {
			},
		},
		{
			name: "pairId bad",
			args: args{
				c:      context.Background(),
				pairId: "abc",
				userId: "1",
				method: http.MethodGet,
			},
			want: want{
				data: nil,
				err:  service.ErrBadRequest.Error(),
			},
			mockFunc: func() {
			},
		},
		{
			name: "service error",
			args: args{
				c:      context.Background(),
				pairId: "2",
				userId: "1",
				method: http.MethodGet,
			},
			want: want{
				data: nil,
				err:  s.mockErr.Error(),
			},
			mockFunc: func() {
				s.mockService.EXPECT().
					GetConversation(gomock.Any(), int64(1), int64(2)).
					Return([]repository.Message{}, s.mockErr)
			},
		},
		{
			name: "success",
			args: args{
				c:      context.Background(),
				pairId: "2",
				userId: "1",
				method: http.MethodGet,
			},
			want: want{
				data: []repository.Message{
					{
						SenderId:   1,
						ReceiverId: 2,
						Text:       "one",
					},
					{
						SenderId:   2,
						ReceiverId: 1,
						Text:       "two",
					},
				},
				err: "",
			},
			mockFunc: func() {
				s.mockService.EXPECT().
					GetConversation(gomock.Any(), int64(1), int64(2)).
					Return([]repository.Message{
						{
							SenderId:   1,
							ReceiverId: 2,
							Text:       "one",
						},
						{
							SenderId:   2,
							ReceiverId: 1,
							Text:       "two",
						},
					}, nil)
			},
		},
	}

	h := messagehttp.Handler{
		Service: s.mockService,
	}
	for _, tt := range testCases {
		s.Run(tt.name, func() {
			server := httptest.NewServer(http.HandlerFunc(h.ManageMesage))
			defer server.Close()

			request, err := http.NewRequest(tt.args.method, server.URL, nil)
			s.Nil(err)
			query := request.URL.Query()
			query.Set("userId", tt.args.userId)
			query.Set("pairId", tt.args.pairId)
			request.URL.RawQuery = query.Encode()

			tt.mockFunc()

			c := &http.Client{}
			resp, err := c.Do(request)
			s.Nil(err)
			defer resp.Body.Close()

			type r struct {
				Error   string               `json:"error,omitempty"`
				Message string               `json:"message,omitempty"`
				Data    []repository.Message `json:"data,omitempty"`
			}
			httpResponse := r{}
			bResp, _ := io.ReadAll(resp.Body)
			json.Unmarshal(bResp, &httpResponse)

			s.Equal(tt.want.err, httpResponse.Error)
			s.Equal(tt.want.data, httpResponse.Data)
		})
	}
}

package test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"

	"github.com/golang/mock/gomock"
	userhttp "github.com/msyamsula/portofolio/domain/user/http"
	"github.com/msyamsula/portofolio/domain/user/repository"
)

func (s *UserTestSuite) TestAddFriend() {

	type (
		args struct {
			c              context.Context
			method         string
			smallId, bigId int64
		}
		want struct {
			err string
		}
		testCase struct {
			name     string
			args     args
			want     want
			mockFunc func()
		}
	)

	h := userhttp.New(userhttp.Dependencies{
		Service: s.mockSvc,
	})

	testCases := []testCase{
		{
			name: "invalid method",
			args: args{
				c:      context.Background(),
				method: http.MethodDelete,
			},
			want: want{
				err: "",
			},
			mockFunc: func() {
			},
		},
		{
			name: "service error",
			args: args{
				c:      context.Background(),
				method: http.MethodPost,
			},
			want: want{
				err: s.mockErr.Error(),
			},
			mockFunc: func() {
				s.mockSvc.EXPECT().
					AddFriend(gomock.Any(), repository.User{
						Username: "",
						Id:       0,
					}, repository.User{
						Username: "",
						Id:       0,
					}).
					Return(s.mockErr)
			},
		},
		{
			name: "success",
			args: args{
				c:      context.Background(),
				method: http.MethodPost,
			},
			want: want{
				err: "",
			},
			mockFunc: func() {
				s.mockSvc.EXPECT().
					AddFriend(gomock.Any(),
						repository.User{
							Username: "",
							Id:       0,
						}, repository.User{
							Username: "",
							Id:       0,
						}).
					Return(nil)
			},
		},
	}

	for _, tt := range testCases {
		s.Run(tt.name, func() {
			server := httptest.NewServer(http.HandlerFunc(h.ManageFriend))
			defer server.Close()

			reqBody := struct {
				SmallId int64 `json:"small_id"`
				BigId   int64 `json:"big_id"`
			}{
				SmallId: tt.args.smallId,
				BigId:   tt.args.bigId,
			}
			bBody, err := json.Marshal(reqBody)
			s.Nil(err)
			body := bytes.NewBuffer(bBody)
			request, err := http.NewRequest(tt.args.method, server.URL, body)
			s.Nil(err)

			tt.mockFunc()

			c := &http.Client{}
			resp, err := c.Do(request)
			s.Nil(err)
			defer resp.Body.Close()

			httpResponse := userhttp.Response{}
			bResp, _ := io.ReadAll(resp.Body)
			json.Unmarshal(bResp, &httpResponse)

			s.Equal(tt.want.err, httpResponse.Error)
		})
	}
}

func (s *UserTestSuite) TestGetFriends() {

	type (
		args struct {
			c      context.Context
			method string
			id     int64
		}
		want struct {
			err  string
			data []repository.User
		}
		testCase struct {
			name     string
			args     args
			want     want
			mockFunc func()
		}
	)

	h := userhttp.New(userhttp.Dependencies{
		Service: s.mockSvc,
	})

	testCases := []testCase{
		{
			name: "invalid method",
			args: args{
				c:      context.Background(),
				method: http.MethodDelete,
			},
			want: want{
				err: "",
			},
			mockFunc: func() {
			},
		},
		{
			name: "service error",
			args: args{
				c:      context.Background(),
				method: http.MethodGet,
			},
			want: want{
				err: s.mockErr.Error(),
			},
			mockFunc: func() {
				s.mockSvc.EXPECT().
					GetFriends(gomock.Any(), gomock.Any()).
					Return([]repository.User{}, s.mockErr)
			},
		},
		{
			name: "success",
			args: args{
				c:      context.Background(),
				method: http.MethodGet,
				id:     12,
			},
			want: want{
				err: "",
				data: []repository.User{
					{
						Username: "1",
						Id:       1,
					},
					{
						Username: "2",
						Id:       2,
					},
				},
			},
			mockFunc: func() {
				s.mockSvc.EXPECT().
					GetFriends(gomock.Any(), repository.User{
						Id: 12,
					}).
					Return([]repository.User{
						{
							Username: "1",
							Id:       1,
						},
						{
							Username: "2",
							Id:       2,
						},
					}, nil)
			},
		},
	}

	for _, tt := range testCases {
		s.Run(tt.name, func() {
			server := httptest.NewServer(http.HandlerFunc(h.ManageFriend))
			defer server.Close()

			request, err := http.NewRequest(tt.args.method, server.URL, nil)
			query := request.URL.Query()
			query.Add("id", fmt.Sprintf("%d", tt.args.id))
			request.URL.RawQuery = query.Encode()

			tt.mockFunc()

			c := &http.Client{}
			resp, err := c.Do(request)
			s.Nil(err)
			defer resp.Body.Close()

			type r struct {
				Message string            `json:"message,omitempty"`
				Error   string            `json:"error,omitempty"`
				Data    []repository.User `json:"data,omitempty"`
			}
			httpResponse := r{}
			bResp, _ := io.ReadAll(resp.Body)
			json.Unmarshal(bResp, &httpResponse)

			s.Equal(tt.want.err, httpResponse.Error)
			s.Equal(tt.want.data, httpResponse.Data)
		})
	}
}

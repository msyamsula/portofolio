package test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"

	"github.com/golang/mock/gomock"
	userhttp "github.com/msyamsula/portofolio/domain/user/http"
	"github.com/msyamsula/portofolio/domain/user/repository"
)

func (s *UserTestSuite) TestSetUser() {

	type (
		args struct {
			c        context.Context
			username string
			method   string
		}
		want struct {
			err  string
			user repository.User
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
				c:        context.Background(),
				username: "admin",
				method:   http.MethodDelete,
			},
			want: want{
				err:  "",
				user: repository.User{},
			},
			mockFunc: func() {
			},
		},
		{
			name: "service error",
			args: args{
				c:        context.Background(),
				username: "admin",
				method:   http.MethodPost,
			},
			want: want{
				err:  s.mockErr.Error(),
				user: repository.User{},
			},
			mockFunc: func() {
				s.mockSvc.EXPECT().
					SetUser(gomock.Any(), repository.User{
						Username: "admin",
					}).Return(repository.User{}, s.mockErr)
			},
		},
		{
			name: "success",
			args: args{
				c:        context.Background(),
				username: "admin",
				method:   http.MethodPost,
			},
			want: want{
				err: "",
				user: repository.User{
					Username: "admin",
					Id:       99,
				},
			},
			mockFunc: func() {
				s.mockSvc.EXPECT().
					SetUser(gomock.Any(), repository.User{
						Username: "admin",
					}).Return(repository.User{
					Username: "admin",
					Id:       99,
				}, nil)
			},
		},
	}

	for _, tt := range testCases {
		s.Run(tt.name, func() {
			server := httptest.NewServer(http.HandlerFunc(h.ManageUser))
			defer server.Close()

			reqBody := struct {
				Username string `json:"username"`
			}{
				Username: tt.args.username,
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
			s.Equal(tt.want.user, httpResponse.Data)
		})
	}
}

func (s *UserTestSuite) TestGetUser() {

	type (
		args struct {
			c        context.Context
			username string
			method   string
		}
		want struct {
			err  string
			user repository.User
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
				c:        context.Background(),
				username: "admin",
				method:   http.MethodDelete,
			},
			want: want{
				err:  "",
				user: repository.User{},
			},
			mockFunc: func() {
			},
		},
		{
			name: "service error",
			args: args{
				c:        context.Background(),
				username: "admin",
				method:   http.MethodGet,
			},
			want: want{
				err:  s.mockErr.Error(),
				user: repository.User{},
			},
			mockFunc: func() {
				s.mockSvc.EXPECT().
					GetUser(gomock.Any(), "admin").
					Return(repository.User{}, s.mockErr)
			},
		},
		{
			name: "success",
			args: args{
				c:        context.Background(),
				username: "admin",
				method:   http.MethodGet,
			},
			want: want{
				err: "",
				user: repository.User{
					Username: "admin",
					Id:       10,
				},
			},
			mockFunc: func() {
				s.mockSvc.EXPECT().
					GetUser(gomock.Any(), "admin").
					Return(repository.User{
						Username: "admin",
						Id:       10,
					}, nil)
			},
		},
	}

	for _, tt := range testCases {
		s.Run(tt.name, func() {
			server := httptest.NewServer(http.HandlerFunc(h.ManageUser))
			defer server.Close()

			request, err := http.NewRequest(tt.args.method, server.URL, nil)
			s.Nil(err)
			query := request.URL.Query()
			query.Set("username", tt.args.username)
			request.URL.RawQuery = query.Encode()

			tt.mockFunc()

			c := &http.Client{}
			resp, err := c.Do(request)
			s.Nil(err)
			defer resp.Body.Close()

			httpResponse := userhttp.Response{}
			bResp, _ := io.ReadAll(resp.Body)
			json.Unmarshal(bResp, &httpResponse)

			s.Equal(tt.want.err, httpResponse.Error)
			s.Equal(tt.want.user, httpResponse.Data)
		})
	}
}

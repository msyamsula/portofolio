package test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	userhttp "github.com/msyamsula/portofolio/domain/user/http"
	"github.com/msyamsula/portofolio/domain/user/repository"
	"github.com/stretchr/testify/assert"
)

func TestSetUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := NewMockService(ctrl)
	h := userhttp.New(userhttp.Dependencies{
		Service: mockSvc,
	})

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

	mockErr := errors.New("ops")
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
				err:  mockErr.Error(),
				user: repository.User{},
			},
			mockFunc: func() {
				mockSvc.EXPECT().
					SetUser(gomock.Any(), repository.User{
						Username: "admin",
					}).Return(repository.User{}, mockErr)
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
				mockSvc.EXPECT().
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
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(h.ManageUser))
			defer server.Close()

			reqBody := struct {
				Username string `json:"username"`
			}{
				Username: tt.args.username,
			}
			bBody, err := json.Marshal(reqBody)
			assert.Nil(t, err)
			body := bytes.NewBuffer(bBody)
			request, err := http.NewRequest(tt.args.method, server.URL, body)
			assert.Nil(t, err)

			tt.mockFunc()

			c := &http.Client{}
			resp, err := c.Do(request)
			assert.Nil(t, err)
			defer resp.Body.Close()

			httpResponse := userhttp.Response{}
			bResp, _ := io.ReadAll(resp.Body)
			json.Unmarshal(bResp, &httpResponse)

			assert.Equal(t, tt.want.err, httpResponse.Error)
			assert.Equal(t, tt.want.user, httpResponse.Data)
		})
	}
}

func TestGetUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := NewMockService(ctrl)
	h := userhttp.New(userhttp.Dependencies{
		Service: mockSvc,
	})

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

	mockErr := errors.New("ops")
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
				err:  mockErr.Error(),
				user: repository.User{},
			},
			mockFunc: func() {
				mockSvc.EXPECT().
					GetUser(gomock.Any(), "admin").
					Return(repository.User{}, mockErr)
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
				mockSvc.EXPECT().
					GetUser(gomock.Any(), "admin").
					Return(repository.User{
						Username: "admin",
						Id:       10,
					}, nil)
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(h.ManageUser))
			defer server.Close()

			request, err := http.NewRequest(tt.args.method, server.URL, nil)
			assert.Nil(t, err)
			query := request.URL.Query()
			query.Set("username", tt.args.username)
			request.URL.RawQuery = query.Encode()

			tt.mockFunc()

			c := &http.Client{}
			resp, err := c.Do(request)
			assert.Nil(t, err)
			defer resp.Body.Close()

			httpResponse := userhttp.Response{}
			bResp, _ := io.ReadAll(resp.Body)
			json.Unmarshal(bResp, &httpResponse)

			assert.Equal(t, tt.want.err, httpResponse.Error)
			assert.Equal(t, tt.want.user, httpResponse.Data)
		})
	}
}

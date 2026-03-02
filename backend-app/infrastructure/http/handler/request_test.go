package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/suite"
)

type RequestTestSuite struct {
	suite.Suite
}

func (s *RequestTestSuite) TestGetUserIDFromContext_Found() {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	ctx := context.WithValue(req.Context(), "user_id", "user-123")
	req = req.WithContext(ctx)

	result := GetUserIDFromContext(req)
	s.Equal("user-123", result)
}

func (s *RequestTestSuite) TestGetUserIDFromContext_NotFound() {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	result := GetUserIDFromContext(req)
	s.Equal("", result)
}

func (s *RequestTestSuite) TestGetUserIDFromContext_WrongType() {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	ctx := context.WithValue(req.Context(), "user_id", 12345)
	req = req.WithContext(ctx)

	result := GetUserIDFromContext(req)
	s.Equal("", result)
}

func (s *RequestTestSuite) TestQueryParam_Exists() {
	req := httptest.NewRequest(http.MethodGet, "/?name=alice", nil)
	result := QueryParam(req, "name")
	s.Equal("alice", result)
}

func (s *RequestTestSuite) TestQueryParam_Missing() {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	result := QueryParam(req, "name")
	s.Equal("", result)
}

func (s *RequestTestSuite) TestQueryParam_Empty() {
	req := httptest.NewRequest(http.MethodGet, "/?name=", nil)
	result := QueryParam(req, "name")
	s.Equal("", result)
}

func (s *RequestTestSuite) TestPathVar_Exists() {
	req := httptest.NewRequest(http.MethodGet, "/users/42", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "42"})

	result := PathVar(req, "id")
	s.Equal("42", result)
}

func (s *RequestTestSuite) TestPathVar_Missing() {
	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	result := PathVar(req, "id")
	s.Equal("", result)
}

func (s *RequestTestSuite) TestBindJSON_Success() {
	type payload struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}
	body, _ := json.Marshal(payload{Name: "alice", Age: 30})
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))

	var target payload
	err := BindJSON(req, &target)
	s.NoError(err)
	s.Equal("alice", target.Name)
	s.Equal(30, target.Age)
}

func (s *RequestTestSuite) TestBindJSON_InvalidJSON() {
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte("invalid")))

	var target map[string]any
	err := BindJSON(req, &target)
	s.Error(err)
}

func (s *RequestTestSuite) TestBindJSON_EmptyBody() {
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte("")))

	var target map[string]any
	err := BindJSON(req, &target)
	s.Error(err)
}

func (s *RequestTestSuite) TestMustPathVar_Success() {
	req := httptest.NewRequest(http.MethodGet, "/users/42", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "42"})
	rec := httptest.NewRecorder()

	value, err := MustPathVar(rec, req, "id")
	s.NoError(err)
	s.Equal("42", value)
}

func (s *RequestTestSuite) TestMustPathVar_Missing() {
	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	rec := httptest.NewRecorder()

	value, err := MustPathVar(rec, req, "id")
	s.NoError(err)
	s.Equal("", value)
	s.Equal(http.StatusBadRequest, rec.Code)
}

func (s *RequestTestSuite) TestMustQuery_Success() {
	req := httptest.NewRequest(http.MethodGet, "/?page=5", nil)
	rec := httptest.NewRecorder()

	value, err := MustQuery(rec, req, "page")
	s.NoError(err)
	s.Equal("5", value)
}

func (s *RequestTestSuite) TestMustQuery_Missing() {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	value, err := MustQuery(rec, req, "page")
	s.NoError(err)
	s.Equal("", value)
	s.Equal(http.StatusBadRequest, rec.Code)
}

func TestRequestSuite(t *testing.T) {
	suite.Run(t, new(RequestTestSuite))
}

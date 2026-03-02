package handler

import (
"encoding/json"
"net/http"
"net/http/httptest"
"testing"

"github.com/stretchr/testify/suite"
)

type ResponseTestSuite struct {
suite.Suite
}

func (s *ResponseTestSuite) decodeResponse(rec *httptest.ResponseRecorder) Response {
var resp Response
err := json.NewDecoder(rec.Body).Decode(&resp)
s.NoError(err)
return resp
}

// --- JSON ---

func (s *ResponseTestSuite) TestJSON_Success() {
rec := httptest.NewRecorder()
err := JSON(rec, http.StatusOK, map[string]string{"key": "value"})
s.NoError(err)
s.Equal(http.StatusOK, rec.Code)
s.Contains(rec.Header().Get("Content-Type"), "application/json")

resp := s.decodeResponse(rec)
s.Empty(resp.Error)
s.NotNil(resp.Data)
}

func (s *ResponseTestSuite) TestJSON_NilData() {
rec := httptest.NewRecorder()
err := JSON(rec, http.StatusOK, nil)
s.NoError(err)
s.Equal(http.StatusOK, rec.Code)
}

// --- JSONWithMeta ---

func (s *ResponseTestSuite) TestJSONWithMeta_Success() {
rec := httptest.NewRecorder()
meta := Meta{ResponseTime: 42.5}
err := JSONWithMeta(rec, http.StatusOK, "data", meta)
s.NoError(err)

resp := s.decodeResponse(rec)
s.Equal(42.5, resp.Meta.ResponseTime)
}

// --- Error ---

func (s *ResponseTestSuite) TestError_Response() {
rec := httptest.NewRecorder()
err := Error(rec, http.StatusBadRequest, "something went wrong")
s.NoError(err)
s.Equal(http.StatusBadRequest, rec.Code)

resp := s.decodeResponse(rec)
s.Equal("something went wrong", resp.Error)
}

// --- Convenience Methods ---

func (s *ResponseTestSuite) TestOK() {
rec := httptest.NewRecorder()
err := OK(rec, "ok data")
s.NoError(err)
s.Equal(http.StatusOK, rec.Code)
}

func (s *ResponseTestSuite) TestCreated() {
rec := httptest.NewRecorder()
err := Created(rec, "created data")
s.NoError(err)
s.Equal(http.StatusCreated, rec.Code)
}

func (s *ResponseTestSuite) TestNoContent() {
rec := httptest.NewRecorder()
NoContent(rec)
s.Equal(http.StatusNoContent, rec.Code)
s.Equal(0, rec.Body.Len())
}

func (s *ResponseTestSuite) TestBadRequest() {
rec := httptest.NewRecorder()
err := BadRequest(rec, "bad input")
s.NoError(err)
s.Equal(http.StatusBadRequest, rec.Code)

resp := s.decodeResponse(rec)
s.Equal("bad input", resp.Error)
}

func (s *ResponseTestSuite) TestUnauthorized() {
rec := httptest.NewRecorder()
err := Unauthorized(rec, "not allowed")
s.NoError(err)
s.Equal(http.StatusUnauthorized, rec.Code)

resp := s.decodeResponse(rec)
s.Equal("not allowed", resp.Error)
}

func (s *ResponseTestSuite) TestForbidden() {
rec := httptest.NewRecorder()
err := Forbidden(rec, "forbidden")
s.NoError(err)
s.Equal(http.StatusForbidden, rec.Code)

resp := s.decodeResponse(rec)
s.Equal("forbidden", resp.Error)
}

func (s *ResponseTestSuite) TestNotFound() {
rec := httptest.NewRecorder()
err := NotFound(rec, "not found")
s.NoError(err)
s.Equal(http.StatusNotFound, rec.Code)

resp := s.decodeResponse(rec)
s.Equal("not found", resp.Error)
}

func (s *ResponseTestSuite) TestConflict() {
rec := httptest.NewRecorder()
err := Conflict(rec, "conflict")
s.NoError(err)
s.Equal(http.StatusConflict, rec.Code)

resp := s.decodeResponse(rec)
s.Equal("conflict", resp.Error)
}

func (s *ResponseTestSuite) TestInternalError() {
rec := httptest.NewRecorder()
err := InternalError(rec, "server error")
s.NoError(err)
s.Equal(http.StatusInternalServerError, rec.Code)

resp := s.decodeResponse(rec)
s.Equal("server error", resp.Error)
}

func TestResponseSuite(t *testing.T) {
suite.Run(t, new(ResponseTestSuite))
}

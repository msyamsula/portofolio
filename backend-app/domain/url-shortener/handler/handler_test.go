package handler

import (
"bytes"
"encoding/json"
"errors"
"net/http"
"net/http/httptest"
"testing"

"github.com/golang/mock/gomock"
"github.com/gorilla/mux"
"github.com/msyamsula/portofolio/backend-app/mock"
"github.com/stretchr/testify/suite"
)

type URLShortenerHandlerTestSuite struct {
suite.Suite
ctrl    *gomock.Controller
mockSvc *mock.MockURLShortenerService
handler *Handler
router  *mux.Router
}

func (s *URLShortenerHandlerTestSuite) SetupTest() {
s.ctrl = gomock.NewController(s.T())
s.mockSvc = mock.NewMockURLShortenerService(s.ctrl)
s.handler = New(s.mockSvc)
s.router = mux.NewRouter()
s.handler.RegisterRoutes(s.router)
}

func (s *URLShortenerHandlerTestSuite) TearDownTest() {
s.ctrl.Finish()
}

func (s *URLShortenerHandlerTestSuite) TestShorten_Success() {
reqBody := ShortenRequest{LongURL: "https://example.com/very/long/url"}
body, _ := json.Marshal(reqBody)

s.mockSvc.EXPECT().Shorten(gomock.Any(), "https://example.com/very/long/url").Return("abc12345", nil)

req := httptest.NewRequest(http.MethodPost, "/shorten", bytes.NewReader(body))
req.Header.Set("Content-Type", "application/json")
rec := httptest.NewRecorder()
s.router.ServeHTTP(rec, req)

s.Equal(http.StatusCreated, rec.Code)

var resp ShortenResponse
err := json.NewDecoder(rec.Body).Decode(&resp)
s.NoError(err)
}

func (s *URLShortenerHandlerTestSuite) TestShorten_InvalidBody() {
req := httptest.NewRequest(http.MethodPost, "/shorten", bytes.NewReader([]byte("invalid")))
req.Header.Set("Content-Type", "application/json")
rec := httptest.NewRecorder()
s.router.ServeHTTP(rec, req)

s.Equal(http.StatusBadRequest, rec.Code)
}

func (s *URLShortenerHandlerTestSuite) TestShorten_EmptyLongURL() {
reqBody := ShortenRequest{LongURL: ""}
body, _ := json.Marshal(reqBody)

req := httptest.NewRequest(http.MethodPost, "/shorten", bytes.NewReader(body))
req.Header.Set("Content-Type", "application/json")
rec := httptest.NewRecorder()
s.router.ServeHTTP(rec, req)

s.Equal(http.StatusBadRequest, rec.Code)
}

func (s *URLShortenerHandlerTestSuite) TestShorten_ServiceError() {
reqBody := ShortenRequest{LongURL: "https://example.com/url"}
body, _ := json.Marshal(reqBody)

s.mockSvc.EXPECT().Shorten(gomock.Any(), "https://example.com/url").Return("", errors.New("service error"))

req := httptest.NewRequest(http.MethodPost, "/shorten", bytes.NewReader(body))
req.Header.Set("Content-Type", "application/json")
rec := httptest.NewRecorder()
s.router.ServeHTTP(rec, req)

s.Equal(http.StatusInternalServerError, rec.Code)
}

func (s *URLShortenerHandlerTestSuite) TestRedirect_Success() {
s.mockSvc.EXPECT().Expand(gomock.Any(), "abc12345").Return("https://example.com/original", nil)

req := httptest.NewRequest(http.MethodGet, "/abc12345", nil)
rec := httptest.NewRecorder()
s.router.ServeHTTP(rec, req)

s.Equal(http.StatusMovedPermanently, rec.Code)
s.Equal("https://example.com/original", rec.Header().Get("Location"))
}

func (s *URLShortenerHandlerTestSuite) TestRedirect_NotFound() {
s.mockSvc.EXPECT().Expand(gomock.Any(), "notfound").Return("", errors.New("not found"))

req := httptest.NewRequest(http.MethodGet, "/notfound", nil)
rec := httptest.NewRecorder()
s.router.ServeHTTP(rec, req)

s.Equal(http.StatusNotFound, rec.Code)
}

func (s *URLShortenerHandlerTestSuite) TestNew_ReturnsHandler() {
h := New(s.mockSvc)
s.NotNil(h)
}

func (s *URLShortenerHandlerTestSuite) TestRegisterRoutes() {
router := mux.NewRouter()
s.handler.RegisterRoutes(router)
s.NotNil(router)
}

func TestURLShortenerHandlerSuite(t *testing.T) {
suite.Run(t, new(URLShortenerHandlerTestSuite))
}

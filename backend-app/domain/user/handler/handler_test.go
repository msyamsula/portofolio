package handler

import (
"errors"
"net/http"
"net/http/httptest"
"testing"

"github.com/golang/mock/gomock"
"github.com/gorilla/mux"
"github.com/msyamsula/portofolio/backend-app/domain/user/dto"
"github.com/msyamsula/portofolio/backend-app/mock"
"github.com/stretchr/testify/suite"
)

type UserHandlerTestSuite struct {
suite.Suite
ctrl    *gomock.Controller
mockSvc *mock.MockUserService
handler *Handler
router  *mux.Router
}

func (s *UserHandlerTestSuite) SetupTest() {
s.ctrl = gomock.NewController(s.T())
s.mockSvc = mock.NewMockUserService(s.ctrl)
s.handler = New(s.mockSvc)
s.router = mux.NewRouter()
s.handler.RegisterRoutes(s.router)
}

func (s *UserHandlerTestSuite) TearDownTest() {
s.ctrl.Finish()
}

func (s *UserHandlerTestSuite) TestGoogleRedirectURL_Success() {
s.mockSvc.EXPECT().GetRedirectURLGoogle(gomock.Any(), gomock.Any()).Return("https://accounts.google.com/o/oauth2/auth?state=abc", nil)

req := httptest.NewRequest(http.MethodGet, "/google/redirect", nil)
rec := httptest.NewRecorder()
s.router.ServeHTTP(rec, req)

s.Equal(http.StatusTemporaryRedirect, rec.Code)
s.Contains(rec.Header().Get("Location"), "https://accounts.google.com")
}

func (s *UserHandlerTestSuite) TestGoogleRedirectURL_ServiceError() {
s.mockSvc.EXPECT().GetRedirectURLGoogle(gomock.Any(), gomock.Any()).Return("", errors.New("oauth error"))

req := httptest.NewRequest(http.MethodGet, "/google/redirect", nil)
rec := httptest.NewRecorder()
s.router.ServeHTTP(rec, req)

s.Equal(http.StatusInternalServerError, rec.Code)
}

func (s *UserHandlerTestSuite) TestGoogleCallback_MissingParams() {
req := httptest.NewRequest(http.MethodGet, "/google/callback", nil)
rec := httptest.NewRecorder()
s.router.ServeHTTP(rec, req)

s.Equal(http.StatusBadRequest, rec.Code)
}

func (s *UserHandlerTestSuite) TestGoogleCallback_MissingStateCookie() {
req := httptest.NewRequest(http.MethodGet, "/google/callback?state=abc&code=xyz", nil)
rec := httptest.NewRecorder()
s.router.ServeHTTP(rec, req)

s.Equal(http.StatusBadRequest, rec.Code)
}

func (s *UserHandlerTestSuite) TestGoogleCallback_StateMismatch() {
req := httptest.NewRequest(http.MethodGet, "/google/callback?state=abc&code=xyz", nil)
req.AddCookie(&http.Cookie{Name: "oauth_state", Value: "different"})
rec := httptest.NewRecorder()
s.router.ServeHTTP(rec, req)

s.Equal(http.StatusBadRequest, rec.Code)
}

func (s *UserHandlerTestSuite) TestGoogleCallback_Success() {
s.mockSvc.EXPECT().GetAppTokenForGoogleUser(gomock.Any(), "teststate", "testcode").Return("jwt-token-here", nil)

req := httptest.NewRequest(http.MethodGet, "/google/callback?state=teststate&code=testcode", nil)
req.AddCookie(&http.Cookie{Name: "oauth_state", Value: "teststate"})
rec := httptest.NewRecorder()
s.router.ServeHTTP(rec, req)

s.Equal(http.StatusOK, rec.Code)
}

func (s *UserHandlerTestSuite) TestGoogleCallback_ServiceError() {
s.mockSvc.EXPECT().GetAppTokenForGoogleUser(gomock.Any(), "teststate", "testcode").Return("", errors.New("token error"))

req := httptest.NewRequest(http.MethodGet, "/google/callback?state=teststate&code=testcode", nil)
req.AddCookie(&http.Cookie{Name: "oauth_state", Value: "teststate"})
rec := httptest.NewRecorder()
s.router.ServeHTTP(rec, req)

s.Equal(http.StatusInternalServerError, rec.Code)
}

func (s *UserHandlerTestSuite) TestValidateToken_Success() {
userData := dto.UserData{ID: "user-1", Email: "test@example.com", Name: "Test User"}
s.mockSvc.EXPECT().ValidateToken(gomock.Any(), "valid-token").Return(userData, nil)

req := httptest.NewRequest(http.MethodGet, "/validate", nil)
req.Header.Set("Authorization", "Bearer valid-token")
rec := httptest.NewRecorder()
s.router.ServeHTTP(rec, req)

s.Equal(http.StatusOK, rec.Code)
}

func (s *UserHandlerTestSuite) TestValidateToken_MissingAuthHeader() {
req := httptest.NewRequest(http.MethodGet, "/validate", nil)
rec := httptest.NewRecorder()
s.router.ServeHTTP(rec, req)

s.Equal(http.StatusUnauthorized, rec.Code)
}

func (s *UserHandlerTestSuite) TestValidateToken_InvalidAuthFormat_NoBearerPrefix() {
req := httptest.NewRequest(http.MethodGet, "/validate", nil)
req.Header.Set("Authorization", "Basic abc123")
rec := httptest.NewRecorder()
s.router.ServeHTTP(rec, req)

s.Equal(http.StatusUnauthorized, rec.Code)
}

func (s *UserHandlerTestSuite) TestValidateToken_InvalidAuthFormat_NoSpace() {
req := httptest.NewRequest(http.MethodGet, "/validate", nil)
req.Header.Set("Authorization", "Bearertoken")
rec := httptest.NewRecorder()
s.router.ServeHTTP(rec, req)

s.Equal(http.StatusUnauthorized, rec.Code)
}

func (s *UserHandlerTestSuite) TestValidateToken_InvalidToken() {
s.mockSvc.EXPECT().ValidateToken(gomock.Any(), "invalid-token").Return(dto.UserData{}, errors.New("invalid token"))

req := httptest.NewRequest(http.MethodGet, "/validate", nil)
req.Header.Set("Authorization", "Bearer invalid-token")
rec := httptest.NewRecorder()
s.router.ServeHTTP(rec, req)

s.Equal(http.StatusUnauthorized, rec.Code)
}

func (s *UserHandlerTestSuite) TestGenerateRandomState() {
state, err := s.handler.generateRandomState()
s.NoError(err)
s.Len(state, 32)

state2, err := s.handler.generateRandomState()
s.NoError(err)
s.NotEqual(state, state2)
}

func (s *UserHandlerTestSuite) TestNew_ReturnsHandler() {
h := New(s.mockSvc)
s.NotNil(h)
}

func (s *UserHandlerTestSuite) TestRegisterRoutes() {
router := mux.NewRouter()
s.handler.RegisterRoutes(router)
s.NotNil(router)
}

func TestUserHandlerSuite(t *testing.T) {
suite.Run(t, new(UserHandlerTestSuite))
}

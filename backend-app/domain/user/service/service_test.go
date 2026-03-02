package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/golang/mock/gomock"
	"github.com/msyamsula/portofolio/backend-app/domain/user/integration"
	"github.com/msyamsula/portofolio/backend-app/mock"
	"github.com/stretchr/testify/suite"
)

const testSecret = "test-secret-key-for-jwt"

// UserServiceTestSuite defines the test suite for user service
type UserServiceTestSuite struct {
	suite.Suite
	ctrl            *gomock.Controller
	mockAuthService *mock.MockAuthService
	svc             Service
	ctx             context.Context
}

func (s *UserServiceTestSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.mockAuthService = mock.NewMockAuthService(s.ctrl)
	s.svc = New(s.mockAuthService, testSecret, 1*time.Hour)
	s.ctx = context.Background()
}

func (s *UserServiceTestSuite) TearDownTest() {
	s.ctrl.Finish()
}

// --- GetRedirectURLGoogle tests ---

func (s *UserServiceTestSuite) TestGetRedirectURLGoogle_Success() {
	expectedURL := "https://accounts.google.com/o/oauth2/auth?state=random-state"

	s.mockAuthService.EXPECT().GetRedirectURLGoogle(s.ctx, "random-state").Return(expectedURL, nil)

	result, err := s.svc.GetRedirectURLGoogle(s.ctx, "random-state")
	s.NoError(err)
	s.Equal(expectedURL, result)
}

func (s *UserServiceTestSuite) TestGetRedirectURLGoogle_Error() {
	expectedErr := errors.New("oauth config error")

	s.mockAuthService.EXPECT().GetRedirectURLGoogle(s.ctx, "state").Return("", expectedErr)

	result, err := s.svc.GetRedirectURLGoogle(s.ctx, "state")
	s.Error(err)
	s.Equal(expectedErr, err)
	s.Empty(result)
}

// --- GetAppTokenForGoogleUser tests ---

func (s *UserServiceTestSuite) TestGetAppTokenForGoogleUser_Success() {
	userData := integration.UserData{
		ID:    "user-123",
		Email: "test@gmail.com",
		Name:  "Test User",
	}

	s.mockAuthService.EXPECT().GetUserDataGoogle(s.ctx, "state", "code").Return(userData, nil)

	token, err := s.svc.GetAppTokenForGoogleUser(s.ctx, "state", "code")
	s.NoError(err)
	s.NotEmpty(token)

	// Verify the token is valid and contains correct claims
	parsed, err := jwt.Parse(token, func(t *jwt.Token) (any, error) {
		return []byte(testSecret), nil
	})
	s.NoError(err)
	s.True(parsed.Valid)

	claims := parsed.Claims.(jwt.MapClaims)
	s.Equal("user-123", claims["id"])
	s.Equal("test@gmail.com", claims["email"])
	s.Equal("Test User", claims["name"])
}

func (s *UserServiceTestSuite) TestGetAppTokenForGoogleUser_AuthError() {
	expectedErr := errors.New("oauth exchange failed")

	s.mockAuthService.EXPECT().GetUserDataGoogle(s.ctx, "state", "code").Return(integration.UserData{}, expectedErr)

	token, err := s.svc.GetAppTokenForGoogleUser(s.ctx, "state", "code")
	s.Error(err)
	s.Equal(expectedErr, err)
	s.Empty(token)
}

// --- ValidateToken tests ---

func (s *UserServiceTestSuite) TestValidateToken_ValidToken() {
	// Create a valid token
	claims := jwt.MapClaims{
		"id":    "user-123",
		"email": "test@gmail.com",
		"name":  "Test User",
		"exp":   time.Now().Add(1 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(testSecret))
	s.NoError(err)

	userData, err := s.svc.ValidateToken(s.ctx, tokenString)
	s.NoError(err)
	s.Equal("user-123", userData.ID)
	s.Equal("test@gmail.com", userData.Email)
	s.Equal("Test User", userData.Name)
}

func (s *UserServiceTestSuite) TestValidateToken_ExpiredToken() {
	// Create an expired token
	claims := jwt.MapClaims{
		"id":    "user-123",
		"email": "test@gmail.com",
		"name":  "Test User",
		"exp":   time.Now().Add(-1 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(testSecret))
	s.NoError(err)

	_, err = s.svc.ValidateToken(s.ctx, tokenString)
	s.Error(err)
}

func (s *UserServiceTestSuite) TestValidateToken_InvalidSecret() {
	// Create a token with a different secret
	claims := jwt.MapClaims{
		"id":    "user-123",
		"email": "test@gmail.com",
		"name":  "Test User",
		"exp":   time.Now().Add(1 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte("wrong-secret"))
	s.NoError(err)

	_, err = s.svc.ValidateToken(s.ctx, tokenString)
	s.Error(err)
}

func (s *UserServiceTestSuite) TestValidateToken_MalformedToken() {
	_, err := s.svc.ValidateToken(s.ctx, "not.a.valid.token")
	s.Error(err)
}

func (s *UserServiceTestSuite) TestValidateToken_EmptyToken() {
	_, err := s.svc.ValidateToken(s.ctx, "")
	s.Error(err)
}

// --- Constructor test ---

func (s *UserServiceTestSuite) TestNew_ReturnsServiceInstance() {
	svc := New(s.mockAuthService, "secret", time.Hour)
	s.NotNil(svc)
}

// Run the test suite
func TestUserServiceSuite(t *testing.T) {
	suite.Run(t, new(UserServiceTestSuite))
}

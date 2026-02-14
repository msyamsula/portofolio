package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/msyamsula/portofolio/backend-app/domain/user/dto"
	"github.com/msyamsula/portofolio/backend-app/domain/user/integration"
	infraLogger "github.com/msyamsula/portofolio/backend-app/infrastructure/telemetry/logger"
)

// Service defines the interface for user authentication business logic
type Service interface {
	// GetRedirectURLGoogle generates the OAuth redirect URL for Google
	GetRedirectURLGoogle(ctx context.Context, state string) (string, error)

	// GetAppTokenForGoogleUser exchanges OAuth code for app token
	GetAppTokenForGoogleUser(ctx context.Context, state, code string) (string, error)

	// ValidateToken validates an app token and returns user data
	ValidateToken(ctx context.Context, token string) (dto.UserData, error)
}

// userService implements the Service interface
type userService struct {
	externalAuthService integration.AuthService
	appTokenSecret      string
	appTokenTTL         time.Duration
}

// New creates a new user service
func New(externalAuthService integration.AuthService, appTokenSecret string, appTokenTTL time.Duration) Service {
	return &userService{
		externalAuthService: externalAuthService,
		appTokenSecret:      appTokenSecret,
		appTokenTTL:         appTokenTTL,
	}
}

// GetRedirectURLGoogle generates the OAuth redirect URL for Google
func (s *userService) GetRedirectURLGoogle(ctx context.Context, state string) (string, error) {
	return s.externalAuthService.GetRedirectURLGoogle(ctx, state)
}

// GetAppTokenForGoogleUser exchanges OAuth code for app token
func (s *userService) GetAppTokenForGoogleUser(ctx context.Context, state, code string) (string, error) {
	// Get user data from external OAuth provider
	userData, err := s.externalAuthService.GetUserDataGoogle(ctx, state, code)
	if err != nil {
		infraLogger.ErrorError("failed to get user data from OAuth provider", err, map[string]any{
			"state": state,
		})
		return "", err
	}

	// Create app token with user data
	token, err := s.createToken(ctx, userData.ID, userData.Email, userData.Name)
	if err != nil {
		infraLogger.ErrorError("failed to create app token", err, map[string]any{
			"state":   state,
			"user_id": userData.ID,
		})
		return "", err
	}

	return token, nil
}

// ValidateToken validates an app token and returns user data
func (s *userService) ValidateToken(ctx context.Context, tokenString string) (dto.UserData, error) {
	// Parse token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		return []byte(s.appTokenSecret), nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))

	if err != nil {
		infraLogger.ErrorError("failed to parse token", err, nil)
		return dto.UserData{}, err
	}

	// Validate token
	if !token.Valid {
		infraLogger.Error("invalid token", nil)
		return dto.UserData{}, err
	}

	// Extract claims
	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		return dto.UserData{
			ID:    claims["id"].(string),
			Email: claims["email"].(string),
			Name:  claims["name"].(string),
		}, nil
	}

	infraLogger.Error("failed to extract claims from token", nil)
	return dto.UserData{}, err
}

// GenerateRandomState generates a random state string for OAuth flow
func (s *userService) GenerateRandomState() (string, error) {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// createToken creates a JWT token for the user
func (s *userService) createToken(_ context.Context, id, email, name string) (string, error) {
	// Create token claims
	claims := jwt.MapClaims{
		"id":    id,
		"email": email,
		"name":  name,
		"exp":   time.Now().Add(s.appTokenTTL).Unix(),
	}

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign token with secret
	tokenString, err := token.SignedString([]byte(s.appTokenSecret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

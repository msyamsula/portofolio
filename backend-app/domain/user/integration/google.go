package integration

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	infraLogger "github.com/msyamsula/portofolio/backend-app/infrastructure/telemetry/logger"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// UserData represents user information from OAuth provider
type UserData struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

// AuthService defines the interface for external OAuth authentication
type AuthService interface {
	// GetRedirectURLGoogle generates the OAuth redirect URL for Google
	GetRedirectURLGoogle(ctx context.Context, state string) (string, error)

	// GetUserDataGoogle exchanges OAuth code for user data
	GetUserDataGoogle(ctx context.Context, state, code string) (UserData, error)
}

// googleAuthService implements AuthService using Google OAuth2
type googleAuthService struct {
	oauthConfig *oauth2.Config
}

// NewGoogleAuthService creates a new Google OAuth service
func NewGoogleAuthService(clientID, clientSecret, redirectURL string) AuthService {
	return &googleAuthService{
		oauthConfig: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  redirectURL,
			Scopes:       []string{"https://www.googleapis.com/auth/userinfo.profile", "https://www.googleapis.com/auth/userinfo.email"},
			Endpoint:     google.Endpoint,
		},
	}
}

// GetRedirectURLGoogle generates the OAuth redirect URL for Google
func (g *googleAuthService) GetRedirectURLGoogle(ctx context.Context, state string) (string, error) {
	_, span := otel.Tracer("user").Start(ctx, "integration.getRedirectUrlGoogle")
	defer span.End()

	url := g.oauthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)
	span.SetStatus(codes.Ok, "")

	return url, nil
}

// GetUserDataGoogle exchanges OAuth code for user data
func (g *googleAuthService) GetUserDataGoogle(ctx context.Context, state, code string) (UserData, error) {
	_, span := otel.Tracer("user").Start(ctx, "integration.getUserDataGoogle")
	defer span.End()

	// Exchange authorization code for token
	token, err := g.oauthConfig.Exchange(ctx, code)
	if err != nil {
		infraLogger.ErrorError("failed to exchange OAuth code", err, map[string]any{
			"state": state,
		})
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to exchange code")
		return UserData{}, fmt.Errorf("failed to exchange token: %w", err)
	}

	// Create HTTP client with token
	client := g.oauthConfig.Client(ctx, token)

	// Get user info from Google API
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		infraLogger.ErrorError("failed to get user info", err, map[string]any{
			"state": state,
		})
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to get user info")
		return UserData{}, fmt.Errorf("failed to get user info: %w", err)
	}
	defer resp.Body.Close()

	// Parse response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		infraLogger.ErrorError("failed to read response body", err, map[string]any{
			"state": state,
		})
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to read response body")
		return UserData{}, fmt.Errorf("failed to read response body: %w", err)
	}

	// Check HTTP status
	if resp.StatusCode != http.StatusOK {
		infraLogger.Error("user info request failed", map[string]any{
			"state":        state,
			"status_code":   resp.StatusCode,
			"response_body": string(body),
		})
		span.SetStatus(codes.Error, "user info request failed")
		return UserData{}, errors.New("user info request failed")
	}

	// Unmarshal JSON response
	var googleUserData UserData
	if err := json.Unmarshal(body, &googleUserData); err != nil {
		infraLogger.ErrorError("failed to unmarshal user data", err, map[string]any{
			"state":        state,
			"response_body": string(body),
		})
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to unmarshal user data")
		return UserData{}, fmt.Errorf("failed to unmarshal user data: %w", err)
	}

	infraLogger.Info("successfully retrieved user data from Google", map[string]any{
		"state":  state,
		"user_id": googleUserData.ID,
		"email":   googleUserData.Email,
	})

	span.SetAttributes(
		attribute.String("user.id", googleUserData.ID),
		attribute.String("user.email", googleUserData.Email),
	)
	span.SetStatus(codes.Ok, "")

	return googleUserData, nil
}

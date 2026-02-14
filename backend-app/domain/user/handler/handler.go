package handler

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strings"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	"github.com/gorilla/mux"
	"github.com/msyamsula/portofolio/backend-app/domain/user/dto"
	"github.com/msyamsula/portofolio/backend-app/domain/user/service"
	infraHandler "github.com/msyamsula/portofolio/backend-app/infrastructure/http/handler"
	infraLogger "github.com/msyamsula/portofolio/backend-app/infrastructure/telemetry/logger"
)

const (
	oauthStateCookieName = "oauth_state"
)

// Handler handles HTTP requests for user authentication
type Handler struct {
	userService service.Service
}

// New creates a new user handler
func New(svc service.Service) *Handler {
	return &Handler{
		userService: svc,
	}
}

// GoogleRedirectURL handles GET /user/google/redirect requests
// @Summary Get Google OAuth redirect URL
// @Description Generates the OAuth redirect URL for Google authentication
// @Tags user
// @Produce json
// @Success 307 {string} string "redirect"
// @Failure 500 {object} map[string]any
// @Router /user/google/redirect [get]
func (h *Handler) GoogleRedirectURL(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	start := time.Now()

	infraLogger.Info("google redirect url request started", map[string]any{
		"method": r.Method,
		"path":   r.URL.Path,
	})

	// Create child span for handler logic
	tracer := otel.Tracer("user")
	ctx, span := tracer.Start(ctx, "handler.googleRedirectUrl")
	defer span.End()

	// Generate random state for OAuth flow
	state, err := h.generateRandomState()
	if err != nil {
		infraLogger.Error("failed to generate random state", map[string]any{
			"method":      r.Method,
			"path":        r.URL.Path,
			"duration_ms": time.Since(start).Milliseconds(),
		})
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to generate state")
		_ = infraHandler.InternalError(w, "failed to generate state")
		return
	}

	// Add attributes to span
	span.SetAttributes(
		attribute.String("user.oauth_state", state),
	)

	// Set state cookie
	http.SetCookie(w, &http.Cookie{
		Name:     oauthStateCookieName,
		Value:    state,
		Path:     "/",
		Secure:    false, // Set to true in production with HTTPS
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   5 * 60, // 5 minutes
	})

	// Get OAuth redirect URL from service
	redirectURL, err := h.userService.GetRedirectURLGoogle(ctx, state)
	if err != nil {
		infraLogger.Error("failed to get redirect url", map[string]any{
			"method":      r.Method,
			"path":        r.URL.Path,
			"duration_ms": time.Since(start).Milliseconds(),
		})
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to get redirect url")
		_ = infraHandler.InternalError(w, "failed to get redirect url")
		return
	}

	// Redirect to OAuth provider
	infraLogger.Info("redirecting to oauth provider", map[string]any{
		"method":        r.Method,
		"path":          r.URL.Path,
		"redirect_url":  redirectURL,
		"duration_ms":   time.Since(start).Milliseconds(),
	})
	http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
}

// GoogleCallback handles GET /user/google/callback requests
// @Summary Handle Google OAuth callback
// @Description Processes the OAuth callback from Google
// @Tags user
// @Produce json
// @Param code query string true "OAuth authorization code"
// @Param state query string true "OAuth state parameter"
// @Success 200 {object} handler.TokenResponse
// @Failure 400 {object} map[string]any
// @Failure 500 {object} map[string]any
// @Router /user/google/callback [get]
func (h *Handler) GoogleCallback(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	start := time.Now()

	infraLogger.Info("google callback request started", map[string]any{
		"method": r.Method,
		"path":   r.URL.Path,
		"query":  r.URL.RawQuery,
	})

	// Create child span for handler logic
	tracer := otel.Tracer("user")
	ctx, span := tracer.Start(ctx, "handler.googleCallback")
	var err error
	defer func() {
		if err != nil {
			infraLogger.Error("google callback request failed", map[string]any{
				"method":      r.Method,
				"path":        r.URL.Path,
				"duration_ms": time.Since(start).Milliseconds(),
			})
			span.RecordError(err)
			span.SetStatus(codes.Error, "failed to process callback")
		} else {
			infraLogger.Info("google callback request completed", map[string]any{
				"method":      r.Method,
				"path":        r.URL.Path,
				"duration_ms": time.Since(start).Milliseconds(),
			})
		}
		span.End()
	}()

	// Get OAuth parameters from query
	state := infraHandler.QueryParam(r, "state")
	code := infraHandler.QueryParam(r, "code")

	if state == "" || code == "" {
		span.SetStatus(codes.Error, "missing oauth parameters")
		_ = infraHandler.BadRequest(w, "missing oauth parameters")
		return
	}

	// Add attributes to span
	span.SetAttributes(
		attribute.String("user.oauth_state", state),
	)

	// Get state from cookie
	browserCookie, err := r.Cookie(oauthStateCookieName)
	if err != nil {
		span.SetStatus(codes.Error, "missing state cookie")
		_ = infraHandler.BadRequest(w, "missing state cookie")
		return
	}

	// Verify state matches cookie
	if browserCookie.Value != state {
		span.SetStatus(codes.Error, "state mismatch")
		_ = infraHandler.BadRequest(w, "state mismatch")
		return
	}

	// Exchange OAuth code for app token
	token, err := h.userService.GetAppTokenForGoogleUser(ctx, state, code)
	if err != nil {
		span.RecordError(err)
		_ = infraHandler.InternalError(w, err.Error())
		return
	}

	// Return success response with token
	resp := dto.TokenResponse{
		Message: "success",
		Token:   token,
	}
	_ = infraHandler.OK(w, resp)
}

// ValidateToken handles GET /user/validate requests
// @Summary Validate token
// @Description Validates an app token and returns user data
// @Tags user
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Success 200 {object} handler.ValidateTokenResponse
// @Failure 401 {object} map[string]any
// @Failure 500 {object} map[string]any
// @Router /user/validate [get]
func (h *Handler) ValidateToken(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	start := time.Now()

	infraLogger.Info("validate token request started", map[string]any{
		"method": r.Method,
		"path":   r.URL.Path,
	})

	// Create child span for handler logic
	tracer := otel.Tracer("user")
	ctx, span := tracer.Start(ctx, "handler.validateToken")
	var err error
	defer func() {
		if err != nil {
			infraLogger.Error("validate token request failed", map[string]any{
				"method":      r.Method,
				"path":        r.URL.Path,
				"duration_ms": time.Since(start).Milliseconds(),
			})
			span.RecordError(err)
			span.SetStatus(codes.Error, "failed to validate token")
		} else {
			infraLogger.Info("validate token request completed", map[string]any{
				"method":      r.Method,
				"path":        r.URL.Path,
				"duration_ms": time.Since(start).Milliseconds(),
			})
		}
		span.End()
	}()

	// Get authorization header
	bearer := r.Header.Get("Authorization")
	if bearer == "" {
		err = http.ErrNoCookie
		span.SetStatus(codes.Error, "missing authorization header")
		_ = infraHandler.Unauthorized(w, "missing authorization header")
		return
	}

	// Parse bearer token
	bearerToken := strings.Split(bearer, " ")
	if len(bearerToken) != 2 || bearerToken[0] != "Bearer" {
		err = http.ErrNotSupported
		span.SetStatus(codes.Error, "invalid authorization format")
		_ = infraHandler.Unauthorized(w, "invalid authorization format")
		return
	}

	token := bearerToken[1]

	// Validate token and get user data
	userData, err := h.userService.ValidateToken(ctx, token)
	if err != nil {
		span.RecordError(err)
		_ = infraHandler.Unauthorized(w, "invalid token")
		return
	}

	// Return success response with user data
	resp := dto.ValidateTokenResponse{
		Message: "success",
		Data:    userData,
	}
	_ = infraHandler.OK(w, resp)
}

// RegisterRoutes registers all user handler routes
func (h *Handler) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/user/google/redirect", h.GoogleRedirectURL).Methods("GET")
	r.HandleFunc("/user/google/callback", h.GoogleCallback).Methods("GET")
	r.HandleFunc("/user/validate", h.ValidateToken).Methods("GET")
}

// generateRandomState generates a random state string for OAuth flow
func (h *Handler) generateRandomState() (string, error) {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

package middleware

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"github.com/msyamsula/portofolio/backend-app/infrastructure/http/handler"
	"github.com/msyamsula/portofolio/backend-app/infrastructure/telemetry/logger"
)

// Chain chains multiple middleware together in reverse order
func Chain(middleware ...mux.MiddlewareFunc) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		for i := len(middleware) - 1; i >= 0; i-- {
			next = middleware[i](next)
		}
		return next
	}
}

// RecoveryMiddleware recovers from panics
func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				logger.Error("PANIC recovered", map[string]any{"error": err})
				_ = handler.InternalError(w, "internal server error")
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// LoggingMiddleware logs all requests with timing
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		logger.Info("request started", map[string]any{
			"method": r.Method,
			"path":   r.URL.Path,
			"query":  r.URL.RawQuery,
		})

		next.ServeHTTP(w, r)

		logger.Info("request completed", map[string]any{
			"method":    r.Method,
			"path":      r.URL.Path,
			"duration": time.Since(start).Milliseconds(),
		})
	})
}

// CORSMiddleware handles CORS headers
func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// AuthMiddleware validates Bearer token authentication
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		if token == "" {
			handler.Unauthorized(w, "missing authorization header")
			return
		}

		if !strings.HasPrefix(token, "Bearer ") {
			handler.Unauthorized(w, "invalid authorization format")
			return
		}

		// Add user ID to context from token
		userID := strings.TrimPrefix(token, "Bearer ")
		ctx := context.WithValue(r.Context(), "user_id", userID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// AdminMiddleware checks if user is an admin
func AdminMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := handler.GetUserIDFromContext(r)
		if userID == "" {
			handler.Unauthorized(w, "authentication required")
			return
		}

		// TODO: Implement actual admin check
		logger.Info("Admin check for user", map[string]any{"user_id": userID})

		next.ServeHTTP(w, r)
	})
}

// ContentTypeMiddleware enforces JSON content type for POST/PUT requests
func ContentTypeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" || r.Method == "PUT" {
			ct := r.Header.Get("Content-Type")
			if !strings.Contains(ct, "application/json") {
				handler.BadRequest(w, "content-type must be application/json")
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

// TracingMiddleware creates OpenTelemetry HTTP spans for each request
func TracingMiddleware(name string) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return otelhttp.NewHandler(next, name)
	}
}

package handler

import (
	"net/http"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	"github.com/msyamsula/portofolio/backend-app/domain/url-shortener/service"
	infraHandler "github.com/msyamsula/portofolio/backend-app/infrastructure/http/handler"
	"github.com/msyamsula/portofolio/backend-app/infrastructure/telemetry/logger"
)

// Handler handles HTTP requests for URL shortening
type Handler struct {
	service service.Service
}

// New creates a new handler
func New(svc service.Service) *Handler {
	return &Handler{
		service: svc,
	}
}

// Shorten handles POST /shorten requests
// @Summary Shorten URL
// @Description Creates a short URL from a long URL
// @Tags url
// @Accept json
// @Produce json
// @Param x-portofolio header string false "x-portofolio"
// @Param body body handler.ShortenRequest true "Shorten request"
// @Success 201 {object} handler.ShortenResponse
// @Failure 400 {object} map[string]any
// @Failure 500 {object} map[string]any
// @Router /url/shorten [post]
func (h *Handler) Shorten(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	start := time.Now()

	logger.Info("shorten request started", map[string]any{
		"method": r.Method,
		"path":   r.URL.Path,
		"query":  r.URL.RawQuery,
	})

	// Create child span for handler logic
	tracer := otel.Tracer("url-shortener")
	ctx, span := tracer.Start(ctx, "handler.shorten")
	defer span.End()

	// Parse request body
	var req ShortenRequest
	if err := infraHandler.BindJSON(r, &req); err != nil {
		logger.WarnError("shorten request invalid body", err, map[string]any{
			"method":      r.Method,
			"path":        r.URL.Path,
			"duration_ms": time.Since(start).Milliseconds(),
		})
		span.RecordError(err)
		span.SetStatus(codes.Error, "invalid request body")
		_ = infraHandler.BadRequest(w, "invalid request body")
		return
	}

	// Validate long URL
	if req.LongURL == "" {
		logger.Warn("shorten request missing long_url", map[string]any{
			"method":      r.Method,
			"path":        r.URL.Path,
			"duration_ms": time.Since(start).Milliseconds(),
		})
		span.SetStatus(codes.Error, "long_url is required")
		_ = infraHandler.BadRequest(w, "long_url is required")
		return
	}

	// Add URL to span attributes (truncated for security)
	span.SetAttributes(
		attribute.Int("url.long_url_length", len(req.LongURL)),
	)

	// Call service to shorten URL
	shortURL, err := h.service.Shorten(ctx, req.LongURL)
	if err != nil {
		logger.ErrorError("shorten request failed", err, map[string]any{
			"method":          r.Method,
			"path":            r.URL.Path,
			"long_url_length": len(req.LongURL),
			"duration_ms":     time.Since(start).Milliseconds(),
		})
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to shorten URL")
		_ = infraHandler.InternalError(w, "failed to shorten URL")
		return
	}

	// Add short URL to span attributes
	span.SetAttributes(
		attribute.String("url.short_url", shortURL),
	)

	// Return response
	resp := ShortenResponse{ShortURL: shortURL}
	_ = infraHandler.Created(w, resp)

	logger.Info("shorten request completed", map[string]any{
		"method":          r.Method,
		"path":            r.URL.Path,
		"long_url_length": len(req.LongURL),
		"short_url":       shortURL,
		"duration_ms":     time.Since(start).Milliseconds(),
	})
}

// Redirect handles GET /{shortCode} requests - redirects to original URL
// @Summary Redirect short URL
// @Description Redirects to original long URL
// @Tags url
// @Param x-portofolio header string true "x-portofolio"
// @Param shortCode path string true "Short code"
// @Success 301 {string} string "redirect"
// @Failure 400 {object} map[string]any
// @Failure 404 {object} map[string]any
// @Router /{shortCode} [get]
func (h *Handler) Redirect(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	start := time.Now()

	logger.Info("expand request started", map[string]any{
		"method": r.Method,
		"path":   r.URL.Path,
		"query":  r.URL.RawQuery,
	})

	// Create child span for handler logic
	tracer := otel.Tracer("url-shortener")
	ctx, span := tracer.Start(ctx, "handler.redirect")
	defer span.End()

	// Get short code from path variable
	shortCode := infraHandler.PathVar(r, "shortCode")
	if shortCode == "" {
		logger.Warn("expand request missing short_code", map[string]any{
			"method":      r.Method,
			"path":        r.URL.Path,
			"duration_ms": time.Since(start).Milliseconds(),
		})
		span.SetStatus(codes.Error, "shortCode is required")
		_ = infraHandler.BadRequest(w, "shortCode is required")
		return
	}

	// Add short code to span attributes
	span.SetAttributes(
		attribute.String("url.short_code", shortCode),
	)

	// Call service to expand URL
	longURL, err := h.service.Expand(ctx, shortCode)
	if err != nil {
		logger.ErrorError("expand request failed", err, map[string]any{
			"method":      r.Method,
			"path":        r.URL.Path,
			"short_code":  shortCode,
			"duration_ms": time.Since(start).Milliseconds(),
		})
		span.RecordError(err)
		span.SetStatus(codes.Error, "URL not found")
		_ = infraHandler.NotFound(w, "URL not found")
		return
	}

	// Add long URL to span attributes
	span.SetAttributes(
		attribute.Int("url.long_url_length", len(longURL)),
	)

	// Redirect to original long URL
	http.Redirect(w, r, longURL, http.StatusMovedPermanently)

	logger.Info("expand request completed", map[string]any{
		"method":          r.Method,
		"path":            r.URL.Path,
		"short_code":      shortCode,
		"long_url_length": len(longURL),
		"duration_ms":     time.Since(start).Milliseconds(),
	})
}

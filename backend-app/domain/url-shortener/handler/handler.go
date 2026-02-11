package handler

import (
	"net/http"

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
func (h *Handler) Shorten(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Create child span for handler logic
	tracer := otel.Tracer("url-shortener")
	ctx, span := tracer.Start(ctx, "handler.shorten")
	defer span.End()

	// Parse request body
	var req ShortenRequest
	if err := infraHandler.BindJSON(r, &req); err != nil {
		logger.Error("failed to parse request", map[string]any{"error": err})
		span.RecordError(err)
		span.SetStatus(codes.Error, "invalid request body")
		_ = infraHandler.BadRequest(w, "invalid request body")
		return
	}

	// Validate long URL
	if req.LongURL == "" {
		logger.Warn("empty long_url in request", nil)
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
		logger.Error("failed to shorten URL", map[string]any{"error": err})
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
}

// Redirect handles GET /{shortCode} requests - redirects to original URL
func (h *Handler) Redirect(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Create child span for handler logic
	tracer := otel.Tracer("url-shortener")
	ctx, span := tracer.Start(ctx, "handler.redirect")
	defer span.End()

	// Get short code from path variable
	shortCode := infraHandler.PathVar(r, "shortCode")
	if shortCode == "" {
		logger.Warn("empty shortCode in request", nil)
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
		logger.Error("failed to expand URL", map[string]any{"shortCode": shortCode, "error": err})
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
}

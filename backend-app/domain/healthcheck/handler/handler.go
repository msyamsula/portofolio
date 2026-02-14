package handler

import (
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	"github.com/gorilla/mux"
	healthService "github.com/msyamsula/portofolio/backend-app/domain/healthcheck/service"
	infraHandler "github.com/msyamsula/portofolio/backend-app/infrastructure/http/handler"
	"github.com/msyamsula/portofolio/backend-app/infrastructure/telemetry/logger"
)

// Handler handles HTTP requests for health checks
type Handler struct {
	service healthService.Service
}

// New creates a new handler
func New(svc healthService.Service) *Handler {
	return &Handler{
		service: svc,
	}
}

// Check handles GET /health requests
// @Summary Health check
// @Description Returns service health status and uptime
// @Tags health
// @Produce json
// @Success 200 {object} handler.HealthResponse
// @Failure 500 {object} map[string]any
// @Router /health [get]
func (h *Handler) Check(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Create child span for handler logic
	tracer := otel.Tracer("healthcheck")
	ctx, span := tracer.Start(ctx, "handler.check")
	defer span.End()

	// Call service to check health
	status, err := h.service.Check(ctx)
	if err != nil {
		logger.Error("health check failed", map[string]any{"error": err})
		span.RecordError(err)
		span.SetStatus(codes.Error, "health check failed")
		_ = infraHandler.InternalError(w, "health check failed")
		return
	}

	// Get uptime from service
	resp := HealthResponse{
		Status: status,
		Uptime: h.service.Uptime(),
	}

	// Add health status to span attributes
	span.SetAttributes(
		attribute.String("health.status", status),
	)

	_ = infraHandler.OK(w, resp)
}

// RegisterRoutes registers all healthcheck handler routes
func (h *Handler) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/health", h.Check).Methods("GET")
}

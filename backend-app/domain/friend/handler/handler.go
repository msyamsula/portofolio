package handler

import (
	"net/http"
	"strconv"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	"github.com/gorilla/mux"
	"github.com/msyamsula/portofolio/backend-app/domain/friend/dto"
	"github.com/msyamsula/portofolio/backend-app/domain/friend/service"
	infraHandler "github.com/msyamsula/portofolio/backend-app/infrastructure/http/handler"
	infraLogger "github.com/msyamsula/portofolio/backend-app/infrastructure/telemetry/logger"
)

// Handler handles HTTP requests for friend management
type Handler struct {
	friendService service.Service
}

// New creates a new friend handler
func New(svc service.Service) *Handler {
	return &Handler{
		friendService: svc,
	}
}

// AddFriend handles POST /friend/add requests
// @Summary Add friend
// @Description Adds a friendship relationship between two users
// @Tags friend
// @Accept json
// @Produce json
// @Param body body handler.AddFriendRequest true "Add friend request"
// @Success 200 {object} handler.AddFriendResponse
// @Failure 400 {object} map[string]any
// @Failure 500 {object} map[string]any
// @Router /friend/add [post]
func (h *Handler) AddFriend(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	start := time.Now()

	infraLogger.Info("add friend request started", map[string]any{
		"method": r.Method,
		"path":   r.URL.Path,
	})

	// Create child span for handler logic
	tracer := otel.Tracer("friend")
	ctx, span := tracer.Start(ctx, "handler.addFriend")
	defer span.End()

	// Parse request body
	var req dto.AddFriendRequest
	if err := infraHandler.BindJSON(r, &req); err != nil {
		infraLogger.WarnError("add friend request invalid body", err, map[string]any{
			"method":      r.Method,
			"path":        r.URL.Path,
			"duration_ms": time.Since(start).Milliseconds(),
		})
		span.RecordError(err)
		span.SetStatus(codes.Error, "invalid request body")
		_ = infraHandler.BadRequest(w, "invalid request body")
		return
	}

	// Add attributes to span
	span.SetAttributes(
		attribute.Int64("friend.small_id", req.SmallID),
		attribute.Int64("friend.big_id", req.BigID),
	)

	// Create user objects from request
	userA := dto.User{ID: req.SmallID}
	userB := dto.User{ID: req.BigID}

	// Call service to add friendship
	err := h.friendService.AddFriend(ctx, userA, userB)
	if err != nil {
		infraLogger.WarnError("add friend request failed", err, map[string]any{
			"method":      r.Method,
			"path":        r.URL.Path,
			"small_id":    req.SmallID,
			"big_id":      req.BigID,
			"duration_ms": time.Since(start).Milliseconds(),
		})
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to add friend")
		_ = infraHandler.InternalError(w, err.Error())
		return
	}

	// Return success response
	resp := dto.AddFriendResponse{Message: "success"}
	_ = infraHandler.OK(w, resp)

	infraLogger.Info("add friend request completed", map[string]any{
		"method":      r.Method,
		"path":        r.URL.Path,
		"small_id":    req.SmallID,
		"big_id":      req.BigID,
		"duration_ms": time.Since(start).Milliseconds(),
	})
}

// GetFriends handles GET /friend/get requests
// @Summary Get friends
// @Description Retrieves all friends for a given user
// @Tags friend
// @Produce json
// @Param id query int true "User ID"
// @Success 200 {object} handler.GetFriendsResponse
// @Failure 400 {object} map[string]any
// @Failure 500 {object} map[string]any
// @Router /friend/get [get]
func (h *Handler) GetFriends(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	start := time.Now()

	infraLogger.Info("get friends request started", map[string]any{
		"method": r.Method,
		"path":   r.URL.Path,
		"query":  r.URL.RawQuery,
	})

	// Create child span for handler logic
	tracer := otel.Tracer("friend")
	ctx, span := tracer.Start(ctx, "handler.getFriends")
	defer span.End()

	// Get user ID from query parameter
	sid := infraHandler.QueryParam(r, "id")
	if sid == "" {
		infraLogger.Warn("get friends request missing user id", map[string]any{
			"method":      r.Method,
			"path":        r.URL.Path,
			"duration_ms": time.Since(start).Milliseconds(),
		})
		span.SetStatus(codes.Error, "user id is required")
		_ = infraHandler.BadRequest(w, "user id is required")
		return
	}

	// Parse user ID
	id, err := strconv.ParseInt(sid, 10, 64)
	if err != nil {
		infraLogger.WarnError("get friends request invalid user id", err, map[string]any{
			"method":      r.Method,
			"path":        r.URL.Path,
			"user_id":     sid,
			"duration_ms": time.Since(start).Milliseconds(),
		})
		span.RecordError(err)
		span.SetStatus(codes.Error, "invalid user id")
		_ = infraHandler.BadRequest(w, "invalid user id")
		return
	}

	// Add attributes to span
	span.SetAttributes(
		attribute.Int64("friend.user_id", id),
	)

	// Create user object
	user := dto.User{ID: id}

	// Call service to get friends
	users, err := h.friendService.GetFriends(ctx, user)
	if err != nil {
		infraLogger.WarnError("get friends request failed", err, map[string]any{
			"method":      r.Method,
			"path":        r.URL.Path,
			"user_id":     id,
			"duration_ms": time.Since(start).Milliseconds(),
		})
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to get friends")
		_ = infraHandler.InternalError(w, err.Error())
		return
	}

	// Return success response
	resp := dto.GetFriendsResponse{
		Message: "success",
		Data:    users,
	}
	_ = infraHandler.OK(w, resp)

	infraLogger.Info("get friends request completed", map[string]any{
		"method":      r.Method,
		"path":        r.URL.Path,
		"user_id":     id,
		"friend_count": len(users),
		"duration_ms": time.Since(start).Milliseconds(),
	})
}

// RegisterRoutes registers all friend handler routes
func (h *Handler) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/friend/add", h.AddFriend).Methods("POST")
	r.HandleFunc("/friend/get", h.GetFriends).Methods("GET")
}

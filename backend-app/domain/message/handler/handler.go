package handler

import (
	"net/http"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	"github.com/gorilla/mux"
	"github.com/msyamsula/portofolio/backend-app/domain/message/dto"
	"github.com/msyamsula/portofolio/backend-app/domain/message/service"
	infraHandler "github.com/msyamsula/portofolio/backend-app/infrastructure/http/handler"
	infraLogger "github.com/msyamsula/portofolio/backend-app/infrastructure/telemetry/logger"
)

// Handler handles HTTP requests for message management
type Handler struct {
	messageService service.Service
}

// New creates a new message handler
func New(svc service.Service) *Handler {
	return &Handler{
		messageService: svc,
	}
}

// InsertMessage handles POST /message/insert requests
// @Summary Insert message
// @Description Inserts a new message into the conversation
// @Tags message
// @Accept json
// @Produce json
// @Param body body handler.InsertMessageRequest true "Insert message request"
// @Success 200 {object} handler.InsertMessageResponse
// @Failure 400 {object} map[string]any
// @Failure 500 {object} map[string]any
// @Router /message/insert [post]
func (h *Handler) InsertMessage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	start := time.Now()

	infraLogger.Info("insert message request started", map[string]any{
		"method": r.Method,
		"path":   r.URL.Path,
	})

	// Create child span for handler logic
	tracer := otel.Tracer("message")
	ctx, span := tracer.Start(ctx, "handler.insertMessage")
	defer span.End()

	// Parse request body
	var req dto.InsertMessageRequest
	if err := infraHandler.BindJSON(r, &req); err != nil {
		infraLogger.WarnError("insert message request invalid body", err, map[string]any{
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
		attribute.Int64("message.sender_id", req.SenderID),
		attribute.Int64("message.receiver_id", req.ReceiverID),
		attribute.String("message.conversation_id", req.ConversationID),
	)

	// Create message object from request
	msg := dto.Message{
		SenderID:       req.SenderID,
		ReceiverID:     req.ReceiverID,
		ConversationID: req.ConversationID,
		Data:           req.Data,
	}

	// Call service to insert message
	result, err := h.messageService.InsertMessage(ctx, msg)
	if err != nil {
		infraLogger.WarnError("insert message request failed", err, map[string]any{
			"method":         r.Method,
			"path":           r.URL.Path,
			"sender_id":      req.SenderID,
			"receiver_id":    req.ReceiverID,
			"conversation_id": req.ConversationID,
			"duration_ms":    time.Since(start).Milliseconds(),
		})
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to insert message")
		_ = infraHandler.InternalError(w, err.Error())
		return
	}

	// Return success response
	resp := dto.InsertMessageResponse{
		Message: "success",
		Data:    result,
	}
	_ = infraHandler.OK(w, resp)

	infraLogger.Info("insert message request completed", map[string]any{
		"method":         r.Method,
		"path":           r.URL.Path,
		"sender_id":      req.SenderID,
		"receiver_id":    req.ReceiverID,
		"conversation_id": req.ConversationID,
		"duration_ms":    time.Since(start).Milliseconds(),
	})
}

// GetConversation handles GET /message/conversation requests
// @Summary Get conversation
// @Description Retrieves all messages for a conversation
// @Tags message
// @Produce json
// @Param conversation_id query string true "Conversation ID"
// @Success 200 {object} handler.ConversationResponse
// @Failure 400 {object} map[string]any
// @Failure 500 {object} map[string]any
// @Router /message/conversation [get]
func (h *Handler) GetConversation(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	start := time.Now()

	infraLogger.Info("get conversation request started", map[string]any{
		"method": r.Method,
		"path":   r.URL.Path,
		"query":  r.URL.RawQuery,
	})

	// Create child span for handler logic
	tracer := otel.Tracer("message")
	ctx, span := tracer.Start(ctx, "handler.getConversation")
	defer span.End()

	// Get conversation ID from query parameter
	conversationID := infraHandler.QueryParam(r, "conversation_id")
	if conversationID == "" {
		infraLogger.Warn("get conversation request missing conversation_id", map[string]any{
			"method":      r.Method,
			"path":        r.URL.Path,
			"duration_ms": time.Since(start).Milliseconds(),
		})
		span.SetStatus(codes.Error, "conversation_id is required")
		_ = infraHandler.BadRequest(w, "conversation_id is required")
		return
	}

	// Add attributes to span
	span.SetAttributes(
		attribute.String("message.conversation_id", conversationID),
	)

	// Call service to get conversation messages
	messages, err := h.messageService.GetConversation(ctx, conversationID)
	if err != nil {
		infraLogger.WarnError("get conversation request failed", err, map[string]any{
			"method":         r.Method,
			"path":           r.URL.Path,
			"conversation_id": conversationID,
			"duration_ms":    time.Since(start).Milliseconds(),
		})
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to get conversation")
		_ = infraHandler.InternalError(w, err.Error())
		return
	}

	// Return success response
	resp := dto.ConversationResponse{
		Message: "success",
		Data:    messages,
	}
	_ = infraHandler.OK(w, resp)

	infraLogger.Info("get conversation request completed", map[string]any{
		"method":         r.Method,
		"path":           r.URL.Path,
		"conversation_id": conversationID,
		"message_count":   len(messages),
		"duration_ms":    time.Since(start).Milliseconds(),
	})
}

// RegisterRoutes registers all message handler routes
func (h *Handler) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/message/insert", h.InsertMessage).Methods("POST")
	r.HandleFunc("/message/conversation", h.GetConversation).Methods("GET")
}

package api

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/google/uuid"

	"message/service"
)

// Handler exposes the message API over HTTP.
type Handler struct {
	sender *service.Sender
}

// NewHandler constructs an HTTP handler wired to the given Sender.
func NewHandler(sender *service.Sender) *Handler {
	return &Handler{sender: sender}
}

// sendRequest is the JSON body for POST /message/send.
// All UUID fields are required. Payload is the message content.
type sendRequest struct {
	MessageID      uuid.UUID `json:"message_id"`
	ConversationID uuid.UUID `json:"conversation_id"`
	SenderID       uuid.UUID `json:"sender_id"`
	ReceiverID     uuid.UUID `json:"receiver_id"`
	Payload        string    `json:"payload"`
}

// sendResponse is returned on success or idempotent duplicate.
type sendResponse struct {
	Status  string `json:"status"`
	EventID string `json:"event_id,omitempty"`
}

// SendMessage handles POST /message/send.
//
// Returns 200 OK for both new sends and duplicate sends (idempotent).
// The caller cannot distinguish between the two — the DB event_id uniqueness
// is the single source of truth. This prevents clients from retrying
// indefinitely when unsure whether a previous call was received.
//
// Returns 400 on malformed JSON or missing UUIDs.
// Returns 500 only on a DB commit failure (caller should retry).
func (h *Handler) SendMessage(w http.ResponseWriter, r *http.Request) {
	var req sendRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	if err := validateRequest(req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	// JSON-encode the string payload so it is valid JSONB when persisted.
	// "hello world" → `"hello world"` (a JSON string value).
	payloadJSON, _ := json.Marshal(req.Payload)

	err := h.sender.SendMessage(r.Context(), service.SendRequest{
		MessageID:      req.MessageID,
		ConversationID: req.ConversationID,
		SenderID:       req.SenderID,
		ReceiverID:     req.ReceiverID,
		Payload:        payloadJSON,
	})

	// ErrAlreadySent is a business-level idempotency signal, not an HTTP error.
	// Return 200 so callers don't retry on their side — the event log is consistent.
	if err != nil && !errors.Is(err, service.ErrAlreadySent) {
		slog.Error("send message failed", "err", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to send message"})
		return
	}

	writeJSON(w, http.StatusOK, sendResponse{Status: "ok"})
}

// Health handles GET /health — used by docker-compose depends_on healthcheck.
func (h *Handler) Health(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func validateRequest(req sendRequest) error {
	if req.MessageID == uuid.Nil {
		return errors.New("message_id is required")
	}
	if req.ConversationID == uuid.Nil {
		return errors.New("conversation_id is required")
	}
	if req.SenderID == uuid.Nil {
		return errors.New("sender_id is required")
	}
	if req.ReceiverID == uuid.Nil {
		return errors.New("receiver_id is required")
	}
	return nil
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v) //nolint:errcheck
}

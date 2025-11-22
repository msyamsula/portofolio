package handler

import (
	"encoding/json"
	"net/http"

	"github.com/msyamsula/portofolio/backend-app/message/service"
)

func newHttpHandler(cfg HttpConfig) *httpHandler {
	return &httpHandler{
		svc: cfg.Svc,
	}
}

type httpHandler struct {
	svc service.Service
}

func (h *httpHandler) GetConversation(w http.ResponseWriter, req *http.Request) {
	var err error
	var resp conversationResponse

	defer func() {
		if err != nil {
			resp.Error = err.Error()
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(&resp)
	}()
	query := req.URL.Query()
	conversation_id := query.Get("conversation_id")

	ctx := req.Context()
	resp.Conversation, err = h.svc.GetConversation(ctx, conversation_id)
	if err != nil {
		return
	}
}

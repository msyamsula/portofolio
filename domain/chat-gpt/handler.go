package chatgpt

import (
	"encoding/json"
	"net/http"
)

type Handler struct {
	svc *service
}

func NewHandler(svc *service) *Handler {
	return &Handler{
		svc: svc,
	}
}

func (h *Handler) CodeReview(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	defer req.Body.Close()

	type request struct {
		Code string `json:"code"`
	}
	reqBody := &request{}
	err := json.NewDecoder(req.Body).Decode(reqBody)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	review, err := h.svc.CodeReview(reqBody.Code)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	response := struct {
		Review string `json:"review"`
	}{
		Review: review,
	}
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

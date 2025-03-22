package http

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/msyamsula/portofolio/domain/message/repository"
	"github.com/msyamsula/portofolio/domain/message/service"
	"go.opentelemetry.io/otel"
)

type Handler struct {
	Service Service
}

type Service interface {
	AddMessage(c context.Context, msg repository.Message) (repository.Message, error)
	GetConversation(c context.Context, senderId, receiverId int64) ([]repository.Message, error)
}

type Dependencies struct {
	Service Service
}

func New(dep Dependencies) *Handler {
	h := &Handler{
		Service: dep.Service,
	}

	return h
}

func (h *Handler) ManageMesage(w http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodGet {
		h.getConversation(w, req)
	} else if req.Method == http.MethodPost {
		h.addMessage(w, req)
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}

	return
}

func (h *Handler) getConversation(w http.ResponseWriter, req *http.Request) {
	ctx, span := otel.Tracer("").Start(req.Context(), "handler.getConversation")
	defer span.End()
	var err error
	type r struct {
		Error   string               `json:"error,omitempty"`
		Message string               `json:"message,omitempty"`
		Data    []repository.Message `json:"data,omitempty"`
	}
	resp := r{}
	var statusCode int
	defer func() {
		w.WriteHeader(statusCode)
		if err != nil {
			resp.Message = "failed"
			resp.Error = err.Error()
			span.RecordError(err)
		} else {
			resp.Message = "success"
		}

		bresp, _ := json.Marshal(resp)
		w.Write(bresp)
	}()

	query := req.URL.Query()
	var userId int64
	userId, err = strconv.ParseInt(query.Get("userId"), 10, 64)
	if err != nil {
		err = service.ErrBadRequest
		statusCode = http.StatusBadRequest
		return
	}

	var pairId int64
	pairId, err = strconv.ParseInt(query.Get("pairId"), 10, 64)
	if err != nil {
		err = service.ErrBadRequest
		statusCode = http.StatusBadRequest
		return
	}

	var msgs []repository.Message
	msgs, err = h.Service.GetConversation(ctx, userId, pairId)
	if err != nil {
		statusCode = http.StatusInternalServerError
		return
	}

	resp.Data = msgs
	statusCode = http.StatusOK
	return
}

func (h *Handler) addMessage(w http.ResponseWriter, req *http.Request) {
	ctx, span := otel.Tracer("").Start(req.Context(), "handler.getConversation")
	defer span.End()
	var err error
	type r struct {
		Error   string             `json:"error,omitempty"`
		Message string             `json:"message,omitempty"`
		Data    repository.Message `json:"data,omitempty"`
	}
	resp := r{}
	var statusCode int
	defer func() {
		w.WriteHeader(statusCode)
		if err != nil {
			resp.Message = "failed"
			resp.Error = err.Error()
			span.RecordError(err)
		} else {
			resp.Message = "success"
		}

		bresp, _ := json.Marshal(resp)
		w.Write(bresp)
	}()

	var bBody []byte
	bBody, err = io.ReadAll(req.Body)
	if err != nil {
		statusCode = http.StatusBadRequest
		return
	}
	msg := repository.Message{}
	err = json.Unmarshal(bBody, &msg)
	if err != nil {
		statusCode = http.StatusInternalServerError
		return
	}

	msg, err = h.Service.AddMessage(ctx, msg)
	if err != nil {
		statusCode = http.StatusInternalServerError
		return
	}

	resp.Data = msg
	statusCode = http.StatusOK
}

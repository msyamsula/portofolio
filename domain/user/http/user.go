package http

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/msyamsula/portofolio/domain/user/repository"
	"go.opentelemetry.io/otel"
)

type Handler struct {
	service Service
}

type Service interface {
	SetUser(c context.Context, user repository.User) (repository.User, error)
	GetUser(c context.Context, username string) (repository.User, error)
	AddFriend(c context.Context, userA, userB repository.User) error
	GetFriends(c context.Context, user repository.User) ([]repository.User, error)
}

type Dependencies struct {
	Service Service
}

func New(dep Dependencies) *Handler {
	return &Handler{
		service: dep.Service,
	}
}

type Response struct {
	Message string          `json:"message,omitempty"`
	Error   string          `json:"error,omitempty"`
	Data    repository.User `json:"data,omitempty"`
}

func (h *Handler) ManageUser(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		h.getUser(w, req)
	case http.MethodPost:
		h.setUser(w, req)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}

}

func (h *Handler) setUser(w http.ResponseWriter, req *http.Request) {

	ctx, span := otel.Tracer("").Start(req.Context(), "handler.setUser")
	defer span.End()

	response := Response{
		Message: "",
		Error:   "",
		Data:    repository.User{},
	}
	var err error
	var statusCode int
	defer func() {
		w.WriteHeader(statusCode)
		if err != nil {
			// error response
			response.Message = "failed"
			response.Error = err.Error()
			span.RecordError(err)

			resp, _ := json.Marshal(response)
			w.Write([]byte(resp))
		} else {
			// success
			response.Message = "success"
			resp, _ := json.Marshal(response)
			w.Write(resp)
		}
	}()

	type body struct {
		Username string `json:"username"`
		Online   bool   `json:"online"`
	}
	reqBody := body{}
	bBody, _ := io.ReadAll(req.Body)
	json.Unmarshal(bBody, &reqBody)

	response.Data.Username = reqBody.Username
	response.Data.Online = reqBody.Online
	response.Data, err = h.service.SetUser(ctx, response.Data)
	if err != nil {
		statusCode = http.StatusInternalServerError
		return
	}

	statusCode = http.StatusOK

}

func (h *Handler) getUser(w http.ResponseWriter, req *http.Request) {
	ctx, span := otel.Tracer("").Start(req.Context(), "handler.getUser")
	defer span.End()

	response := Response{
		Message: "success",
		Error:   "",
		Data:    repository.User{},
	}
	var err error
	var statusCode int
	defer func() {
		w.WriteHeader(statusCode)
		if err != nil {
			// failed
			response.Message = "failed"
			span.RecordError(err)
			response.Error = err.Error()
			resp, _ := json.Marshal(response)
			w.Write([]byte(resp))
		} else {
			// success
			response.Message = "success"
			resp, _ := json.Marshal(response)
			w.Write(resp)
		}
	}()

	query := req.URL.Query()
	username := query.Get("username")
	response.Data.Username = username

	response.Data, err = h.service.GetUser(ctx, response.Data.Username)
	if err != nil {
		statusCode = http.StatusInternalServerError
		return
	}

	statusCode = http.StatusOK

}

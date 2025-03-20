package http

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/msyamsula/portofolio/domain/user/repository"
)

type Handler struct {
	service Service
}

type Service interface {
	SetUser(c context.Context, user repository.User) (repository.User, error)
	GetUser(c context.Context, username string) (repository.User, error)
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
	if req.Method == http.MethodGet {
		h.getUser(w, req)
	} else if req.Method == http.MethodPost {
		h.setUser(w, req)
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}

	return
}

func (h *Handler) setUser(w http.ResponseWriter, req *http.Request) {

	response := Response{
		Message: "success",
		Error:   "",
		Data:    repository.User{},
	}

	type body struct {
		Username string `json:"username"`
	}
	reqBody := body{}
	bBody, _ := io.ReadAll(req.Body)
	json.Unmarshal(bBody, &reqBody)

	ctx := req.Context()
	response.Data.Username = reqBody.Username
	var err error

	response.Data, err = h.service.SetUser(ctx, response.Data)
	if err != nil {
		// fmt.Println("here", user, err, username)
		response.Error = err.Error()
		resp, err := json.Marshal(response)
		if err != nil {
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(resp))
		return
	}

	resp, err := json.Marshal(response)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(resp)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(resp)

}

func (h *Handler) getUser(w http.ResponseWriter, req *http.Request) {

	response := Response{
		Message: "success",
		Error:   "",
		Data:    repository.User{},
	}

	query := req.URL.Query()
	username := query.Get("username")
	ctx := req.Context()
	response.Data.Username = username
	var err error

	response.Data, err = h.service.GetUser(ctx, response.Data.Username)
	if err != nil {
		// fmt.Println("here", user, err, username)
		response.Error = err.Error()
		resp, err := json.Marshal(response)
		if err != nil {
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(resp))
		return
	}

	resp, err := json.Marshal(response)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(resp)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(resp)

}

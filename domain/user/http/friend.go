package http

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/msyamsula/portofolio/domain/user/repository"
	"go.opentelemetry.io/otel"
)

func (h *Handler) ManageFriend(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodPost:
		h.addFriend(w, req)
	case http.MethodGet:
		h.getFriends(w, req)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (h *Handler) addFriend(w http.ResponseWriter, req *http.Request) {

	ctx, span := otel.Tracer("").Start(req.Context(), "handler.addFriend")
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
		SmallId int64 `json:"small_id"`
		BigId   int64 `json:"big_id"`
	}
	reqBody := body{}
	bBody, _ := io.ReadAll(req.Body)
	json.Unmarshal(bBody, &reqBody)

	userA := repository.User{
		Id: reqBody.SmallId,
	}
	userB := repository.User{
		Id: reqBody.BigId,
	}
	err = h.service.AddFriend(ctx, userA, userB)
	if err != nil {
		statusCode = http.StatusInternalServerError
		return
	}

	statusCode = http.StatusOK

}

func (h *Handler) getFriends(w http.ResponseWriter, req *http.Request) {

	ctx, span := otel.Tracer("").Start(req.Context(), "handler.getFriends")
	defer span.End()

	type r struct {
		Message string            `json:"message,omitempty"`
		Error   string            `json:"error,omitempty"`
		Data    []repository.User `json:"data,omitempty"`
	}

	var response r
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

	query := req.URL.Query()
	sid := query.Get("id")
	var id int64
	id, err = strconv.ParseInt(sid, 10, 64)
	if err != nil {
		statusCode = http.StatusInternalServerError
		return
	}

	var users []repository.User
	users, err = h.service.GetFriends(ctx, repository.User{
		Id: id,
	})
	if err != nil {
		statusCode = http.StatusInternalServerError
		return
	}

	statusCode = http.StatusOK
	response.Data = users

}

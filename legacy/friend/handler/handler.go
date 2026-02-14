package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	repo "github.com/msyamsula/portofolio/backend-app/friend/persistent"
	"github.com/msyamsula/portofolio/backend-app/friend/service"
	"go.opentelemetry.io/otel"
)

type handler struct {
	svc service.Service
}

func (h *handler) AddFriend(w http.ResponseWriter, req *http.Request) {

	ctx, span := otel.Tracer("").Start(req.Context(), "handler.addFriend")
	defer span.End()

	response := Response{
		Message: "",
		Error:   "",
		Data:    repo.User{},
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

	userA := repo.User{
		Id: reqBody.SmallId,
	}
	userB := repo.User{
		Id: reqBody.BigId,
	}
	err = h.svc.AddFriend(ctx, userA, userB)
	if err != nil {
		statusCode = http.StatusInternalServerError
		return
	}

	statusCode = http.StatusOK

}

func (h *handler) GetFriends(w http.ResponseWriter, req *http.Request) {

	ctx, span := otel.Tracer("").Start(req.Context(), "handler.getFriends")
	defer span.End()

	var response struct {
		Message string      `json:"message,omitempty"`
		Error   string      `json:"error,omitempty"`
		Data    []repo.User `json:"data,omitempty"`
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

	query := req.URL.Query()
	sid := query.Get("id")
	var id int64
	id, err = strconv.ParseInt(sid, 10, 64)
	if err != nil {
		statusCode = http.StatusInternalServerError
		return
	}

	var users []repo.User
	users, err = h.svc.GetFriends(ctx, repo.User{
		Id: id,
	})
	if err != nil {
		statusCode = http.StatusInternalServerError
		return
	}

	statusCode = http.StatusOK
	response.Data = users

}

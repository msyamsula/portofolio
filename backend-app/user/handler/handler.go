package handler

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/msyamsula/portofolio/backend-app/user/persistent"
	"github.com/msyamsula/portofolio/backend-app/user/service"
	"go.opentelemetry.io/otel"
)

type handler struct {
	svc service.Service
}

func (h *handler) InsertUser(w http.ResponseWriter, req *http.Request) {
	ctx, span := otel.Tracer("").Start(req.Context(), "handler.setUser")
	defer span.End()

	response := Response{
		Message: "",
		Error:   "",
		Data:    persistent.User{},
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
	response.Data, err = h.svc.SetUser(ctx, response.Data)
	if err != nil {
		statusCode = http.StatusInternalServerError
		return
	}

	statusCode = http.StatusOK
}

func (h *handler) GetUser(w http.ResponseWriter, req *http.Request) {
	ctx, span := otel.Tracer("").Start(req.Context(), "handler.getUser")
	defer span.End()

	response := Response{
		Message: "success",
		Error:   "",
		Data:    persistent.User{},
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

	response.Data, err = h.svc.GetUser(ctx, response.Data.Username)
	if err != nil {
		statusCode = http.StatusInternalServerError
		return
	}

	statusCode = http.StatusOK
}

package handler

import (
	"encoding/json"
	"net/http"

	"github.com/msyamsula/portofolio/backend-app/user/service"
	"go.opentelemetry.io/otel"
)

type handler struct {
	svc service.Service
}

func (h *handler) GoogleRedirectUrl(w http.ResponseWriter, req *http.Request) {
	ctx, span := otel.Tracer("").Start(req.Context(), "handler.setUser")
	defer span.End()

	var err error
	defer func() {
		if err != nil {
			// error response
			span.RecordError(err)
		}
	}()

	var browserCookie *http.Cookie
	browserCookie, err = req.Cookie("session_id")
	var url string
	url, err = h.svc.GetRedirectUrlGoogle(ctx, browserCookie.Value)
	if err != nil {
		return
	}

	http.Redirect(w, req, url, http.StatusTemporaryRedirect)
}

func (h *handler) GetAppTokenForGoogle(w http.ResponseWriter, req *http.Request) {
	ctx, span := otel.Tracer("").Start(req.Context(), "handler.getUser")
	defer span.End()

	var response TokenResponse
	var err error
	var statusCode int
	defer func() {
		w.WriteHeader(statusCode)
		if err != nil {
			// failed
			response.Message = "failed"
			span.RecordError(err)
			response.Error = err.Error()
		} else {
			// success
			response.Message = "success"
		}

		json.NewEncoder(w).Encode(response)
	}()

	query := req.URL.Query()
	state := query.Get("state")
	code := query.Get("code")

	var browserCookie *http.Cookie
	browserCookie, err = req.Cookie("session_id")
	if err != nil {
		return
	}

	response.Token, err = h.svc.GetAppTokenForGoogleUser(ctx, browserCookie.Raw, state, code)
}

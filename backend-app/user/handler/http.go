package handler

import (
	"encoding/json"
	"net/http"

	"github.com/msyamsula/portofolio/backend-app/pkg/logger"
	"github.com/msyamsula/portofolio/backend-app/pkg/randomizer"
	"github.com/msyamsula/portofolio/backend-app/user/service"
	"go.opentelemetry.io/otel"
)

var (
	oauthStateCookieKey = "oauth_state"
)

type httpHandler struct {
	svc        service.Service
	randomizer randomizer.Randomizer
}

func (h *httpHandler) GoogleRedirectUrl(w http.ResponseWriter, req *http.Request) {
	ctx, span := otel.Tracer("").Start(req.Context(), "handler.setUser")
	defer span.End()

	var err error
	defer func() {
		if err != nil {
			// error response
			span.RecordError(err)
		}
	}()

	var state string
	state, err = h.randomizer.String()
	if err != nil {
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     oauthStateCookieKey,
		Value:    state,
		Path:     "/",
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	// log.Println(browserCookie, "DEBUG", browserCookie == nil)
	var url string
	url, err = h.svc.GetRedirectUrlGoogle(ctx, state)
	if err != nil {
		return
	}

	logger.Logger.Info("redirected")
	http.Redirect(w, req, url, http.StatusTemporaryRedirect)
}

func (h *httpHandler) GetAppTokenForGoogle(w http.ResponseWriter, req *http.Request) {
	ctx, span := otel.Tracer("").Start(req.Context(), "handler.getUser")
	defer span.End()

	var response TokenResponse
	var err error
	defer func() {
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
	browserCookie, err = req.Cookie(oauthStateCookieKey)
	if err != nil {
		return
	}

	logger.Logger.Info(browserCookie.Value)
	logger.Logger.Info(state)
	if browserCookie.Value != state {
		// reject if cookie is not the same with state
		return
	}

	response.Token, err = h.svc.GetAppTokenForGoogleUser(ctx, state, code)
}

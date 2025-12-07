package handler

import (
	"encoding/json"
	"net/http"

	"github.com/msyamsula/portofolio/backend-app/pkg/logger"
	"github.com/msyamsula/portofolio/backend-app/pkg/randomizer"
	"github.com/msyamsula/portofolio/backend-app/user/service"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
)

var (
	oauthStateCookieKey = "oauth_state"
)

type httpHandler struct {
	randomizer randomizer.Randomizer
	svc        service.Service
}

func (h *httpHandler) GoogleRedirectUrl(w http.ResponseWriter, req *http.Request) {
	ctx, span := otel.Tracer("").Start(req.Context(), "handler.GoogleRedirectUrl")
	var err error
	defer func() {
		if err != nil {
			// error response
			logger.Logger.Error(err.Error())
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}
		span.End()
	}()

	var state string
	state, err = h.randomizer.String()
	if err != nil {
		logger.Logger.Error(err.Error())
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
		logger.Logger.Error(err.Error())
		return
	}

	logger.Logger.Info("redirected")
	http.Redirect(w, req, url, http.StatusTemporaryRedirect)
}

func (h *httpHandler) GetAppTokenForGoogle(w http.ResponseWriter, req *http.Request) {
	ctx, span := otel.Tracer("").Start(req.Context(), "handler.GetAppTokenForGoogle")
	var err error

	var response TokenResponse
	defer func() {
		if err != nil {
			// failed
			logger.Logger.Error(err.Error())
			response.Message = "failed"
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			response.Error = err.Error()
		} else {
			// success
			response.Message = "success"
		}
		json.NewEncoder(w).Encode(response)
		span.End()
	}()

	query := req.URL.Query()
	state := query.Get("state")
	code := query.Get("code")

	var browserCookie *http.Cookie
	browserCookie, err = req.Cookie(oauthStateCookieKey)
	if err != nil {
		logger.Logger.Error(err.Error())
		return
	}

	if browserCookie.Value != state {
		// reject if cookie is not the same with state
		logger.Logger.Error("browser cookie and state mismatch")
		return
	}

	response.Token, err = h.svc.GetAppTokenForGoogleUser(ctx, state, code)
}

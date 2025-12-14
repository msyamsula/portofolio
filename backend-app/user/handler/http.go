package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/msyamsula/portofolio/backend-app/pkg/logger"
	"github.com/msyamsula/portofolio/backend-app/pkg/randomizer"
	"github.com/msyamsula/portofolio/backend-app/user/service"
	internaltoken "github.com/msyamsula/portofolio/backend-app/user/service/internal-token"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

var (
	oauthStateCookieKey = "oauth_state"
)

type httpHandler struct {
	randomizer randomizer.Randomizer
	svc        service.Service
	internal   internaltoken.InternalToken
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
		MaxAge:   1 * 60, // hardcoded 1 minute
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

// handler for SayHello
func (h *httpHandler) ValidateToken(w http.ResponseWriter, req *http.Request) {
	var span trace.Span
	var ctx context.Context
	ctx, span = otel.Tracer("").Start(req.Context(), "handler.ValidateToken")
	var err error
	type response struct {
		Header
		internaltoken.UserData `json:"data"`
	}
	var resp response
	defer func() {
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			logger.Logger.Error(err.Error())
			resp.Error = err.Error()
			w.WriteHeader(http.StatusInternalServerError)
		}
		span.End()
		json.NewEncoder(w).Encode(&resp)
	}()

	bearer := req.Header.Get("Authorization")
	if bearer == "" {
		err = errors.New("bearer token not found")
		return
	}
	bearerToken := strings.Split(bearer, " ")
	if len(bearerToken) != 2 {
		err = errors.New("invalid bearer format")
		return
	}

	token := bearerToken[1]

	var userData internaltoken.UserData
	userData, err = h.internal.ValidateToken(ctx, token)
	if err != nil {
		return
	}
	resp.UserData = userData
}

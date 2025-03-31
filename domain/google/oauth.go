package google

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	userrepo "github.com/msyamsula/portofolio/domain/user/repository"
	usersvc "github.com/msyamsula/portofolio/domain/user/service"
	"go.opentelemetry.io/otel"
)

type Service struct {
	googleClientId      string
	googleSecret        string
	googleRedirectOauth string
	userSvc             *usersvc.Service
	redirectChat        string
}

type Dependencies struct {
	GoogleClientId      string
	GoogleRedirectOauth string
	GoogleSecret        string
	UserSvc             *usersvc.Service
	RedirectChat        string
}

func New(dep Dependencies) *Service {
	svc := &Service{
		googleClientId:      dep.GoogleClientId,
		googleRedirectOauth: dep.GoogleRedirectOauth,
		googleSecret:        dep.GoogleSecret,
		userSvc:             dep.UserSvc,
		redirectChat:        dep.RedirectChat,
	}

	return svc
}

var (
	pkce = make(map[string]string)
)

func (s *Service) RedirectToOauthServer(w http.ResponseWriter, r *http.Request) {

	ctx, span := otel.Tracer("").Start(r.Context(), "oauth.RedirectToOauthServer")
	defer span.End()

	var err error
	var errCode int
	defer func() {
		if err != nil {
			span.RecordError(err)
			w.WriteHeader(errCode)
		}
	}()

	var req *http.Request
	url := "https://accounts.google.com/o/oauth2/v2/auth"
	req, err = http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		errCode = http.StatusInternalServerError
		return
	}
	req.WithContext(ctx)

	query := req.URL.Query()
	query.Set("client_id", s.googleClientId)
	query.Set("redirect_uri", s.googleRedirectOauth)
	query.Set("scope", "email openid profile")
	query.Set("response_type", "code")
	req.URL.RawQuery = query.Encode()

	http.Redirect(w, req, req.URL.String(), http.StatusPermanentRedirect)
}

func (s *Service) getAccessToken(c context.Context, code string) (string, error) {

	ctx, span := otel.Tracer("").Start(c, "oauth.getAccessToken")
	defer span.End()

	// Create POST request
	data := url.Values{}
	data.Set("client_id", s.googleClientId)
	data.Set("redirect_uri", s.googleRedirectOauth)
	data.Set("grant_type", "authorization_code")
	data.Set("client_secret", s.googleSecret)
	data.Set("code", code)
	data.Set("scope", "")

	tokenUrl := "https://oauth2.googleapis.com/token"
	req, err := http.NewRequest("POST", tokenUrl, bytes.NewBufferString(data.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = req.WithContext(ctx)

	// Send request
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Parse response
	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	token, ok := result["access_token"].(string)
	if !ok {
		return "", errors.New("ops")
	}

	return token, nil
}

func (s *Service) getUserInfo(c context.Context, token string) (string, error) {
	_, span := otel.Tracer("").Start(c, "oauth.getUserInfo")
	defer span.End()

	url := "https://www.googleapis.com/oauth2/v3/userinfo"

	var req *http.Request
	var err error
	req, err = http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	var resp *http.Response
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	resp, err = client.Do(req)
	if err != nil {
		return "", err
	}

	result := map[string]interface{}{}
	json.NewDecoder(resp.Body).Decode(&result)

	name, ok := result["name"].(string)
	if !ok {
		return "", errors.New("ops")
	}

	return name, nil
}

func (s *Service) RedirectToChat(w http.ResponseWriter, r *http.Request) {

	ctx, span := otel.Tracer("").Start(r.Context(), "oauth.RedirectToChat")
	defer span.End()

	var err error
	var errCode int
	defer func() {
		if err != nil {
			span.RecordError(err)
			w.WriteHeader(errCode)
		}
	}()

	query := r.URL.Query()
	code := query.Get("code")

	// get access token
	var token string
	token, err = s.getAccessToken(ctx, code)
	if err != nil {
		errCode = http.StatusInternalServerError
		return
	}

	// get user info
	var name string
	name, err = s.getUserInfo(ctx, token)
	if err != nil {
		errCode = http.StatusInternalServerError
		return
	}

	// register user
	user := userrepo.User{
		Username: name,
		Online:   true,
	}
	user, err = s.userSvc.SetUser(ctx, user)
	if err != nil {
		errCode = http.StatusInternalServerError
		return
	}

	url := fmt.Sprintf("%s?username=%s&id=%d", s.redirectChat, user.Username, user.Id)
	http.Redirect(w, r, url, http.StatusPermanentRedirect)
}

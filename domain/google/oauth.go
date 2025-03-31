package google

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/msyamsula/portofolio/domain/url/hasher"
	"github.com/msyamsula/portofolio/domain/user/repository"
	usersvc "github.com/msyamsula/portofolio/domain/user/service"
	"go.opentelemetry.io/otel"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type Service struct {
	userSvc      *usersvc.Service
	redirectChat string
	oauthConfig  *oauth2.Config
	oauthState   string
}

type Dependencies struct {
	GoogleClientId      string
	GoogleRedirectOauth string
	GoogleSecret        string
	UserSvc             *usersvc.Service
	RedirectChat        string
	OauthStateLength    int64
	OauthCharacters     string
}

func New(dep Dependencies) *Service {
	var oauthConfig = &oauth2.Config{
		ClientID:     dep.GoogleClientId,
		ClientSecret: dep.GoogleSecret,
		RedirectURL:  dep.GoogleRedirectOauth,
		Scopes:       []string{"email", "profile", "openid"},
		Endpoint:     google.Endpoint,
	}

	h := hasher.New(hasher.Config{
		Length: dep.OauthStateLength,
		Word:   dep.OauthCharacters,
	})

	svc := &Service{
		userSvc:      dep.UserSvc,
		redirectChat: dep.RedirectChat,
		oauthConfig:  oauthConfig,
		oauthState:   h.Hash(context.Background()),
	}

	return svc
}

func (s *Service) HandleLogin(w http.ResponseWriter, r *http.Request) {
	_, span := otel.Tracer("").Start(r.Context(), "oauth.Login")
	defer span.End()

	url := s.oauthConfig.AuthCodeURL(s.oauthState, oauth2.AccessTypeOffline)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		http.Error(w, "error create request oauth", http.StatusInternalServerError)
		return
	}
	query := req.URL.Query()
	query.Set("state", s.oauthState)
	req.URL.RawQuery = query.Encode()

	http.Redirect(w, req, req.URL.String(), http.StatusFound)
}

func (s *Service) HandleCallback(w http.ResponseWriter, r *http.Request) {
	ctx, span := otel.Tracer("").Start(r.Context(), "oauth.Login")
	defer span.End()

	// use state if necessary

	code := r.URL.Query().Get("code")
	token, err := s.oauthConfig.Exchange(ctx, code)
	if err != nil {
		http.Error(w, "Failed to exchange code for token", http.StatusInternalServerError)
		return
	}

	client := s.oauthConfig.Client(ctx, token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v3/userinfo")
	if err != nil {
		http.Error(w, "Failed to get user info", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()
	result := map[string]interface{}{}
	json.NewDecoder(resp.Body).Decode(&result)

	email, ok := result["email"].(string)
	if !ok {
		http.Error(w, "user not found", http.StatusInternalServerError)
		return
	}
	var name strings.Builder
	for i := range email {
		if email[i] == '@' {
			break
		}
		name.WriteByte(email[i])
	}

	user := repository.User{
		Username: name.String(),
		Online:   true,
	}
	user, err = s.userSvc.SetUser(ctx, user)
	if err != nil {
		http.Error(w, "set user failed", http.StatusInternalServerError)
		return
	}

	url := fmt.Sprintf("%s?username=%s&id=%d", s.redirectChat, user.Username, user.Id)
	http.Redirect(w, r, url, http.StatusPermanentRedirect)
}

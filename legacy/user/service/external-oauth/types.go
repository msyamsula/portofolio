package externaloauth

import (
	"golang.org/x/oauth2"
)

type UserData struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

type AuthConfig struct {
	GoogleOauthConfig *oauth2.Config
}

package externaloauth

import (
	"context"
)

type AuthService interface {
	GetRedirectUrlGoogle(ctx context.Context, state string) (string, error)
	GetUserDataGoogle(ctx context.Context, state, code string) (UserData, error)
}

func NewAuthService(config AuthConfig) AuthService {
	return &authService{
		oauthConfigGoogle: config.GoogleOauthConfig,
	}
}

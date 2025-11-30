package externaloauth

import (
	"context"

	"github.com/msyamsula/portofolio/backend-app/pkg/randomizer"
)

type AuthService interface {
	GetRedirectUrlGoogle(ctx context.Context, browserCookies string) (string, error)
	GetUserDataGoogle(ctx context.Context, browserCookies, state, code string) (UserData, error)
}

func NewAuthService(config AuthConfig) AuthService {
	return &authService{
		oauthConfigGoogle: config.GoogleOauthConfig,
		randomizer: randomizer.NewStringRandomizer(randomizer.StringRandomizerConfig{
			Size:          20,                                                               // internal config no need to expose it, hardly change
			CharacterPool: "abcdefghijklmnopqrstuvwxyz1234567890ABCDEFGHIJKLMNOPQRSTUVWXYZ", // internal config no need to expose it, hardly change
		}),
	}
}

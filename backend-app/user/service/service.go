package service

// mockgen -destination=test/mock_persistence.go -package=test . PersistenceLayer

import (
	"context"
	"errors"
	"time"

	"github.com/msyamsula/portofolio/backend-app/pkg/cache"
	externaloauth "github.com/msyamsula/portofolio/backend-app/user/service/external-oauth"
	internaltoken "github.com/msyamsula/portofolio/backend-app/user/service/internal-token"
	"go.opentelemetry.io/otel"
)

type service struct {
	external          externaloauth.AuthService
	internal          internaltoken.InternalToken
	sessionManagement cache.Cache

	userLoginTtl time.Duration
}

func (s *service) GetAppTokenForGoogleUser(c context.Context, cookies, state, code string) (string, error) {
	ctx, span := otel.Tracer("").Start(c, "service.SetUser")
	defer span.End()

	var err error
	defer func() {
		if err != nil {
			span.RecordError(err)
		}
	}()

	var savedState string
	savedState, err = s.sessionManagement.Get(ctx, cookies)
	if err != nil {
		return "", err
	}

	if savedState != state {
		return "", errors.New("mismatch state")
	}

	// allowed exchange
	var userData externaloauth.UserData
	userData, err = s.external.GetUserDataGoogle(ctx, cookies, state, code)
	if err != nil {
		return "", err
	}

	// create token with expiry time
	var appToken string
	appToken, err = s.internal.CreateToken(userData.ID, userData.Email, userData.Name)
	if err != nil {
		return "", err
	}

	// save token to session
	// will be deleted when logout or 3 hours
	err = s.sessionManagement.Set(ctx, appToken, "true", s.userLoginTtl)
	if err != nil {
		return "", err
	}

	return appToken, nil
}

func (s *service) GetRedirectUrlGoogle(c context.Context, browserCookies string) (string, error) {
	ctx, span := otel.Tracer("").Start(c, "service.GetUser")
	defer span.End()

	var err error
	defer func() {
		if err != nil {
			span.RecordError(err)
		}
	}()

	var url string
	url, err = s.external.GetRedirectUrlGoogle(ctx, browserCookies)
	if err != nil {
		return "", err
	}

	return url, nil
}

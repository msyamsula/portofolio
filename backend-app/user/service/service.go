package service

// mockgen -destination=test/mock_persistence.go -package=test . PersistenceLayer

import (
	"context"

	"github.com/msyamsula/portofolio/backend-app/pkg/cache"
	"github.com/msyamsula/portofolio/backend-app/pkg/logger"
	externaloauth "github.com/msyamsula/portofolio/backend-app/user/service/external-oauth"
	internaltoken "github.com/msyamsula/portofolio/backend-app/user/service/internal-token"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

type service struct {
	external          externaloauth.AuthService
	internal          internaltoken.InternalToken
	sessionManagement cache.Cache
}

func (s *service) GetAppTokenForGoogleUser(ctx context.Context, state, code string) (string, error) {
	var span trace.Span
	ctx, span = otel.Tracer("").Start(ctx, "service.GetAppTokenForGoogleUser")

	var err error
	defer func() {
		if err != nil {
			span.RecordError(err)
		}
		span.End()
	}()

	// allowed exchange
	var userData externaloauth.UserData
	userData, err = s.external.GetUserDataGoogle(ctx, state, code)
	if err != nil {
		logger.Logger.Error(err.Error())
		return "", err
	}
	logger.Logger.Info(userData)

	// create token with expiry time
	var appToken string
	appToken, err = s.internal.CreateToken(ctx, userData.ID, userData.Email, userData.Name)
	if err != nil {
		logger.Logger.Error(err.Error())
		return "", err
	}

	return appToken, nil
}

func (s *service) GetRedirectUrlGoogle(ctx context.Context, state string) (string, error) {
	var span trace.Span
	ctx, span = otel.Tracer("").Start(ctx, "service.GetRedirectUrlGoogle")
	var err error
	defer func() {
		if err != nil {
			span.RecordError(err)
		}
		span.End()
	}()

	var url string
	url, err = s.external.GetRedirectUrlGoogle(ctx, state)
	if err != nil {
		logger.Logger.Error(err.Error())
		return "", err
	}

	return url, nil
}

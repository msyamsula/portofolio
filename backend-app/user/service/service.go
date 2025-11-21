package service

// mockgen -destination=test/mock_persistence.go -package=test . PersistenceLayer

import (
	"context"

	"github.com/msyamsula/portofolio/backend-app/user/persistent"
	"go.opentelemetry.io/otel"
)

type service struct {
	persistence persistent.Repository
}

type ServiceConfig struct {
	Persistence persistent.Repository
}

func (s *service) SetUser(c context.Context, user persistent.User) (persistent.User, error) {
	ctx, span := otel.Tracer("").Start(c, "service.SetUser")
	defer span.End()

	var err error
	defer func() {
		if err != nil {
			span.RecordError(err)
		}
	}()

	user, err = s.persistence.InsertUser(ctx, user)
	if err != nil {
		return user, err
	}

	return user, nil
}

func (s *service) GetUser(c context.Context, username string) (persistent.User, error) {
	ctx, span := otel.Tracer("").Start(c, "service.GetUser")
	defer span.End()

	var err error
	defer func() {
		if err != nil {
			span.RecordError(err)
		}
	}()

	var result persistent.User

	// cache miss
	result, err = s.persistence.GetUser(ctx, username)
	if err != nil {
		return result, err
	}

	return result, nil
}

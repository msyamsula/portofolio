package internaltoken

import "context"

type InternalToken interface {
	CreateToken(ctx context.Context, id, email, name string) (string, error)
	ValidateToken(ctx context.Context, tokenString string) (UserData, error)
}

func NewInternalToken(cfg InternalTokenConfig) InternalToken {
	return &internalToken{
		appTokenSecret: cfg.AppTokenSecret,
		appTokenTtl:    cfg.AppTokenTtl,
	}
}

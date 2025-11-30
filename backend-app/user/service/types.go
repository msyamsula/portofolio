package service

import (
	"time"

	"github.com/msyamsula/portofolio/backend-app/pkg/cache"
	externaloauth "github.com/msyamsula/portofolio/backend-app/user/service/external-oauth"
	internaltoken "github.com/msyamsula/portofolio/backend-app/user/service/internal-token"
)

type UserData struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

type ServiceConfig struct {
	External          externaloauth.AuthService
	Internal          internaltoken.InternalToken
	SessionManagement cache.Cache

	UserLoginTtl time.Duration
}

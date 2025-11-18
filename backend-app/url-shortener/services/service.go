package services

import (
	"context"

	"github.com/msyamsula/portofolio/backend-app/url-shortener/cache"
	"github.com/msyamsula/portofolio/backend-app/url-shortener/persistent"
)

type Config struct {
	Persistence persistent.Repository
	Cache       cache.Repository

	CharacterPool string
	Size          int
	CallbackUri   string
}

type Service interface {
	Short(context.Context, string) (string, error)
	SetLongUrl(context.Context, string, string) error
	GetLongUrl(context.Context, string) (string, error)
}

func New(config Config) Service {
	return &urlShortener{
		persistence:   config.Persistence,
		cache:         config.Cache,
		characterPool: config.CharacterPool,
		size:          config.Size,
		callbackUri:   config.CallbackUri,
	}
}

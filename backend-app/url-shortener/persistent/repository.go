package persistent

import (
	"context"
)

type Repository interface {
	GetShortUrl(c context.Context, longUrl string) (string, error)
	SaveShortUrl(c context.Context, shortUrl, longUrl string) error
	GetLongUrl(c context.Context, shortUrl string) (string, error)
}

package services

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"

	"github.com/msyamsula/portofolio/backend-app/url-shortener/cache"
	"github.com/msyamsula/portofolio/backend-app/url-shortener/persistent"
)

type urlShortener struct {
	persistence persistent.Repository
	cache       cache.Repository

	characterPool string
	size          int
	callbackUri   string
}

func (s *urlShortener) SetLongUrl(ctx context.Context, shortUrl, longUrl string) error {
	err := s.persistence.SaveShortUrl(ctx, shortUrl, longUrl)
	if err != nil {
		return err
	}

	// set to redis
	if s.cache != nil {
		s.cache.Set(ctx, shortUrl, longUrl)
	}
	return nil
}

func (s *urlShortener) Short(ctx context.Context, longUrl string) (string, error) {

	if s.cache != nil {
		if shortUrl, err := s.cache.Get(ctx, longUrl); err == nil {
			// cache hit
			return fmt.Sprintf("%s/%s", s.callbackUri, shortUrl), nil
		}
	}

	shortUrl, err := s.persistence.GetShortUrl(ctx, longUrl)
	if err == nil {
		// exist in database
		if s.cache != nil {
			s.cache.Set(ctx, shortUrl, longUrl)
			s.cache.Set(ctx, longUrl, shortUrl)
		}

		return fmt.Sprintf("%s/%s", s.callbackUri, shortUrl), nil
	}

	// need to be shorten, create one
	shortUrl = s.createShort()

	// save to db
	err = s.persistence.SaveShortUrl(ctx, shortUrl, longUrl)
	if err != nil {
		return "", err
	}

	if s.cache != nil {
		// save to cache
		s.cache.Set(ctx, shortUrl, longUrl)
		s.cache.Set(ctx, longUrl, shortUrl)
	}

	return fmt.Sprintf("%s/%s", s.callbackUri, shortUrl), nil
}

func (s *urlShortener) createShort() string {
	var short strings.Builder
	for i := 0; i < s.size; i++ {
		bidx, _ := rand.Int(rand.Reader, big.NewInt(int64(len(s.characterPool))))
		idx := int(bidx.Int64())
		short.WriteByte(s.characterPool[int(idx)])
	}

	return short.String()
}

func (s *urlShortener) GetLongUrl(ctx context.Context, shortUrl string) (string, error) {
	if s.cache != nil {
		if longUrl, err := s.cache.Get(ctx, shortUrl); err == nil {
			// cache hit
			return longUrl, nil
		}
	}

	// check to db
	longUrl, err := s.persistence.GetLongUrl(ctx, shortUrl)
	if err != nil {
		return "", err
	}

	// set to cache
	if s.cache != nil {
		s.cache.Set(ctx, longUrl, shortUrl)
		s.cache.Set(ctx, shortUrl, longUrl)
	}

	return longUrl, nil
}

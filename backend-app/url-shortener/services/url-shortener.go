package services

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"

	"github.com/msyamsula/portofolio/backend-app/url-shortener/cache"
	"github.com/msyamsula/portofolio/backend-app/url-shortener/persistent"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
)

type urlShortener struct {
	persistence persistent.Repository
	cache       cache.Repository

	characterPool string
	size          int
	callbackUri   string
}

func (s *urlShortener) Short(c context.Context, longUrl string) (string, error) {
	var err error
	ctx, span := otel.Tracer("service").Start(c, "Short")
	defer func() {
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}

		span.End()
	}()

	if s.cache != nil {
		var shortUrl string
		if shortUrl, err = s.cache.Get(ctx, longUrl); err == nil {
			// cache hit
			return fmt.Sprintf("%s/%s", s.callbackUri, shortUrl), nil
		}
	}

	var shortUrl string
	shortUrl, err = s.persistence.GetShortUrl(ctx, longUrl)
	if err == nil {
		// exist in database
		if s.cache != nil {
			s.cache.Set(ctx, shortUrl, longUrl)
			s.cache.Set(ctx, longUrl, shortUrl)
		}

		return fmt.Sprintf("%s/%s", s.callbackUri, shortUrl), nil
	}

	// need to be shorten, create one
	shortUrl, err = s.createShort(ctx)
	if err != nil {
		return "", err
	}

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

func (s *urlShortener) createShort(c context.Context) (string, error) {
	var err error
	_, span := otel.Tracer("service").Start(c, "createShort")
	defer func() {
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}

		span.End()
	}()

	var short strings.Builder
	for i := 0; i < s.size; i++ {
		var bidx *big.Int
		bidx, err = rand.Int(rand.Reader, big.NewInt(int64(len(s.characterPool))))
		if err != nil {
			return "", err
		}
		idx := int(bidx.Int64())
		short.WriteByte(s.characterPool[int(idx)])
	}

	return short.String(), nil
}

func (s *urlShortener) GetLongUrl(c context.Context, shortUrl string) (string, error) {
	ctx, span := otel.Tracer("service").Start(c, "getlongurl")
	var err error
	defer func() {
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}

		span.End()
	}()

	var longUrl string
	if s.cache != nil {
		if longUrl, err = s.cache.Get(ctx, shortUrl); err == nil {
			// cache hit
			return longUrl, nil
		}
	}

	// check to db
	longUrl, err = s.persistence.GetLongUrl(ctx, shortUrl)
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

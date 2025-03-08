package url

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
)

func (s *Service) GetLongUrl(c context.Context, shortUrl string) (string, error) {
	ctx, span := otel.Tracer("").Start(c, "service.GetLongUrl")
	defer span.End()

	return s.repo.GetLongUrl(ctx, shortUrl)
}

func (s *Service) SetShortUrl(c context.Context, longUrl string) (string, error) {
	ctx, span := otel.Tracer("").Start(c, "service.SetShortUrl")
	defer span.End()

	shortUrl := s.hasher.Hash(ctx)
	err := s.repo.SetShortUrl(ctx, shortUrl, longUrl)
	if err != nil {
		return "", err
	}
	shortUrl = fmt.Sprintf("%s/%s", s.host, shortUrl)
	return shortUrl, nil
}

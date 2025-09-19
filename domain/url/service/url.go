package url

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"

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

	shortUrl := s.shortenUrl(ctx)
	err := s.repo.SetShortUrl(ctx, shortUrl, longUrl)
	if err != nil {
		return "", err
	}
	shortUrl = fmt.Sprintf("%s/%s", s.host, shortUrl)
	return shortUrl, nil
}

func (s *Service) shortenUrl(c context.Context) string {
	_, span := otel.Tracer("").Start(c, "service.Hash")
	defer span.End()

	var result strings.Builder
	limit := big.NewInt(int64(len(s.characterPool)))
	for i := 0; i < int(s.length); i++ {
		randomIdx, _ := rand.Int(rand.Reader, limit)
		idx := randomIdx.Int64()
		idx %= limit.Int64()
		result.WriteByte(s.characterPool[idx])
	}

	return result.String()
}

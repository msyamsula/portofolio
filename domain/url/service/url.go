package url

import (
	"context"
)

func (s *Service) GetLongUrl(ctx context.Context, shortUrl string) (string, error) {
	return s.repo.GetLongUrl(ctx, shortUrl)
}

func (s *Service) SetShortUrl(ctx context.Context, longUrl string) (string, error) {
	shortUrl := s.hasher.Hash()
	err := s.repo.SetShortUrl(ctx, shortUrl, longUrl)
	if err != nil {
		return "", err
	}
	return shortUrl, nil
}

package service

import (
	"context"
	"errors"

	"auto-pr-review/internal/githubprdescriptor"
)

var (
	ErrInvalidGitHubLinkID     = errors.New("invalid github link id")
	ErrGitHubLinkAlreadyExists = errors.New("github link already exists")
	ErrGitHubLinkNotFound      = errors.New("github link not found")
)

type GitHubLink struct {
	ID               int64   `json:"id"`
	URL              string  `json:"url"`
	PeriodicInterval *string `json:"periodic_interval,omitempty"`
}

type GitHubLinkRepository interface {
	CreateGitHubLink(ctx context.Context, githubURL string) (GitHubLink, error)
	ListGitHubLinks(ctx context.Context) ([]GitHubLink, error)
	DeleteGitHubLink(ctx context.Context, id int64) error
}

type GitHubLinkService struct {
	repo GitHubLinkRepository
}

func NewGitHubLinkService(repo GitHubLinkRepository) *GitHubLinkService {
	return &GitHubLinkService{repo: repo}
}

func (s *GitHubLinkService) Register(ctx context.Context, rawURL string) (GitHubLink, error) {
	normalizedURL, err := githubprdescriptor.NormalizeGitHubURL(rawURL)
	if err != nil {
		return GitHubLink{}, err
	}

	return s.repo.CreateGitHubLink(ctx, normalizedURL)
}

func (s *GitHubLinkService) List(ctx context.Context) ([]GitHubLink, error) {
	return s.repo.ListGitHubLinks(ctx)
}

func (s *GitHubLinkService) Delete(ctx context.Context, id int64) error {
	if id < 1 {
		return ErrInvalidGitHubLinkID
	}

	return s.repo.DeleteGitHubLink(ctx, id)
}

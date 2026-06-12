package service

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/url"
	"strings"
)

var (
	ErrInvalidGitHubURL        = errors.New("invalid github url")
	ErrInvalidGitHubLinkID     = errors.New("invalid github link id")
	ErrGitHubLinkAlreadyExists = errors.New("github link already exists")
	ErrGitHubLinkNotFound      = errors.New("github link not found")
)

type GitHubLink struct {
	ID  int64  `json:"id"`
	URL string `json:"url"`
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
	normalizedURL, err := NormalizeGitHubURL(rawURL)
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

func NormalizeGitHubURL(rawURL string) (string, error) {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return "", ErrInvalidGitHubURL
	}

	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("%w: malformed url", ErrInvalidGitHubURL)
	}
	if !parsed.IsAbs() || parsed.Opaque != "" {
		return "", fmt.Errorf("%w: absolute url is required", ErrInvalidGitHubURL)
	}
	if parsed.Scheme != "https" {
		return "", fmt.Errorf("%w: scheme must be https", ErrInvalidGitHubURL)
	}
	if parsed.User != nil {
		return "", fmt.Errorf("%w: user info is not allowed", ErrInvalidGitHubURL)
	}
	if parsed.Host == "" {
		return "", fmt.Errorf("%w: host is required", ErrInvalidGitHubURL)
	}

	host := strings.ToLower(parsed.Hostname())
	if host != "github.com" && host != "www.github.com" {
		return "", fmt.Errorf("%w: host must be github.com", ErrInvalidGitHubURL)
	}
	if ip := net.ParseIP(host); ip != nil {
		return "", fmt.Errorf("%w: ip host is not allowed", ErrInvalidGitHubURL)
	}
	if parsed.Path == "" || parsed.Path == "/" {
		return "", fmt.Errorf("%w: path is required", ErrInvalidGitHubURL)
	}
	if parsed.Fragment != "" {
		return "", fmt.Errorf("%w: fragments are not allowed", ErrInvalidGitHubURL)
	}

	parsed.Scheme = "https"
	parsed.Host = host
	return parsed.String(), nil
}

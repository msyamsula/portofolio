package githubprdescriptor

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	ErrInvalidGitHubURL = errors.New("invalid github url")
	ErrInvalidURL       = errors.New("invalid github pull request url")
	ErrNotFound         = errors.New("github resource not found")
)

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

type APIError struct {
	StatusCode int
	Body       string
	URL        string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("github request failed: status %d: %s", e.StatusCode, e.Body)
}

type PRRef struct {
	Owner  string
	Repo   string
	Number int
}

type PRDiscussion struct {
	URL            string
	Title          string
	State          string
	Author         string
	Body           string
	Branch         string
	IssueComments  []Comment
	ReviewComments []ReviewComment
	Reviews        []Review
}

type PR struct {
	HTMLURL string `json:"html_url"`
	Title   string `json:"title"`
	State   string `json:"state"`
	Body    string `json:"body"`
	User    struct {
		Login string `json:"login"`
	} `json:"user"`
	Head struct {
		Ref string `json:"ref"`
	} `json:"head"`
}

type Comment struct {
	User struct {
		Login string `json:"login"`
	} `json:"user"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ReviewComment struct {
	User struct {
		Login string `json:"login"`
	} `json:"user"`
	Path      string    `json:"path"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Review struct {
	User struct {
		Login string `json:"login"`
	} `json:"user"`
	State       string    `json:"state"`
	Body        string    `json:"body"`
	SubmittedAt time.Time `json:"submitted_at"`
}

type PoolConfig struct {
	MaxIdleConns        int
	MaxIdleConnsPerHost int
	MaxConnsPerHost     int
}

type Client struct {
	httpClient *http.Client
	tokenPath  string
}

func NewClient(client *http.Client, tokenPath string, pool PoolConfig) *Client {
	if client == nil {
		if pool.MaxIdleConns <= 0 {
			pool.MaxIdleConns = 10
		}
		if pool.MaxIdleConnsPerHost <= 0 {
			pool.MaxIdleConnsPerHost = 10
		}
		if pool.MaxConnsPerHost <= 0 {
			pool.MaxConnsPerHost = 10
		}
		client = &http.Client{
			Transport: &http.Transport{
				MaxIdleConns:        pool.MaxIdleConns,
				MaxIdleConnsPerHost: pool.MaxIdleConnsPerHost,
				MaxConnsPerHost:     pool.MaxConnsPerHost,
				IdleConnTimeout:     90 * time.Second,
			},
		}
	}
	return &Client{
		httpClient: client,
		tokenPath:  strings.TrimSpace(tokenPath),
	}
}

func ParseURL(rawURL string) (PRRef, error) {
	normalizedURL, err := NormalizeGitHubURL(rawURL)
	if err != nil {
		return PRRef{}, err
	}

	parsed, err := url.Parse(normalizedURL)
	if err != nil {
		return PRRef{}, fmt.Errorf("%w: malformed url", ErrInvalidURL)
	}

	parts := strings.Split(strings.Trim(parsed.Path, "/"), "/")
	if len(parts) < 4 || parts[2] != "pull" {
		return PRRef{}, fmt.Errorf("%w: expected /owner/repo/pull/number", ErrInvalidURL)
	}

	number, err := strconv.Atoi(parts[3])
	if err != nil || number < 1 {
		return PRRef{}, fmt.Errorf("%w: invalid pull request number", ErrInvalidURL)
	}

	return PRRef{
		Owner:  parts[0],
		Repo:   parts[1],
		Number: number,
	}, nil
}

func (c *Client) FetchDiscussion(ctx context.Context, pr PRRef) (PRDiscussion, error) {
	if c == nil {
		return PRDiscussion{}, fmt.Errorf("github client is required")
	}
	basePath := fmt.Sprintf("https://api.github.com/repos/%s/%s", url.PathEscape(pr.Owner), url.PathEscape(pr.Repo))

	var pull PR
	if err := c.get(ctx, fmt.Sprintf("%s/pulls/%d", basePath, pr.Number), &pull); err != nil {
		return PRDiscussion{}, err
	}

	var issueComments []Comment
	if err := c.get(ctx, fmt.Sprintf("%s/issues/%d/comments?per_page=100", basePath, pr.Number), &issueComments); err != nil {
		return PRDiscussion{}, err
	}

	var reviewComments []ReviewComment
	if err := c.get(ctx, fmt.Sprintf("%s/pulls/%d/comments?per_page=100", basePath, pr.Number), &reviewComments); err != nil {
		return PRDiscussion{}, err
	}

	var reviews []Review
	if err := c.get(ctx, fmt.Sprintf("%s/pulls/%d/reviews?per_page=100", basePath, pr.Number), &reviews); err != nil {
		return PRDiscussion{}, err
	}

	return PRDiscussion{
		URL:            pull.HTMLURL,
		Title:          pull.Title,
		State:          pull.State,
		Author:         pull.User.Login,
		Body:           pull.Body,
		Branch:         pull.Head.Ref,
		IssueComments:  issueComments,
		ReviewComments: reviewComments,
		Reviews:        reviews,
	}, nil
}

func (c *Client) get(ctx context.Context, requestURL string, target any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	token, err := c.token()
	if err != nil {
		return err
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("github request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return &APIError{
			StatusCode: resp.StatusCode,
			Body:       strings.TrimSpace(string(body)),
			URL:        requestURL,
		}
	}

	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		return fmt.Errorf("decode github response: %w", err)
	}

	return nil
}

func (c *Client) token() (string, error) {
	tokenPath := strings.TrimSpace(c.tokenPath)
	if tokenPath == "" {
		return "", fmt.Errorf("github token path is required")
	}
	data, err := os.ReadFile(tokenPath)
	if err != nil {
		return "", fmt.Errorf("read github token file: %w", err)
	}
	token := strings.TrimSpace(string(data))
	if token == "" {
		return "", fmt.Errorf("github token file is empty")
	}
	return token, nil
}

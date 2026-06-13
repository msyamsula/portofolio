package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"auto-pr-review/internal/llm"
)

var ErrInvalidPullRequestURL = errors.New("invalid github pull request url")

type PRDiscussionReviewConfig struct {
	Analyzer     llm.Analyzer
	LLM          llm.Config
	GitHubToken  string
	OutputDir    string
	HTTPClient   *http.Client
	GitHubClient *GitHubClient
	FileHandler  FileHandler
}

type PRReviewer struct {
	githubClient *GitHubClient
	fileHandler  FileHandler
	config       PRDiscussionReviewConfig
	analyzer     llm.Analyzer
}

type GitHubClient struct {
	httpClient *http.Client
	token      string
}

type FileHandler interface {
	Replace(path string, data []byte) error
}

type AtomicFileHandler struct{}

func NewPRReviewer(cfg PRDiscussionReviewConfig) (*PRReviewer, error) {
	githubClient := cfg.GitHubClient
	if githubClient == nil {
		client := cfg.HTTPClient
		if client == nil {
			client = newPooledGitHubHTTPClient()
		}
		githubClient = NewGitHubClient(client, strings.TrimSpace(cfg.GitHubToken))
	}

	fileHandler := cfg.FileHandler
	if fileHandler == nil {
		fileHandler = AtomicFileHandler{}
	}

	analyzer := cfg.Analyzer
	if analyzer == nil {
		var err error
		analyzer, err = llm.NewAnalyzer(cfg.LLM)
		if err != nil {
			return nil, err
		}
	}

	return &PRReviewer{
		githubClient: githubClient,
		fileHandler:  fileHandler,
		config:       cfg,
		analyzer:     analyzer,
	}, nil
}

func NewGitHubClient(client *http.Client, token string) *GitHubClient {
	if client == nil {
		client = newPooledGitHubHTTPClient()
	}
	return &GitHubClient{
		httpClient: client,
		token:      strings.TrimSpace(token),
	}
}

type pullRequestRef struct {
	Owner  string
	Repo   string
	Number int
}

type pullRequestDiscussion struct {
	URL            string
	Title          string
	State          string
	Author         string
	Body           string
	IssueComments  []githubComment
	ReviewComments []githubReviewComment
	Reviews        []githubReview
}

type githubPullRequest struct {
	HTMLURL string `json:"html_url"`
	Title   string `json:"title"`
	State   string `json:"state"`
	Body    string `json:"body"`
	User    struct {
		Login string `json:"login"`
	} `json:"user"`
}

type githubComment struct {
	User struct {
		Login string `json:"login"`
	} `json:"user"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type githubReviewComment struct {
	User struct {
		Login string `json:"login"`
	} `json:"user"`
	Path      string    `json:"path"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type githubReview struct {
	User struct {
		Login string `json:"login"`
	} `json:"user"`
	State       string    `json:"state"`
	Body        string    `json:"body"`
	SubmittedAt time.Time `json:"submitted_at"`
}

func (r *PRReviewer) WriteReviewPlanMarkdown(ctx context.Context, githubLink, outputPath string) error {
	if strings.TrimSpace(outputPath) == "" {
		return fmt.Errorf("output path is required")
	}
	if r == nil {
		return fmt.Errorf("PR reviewer is required")
	}

	pr, err := parsePullRequestURL(githubLink)
	if err != nil {
		return err
	}

	discussion, err := r.githubClient.FetchPullRequestDiscussion(ctx, pr)
	if err != nil {
		return err
	}

	summary, err := summarizePullRequestDiscussion(ctx, r.analyzer, discussion)
	if err != nil {
		return err
	}

	if err := r.fileHandler.Replace(outputPath, []byte(summary)); err != nil {
		return fmt.Errorf("write PR summary: %w", err)
	}

	return nil
}

func (r *PRReviewer) WriteReviewPlanMarkdownToDefaultFile(ctx context.Context, githubLink string) (string, error) {
	pr, err := parsePullRequestURL(githubLink)
	if err != nil {
		return "", err
	}

	outputDir := strings.TrimSpace(r.config.OutputDir)
	if outputDir == "" {
		outputDir = "reports"
	}
	outputPath := filepath.Join(outputDir, fmt.Sprintf("pr-review-%s-%s-%d.md", sanitizeFilePart(pr.Owner), sanitizeFilePart(pr.Repo), pr.Number))
	if err := r.WriteReviewPlanMarkdown(ctx, githubLink, outputPath); err != nil {
		return "", err
	}

	return outputPath, nil
}

func parsePullRequestURL(rawURL string) (pullRequestRef, error) {
	normalizedURL, err := NormalizeGitHubURL(rawURL)
	if err != nil {
		return pullRequestRef{}, err
	}

	parsed, err := url.Parse(normalizedURL)
	if err != nil {
		return pullRequestRef{}, fmt.Errorf("%w: malformed url", ErrInvalidPullRequestURL)
	}

	parts := strings.Split(strings.Trim(parsed.Path, "/"), "/")
	if len(parts) < 4 || parts[2] != "pull" {
		return pullRequestRef{}, fmt.Errorf("%w: expected /owner/repo/pull/number", ErrInvalidPullRequestURL)
	}

	number, err := strconv.Atoi(parts[3])
	if err != nil || number < 1 {
		return pullRequestRef{}, fmt.Errorf("%w: invalid pull request number", ErrInvalidPullRequestURL)
	}

	return pullRequestRef{
		Owner:  parts[0],
		Repo:   parts[1],
		Number: number,
	}, nil
}

func sanitizeFilePart(value string) string {
	value = strings.TrimSpace(value)
	var b strings.Builder
	for _, r := range value {
		switch {
		case r >= 'a' && r <= 'z':
			b.WriteRune(r)
		case r >= 'A' && r <= 'Z':
			b.WriteRune(r)
		case r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == '-' || r == '_' || r == '.':
			b.WriteRune(r)
		default:
			b.WriteByte('-')
		}
	}
	if b.Len() == 0 {
		return "unknown"
	}
	return b.String()
}

func (c *GitHubClient) FetchPullRequestDiscussion(ctx context.Context, pr pullRequestRef) (pullRequestDiscussion, error) {
	if c == nil {
		return pullRequestDiscussion{}, fmt.Errorf("github client is required")
	}
	basePath := fmt.Sprintf("https://api.github.com/repos/%s/%s", url.PathEscape(pr.Owner), url.PathEscape(pr.Repo))

	var pull githubPullRequest
	if err := c.get(ctx, fmt.Sprintf("%s/pulls/%d", basePath, pr.Number), &pull); err != nil {
		return pullRequestDiscussion{}, err
	}

	var issueComments []githubComment
	if err := c.get(ctx, fmt.Sprintf("%s/issues/%d/comments?per_page=100", basePath, pr.Number), &issueComments); err != nil {
		return pullRequestDiscussion{}, err
	}

	var reviewComments []githubReviewComment
	if err := c.get(ctx, fmt.Sprintf("%s/pulls/%d/comments?per_page=100", basePath, pr.Number), &reviewComments); err != nil {
		return pullRequestDiscussion{}, err
	}

	var reviews []githubReview
	if err := c.get(ctx, fmt.Sprintf("%s/pulls/%d/reviews?per_page=100", basePath, pr.Number), &reviews); err != nil {
		return pullRequestDiscussion{}, err
	}

	return pullRequestDiscussion{
		URL:            pull.HTMLURL,
		Title:          pull.Title,
		State:          pull.State,
		Author:         pull.User.Login,
		Body:           pull.Body,
		IssueComments:  issueComments,
		ReviewComments: reviewComments,
		Reviews:        reviews,
	}, nil
}

func (c *GitHubClient) get(ctx context.Context, requestURL string, target any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("github request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("github request failed: status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		return fmt.Errorf("decode github response: %w", err)
	}

	return nil
}

func summarizePullRequestDiscussion(ctx context.Context, analyzer llm.Analyzer, discussion pullRequestDiscussion) (string, error) {
	prompt := buildPullRequestSummaryPrompt(discussion)
	return analyzer.AnalyzePRandReviewPlan(ctx, prompt)
}

func buildPullRequestSummaryPrompt(discussion pullRequestDiscussion) string {
	var b strings.Builder
	b.WriteString("Review this GitHub pull request discussion and summarize what is happening.\n")
	b.WriteString("Write concise Markdown with: Overview, Key discussion points, Fix plan, Open discussion or questions, and Current status.\n")
	b.WriteString("For every clearly actionable request, include a concrete fix plan with execution steps. Leave unclear, disputed, or unanswered items in Open discussion or questions instead of pretending they are ready to execute.\n")
	b.WriteString("Focus on the discussion, reviewer feedback, disagreements, and unresolved questions. Do not invent facts.\n\n")
	fmt.Fprintf(&b, "PR: %s\nTitle: %s\nState: %s\nAuthor: %s\n\n", discussion.URL, discussion.Title, discussion.State, discussion.Author)
	if strings.TrimSpace(discussion.Body) != "" {
		fmt.Fprintf(&b, "Description:\n%s\n\n", truncateForPrompt(discussion.Body))
	}

	b.WriteString("Issue comments:\n")
	for _, comment := range discussion.IssueComments {
		fmt.Fprintf(&b, "- %s at %s: %s\n", comment.User.Login, comment.CreatedAt.Format(time.RFC3339), truncateForPrompt(comment.Body))
	}
	if len(discussion.IssueComments) == 0 {
		b.WriteString("- none\n")
	}

	b.WriteString("\nReview summaries:\n")
	for _, review := range discussion.Reviews {
		fmt.Fprintf(&b, "- %s %s at %s: %s\n", review.User.Login, review.State, review.SubmittedAt.Format(time.RFC3339), truncateForPrompt(review.Body))
	}
	if len(discussion.Reviews) == 0 {
		b.WriteString("- none\n")
	}

	b.WriteString("\nReview comments:\n")
	for _, comment := range discussion.ReviewComments {
		fmt.Fprintf(&b, "- %s on %s at %s: %s\n", comment.User.Login, comment.Path, comment.CreatedAt.Format(time.RFC3339), truncateForPrompt(comment.Body))
	}
	if len(discussion.ReviewComments) == 0 {
		b.WriteString("- none\n")
	}

	return b.String()
}

func truncateForPrompt(value string) string {
	value = strings.TrimSpace(value)
	const limit = 4000
	if len(value) <= limit {
		return value
	}
	return value[:limit] + "\n[truncated]"
}

func newPooledGitHubHTTPClient() *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			MaxIdleConns:        10,
			MaxIdleConnsPerHost: 10,
			MaxConnsPerHost:     10,
			IdleConnTimeout:     90 * time.Second,
		},
	}
}

func (AtomicFileHandler) Replace(path string, data []byte) error {
	return replaceFile(path, data)
}

func replaceFile(path string, data []byte) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create summary directory: %w", err)
	}

	tmpFile, err := os.CreateTemp(dir, "."+filepath.Base(path)+".tmp-*")
	if err != nil {
		return err
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)

	if _, err := tmpFile.Write(data); err != nil {
		_ = tmpFile.Close()
		return err
	}
	if err := tmpFile.Close(); err != nil {
		return err
	}

	return os.Rename(tmpPath, path)
}

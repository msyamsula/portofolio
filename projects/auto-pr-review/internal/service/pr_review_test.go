package service

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

type analyzerFunc func(context.Context, string) (string, error)

func (f analyzerFunc) AnalyzePRandReviewPlan(ctx context.Context, prompt string) (string, error) {
	return f(ctx, prompt)
}

type fileHandlerFunc func(string, []byte) error

func (f fileHandlerFunc) Replace(path string, data []byte) error {
	return f(path, data)
}

func TestParsePullRequestURL(t *testing.T) {
	t.Parallel()

	got, err := parsePullRequestURL("https://github.com/octocat/Hello-World/pull/123")
	if err != nil {
		t.Fatalf("parse pull request url: %v", err)
	}
	if got.Owner != "octocat" {
		t.Fatalf("owner mismatch: got %q", got.Owner)
	}
	if got.Repo != "Hello-World" {
		t.Fatalf("repo mismatch: got %q", got.Repo)
	}
	if got.Number != 123 {
		t.Fatalf("number mismatch: got %d", got.Number)
	}
}

func TestParsePullRequestURLRejectsNonPullRequest(t *testing.T) {
	t.Parallel()

	_, err := parsePullRequestURL("https://github.com/octocat/Hello-World")
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestSanitizeFilePart(t *testing.T) {
	t.Parallel()

	got := sanitizeFilePart("hello/world name")
	if got != "hello-world-name" {
		t.Fatalf("sanitized file part mismatch: got %q", got)
	}
}

func TestBuildPullRequestSummaryPromptRequestsFixPlanAndOpenQuestions(t *testing.T) {
	t.Parallel()

	prompt := buildPullRequestSummaryPrompt(pullRequestDiscussion{})
	if !strings.Contains(prompt, "Fix plan") {
		t.Fatalf("prompt does not request fix plan: %q", prompt)
	}
	if !strings.Contains(prompt, "For every clearly actionable request, include a concrete fix plan with execution steps") {
		t.Fatalf("prompt does not request executable fix steps: %q", prompt)
	}
	if !strings.Contains(prompt, "Leave unclear, disputed, or unanswered items in Open discussion or questions") {
		t.Fatalf("prompt does not preserve open questions: %q", prompt)
	}
}

func TestPRReviewerWriteReviewPlanMarkdown(t *testing.T) {
	t.Parallel()

	client := &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			switch {
			case req.URL.Host == "api.github.com" && strings.HasSuffix(req.URL.Path, "/pulls/123"):
				return jsonResponse(200, `{"html_url":"https://github.com/octocat/Hello-World/pull/123","title":"Add feature","state":"open","body":"PR body","user":{"login":"octocat"}}`), nil
			case req.URL.Host == "api.github.com" && strings.HasSuffix(req.URL.Path, "/issues/123/comments"):
				return jsonResponse(200, `[{"user":{"login":"reviewer"},"body":"Please clarify the behavior.","created_at":"2026-06-13T00:00:00Z","updated_at":"2026-06-13T00:00:00Z"}]`), nil
			case req.URL.Host == "api.github.com" && strings.HasSuffix(req.URL.Path, "/pulls/123/comments"):
				return jsonResponse(200, `[]`), nil
			case req.URL.Host == "api.github.com" && strings.HasSuffix(req.URL.Path, "/pulls/123/reviews"):
				return jsonResponse(200, `[]`), nil
			default:
				return nil, fmt.Errorf("unexpected request: %s %s", req.Method, req.URL.String())
			}
		}),
	}

	var analyzerCalls int
	analyzer := analyzerFunc(func(ctx context.Context, prompt string) (string, error) {
		analyzerCalls++
		if !strings.Contains(prompt, "Please clarify the behavior.") {
			return "", fmt.Errorf("prompt missing discussion: %q", prompt)
		}
		return "# Summary\n\nThe PR needs clarification.\n", nil
	})

	outputPath := filepath.Join(t.TempDir(), "summary.md")
	if err := os.WriteFile(outputPath, []byte("old summary"), 0o644); err != nil {
		t.Fatalf("write old summary: %v", err)
	}
	var fileHandlerCalls int
	fileHandler := fileHandlerFunc(func(path string, data []byte) error {
		fileHandlerCalls++
		if path != outputPath {
			return fmt.Errorf("output path mismatch: got %q want %q", path, outputPath)
		}
		return AtomicFileHandler{}.Replace(path, data)
	})

	reviewer, err := NewPRReviewer(PRDiscussionReviewConfig{
		Analyzer:    analyzer,
		HTTPClient:  client,
		FileHandler: fileHandler,
	})
	if err != nil {
		t.Fatalf("new reviewer: %v", err)
	}

	err = reviewer.WriteReviewPlanMarkdown(context.Background(), "https://github.com/octocat/Hello-World/pull/123", outputPath)
	if err != nil {
		t.Fatalf("review pull request discussion: %v", err)
	}

	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("read output file: %v", err)
	}
	if got := string(data); !strings.Contains(got, "The PR needs clarification.") {
		t.Fatalf("summary mismatch: got %q", got)
	}
	if strings.Contains(string(data), "old summary") {
		t.Fatalf("expected output file to be replaced, got %q", string(data))
	}
	if analyzerCalls != 1 {
		t.Fatalf("analyzer call count mismatch: got %d want 1", analyzerCalls)
	}
	if fileHandlerCalls != 1 {
		t.Fatalf("file handler call count mismatch: got %d want 1", fileHandlerCalls)
	}
}

func jsonResponse(statusCode int, body string) *http.Response {
	return &http.Response{
		StatusCode: statusCode,
		Header:     make(http.Header),
		Body:       ioNopCloser{strings.NewReader(body)},
	}
}

type ioNopCloser struct {
	*strings.Reader
}

func (c ioNopCloser) Close() error {
	return nil
}

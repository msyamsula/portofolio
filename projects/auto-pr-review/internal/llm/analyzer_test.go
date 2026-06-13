package llm

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestClaudeAnalyzerRetriesAndReturnsText(t *testing.T) {
	t.Parallel()

	var calls int
	analyzer := ClaudeAnalyzer{
		retryingAnalyzer: retryingAnalyzer{
			client: &http.Client{
				Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
					calls++
					if req.URL.Host != "api.anthropic.com" || req.URL.Path != "/v1/messages" {
						return nil, fmt.Errorf("unexpected request: %s", req.URL.String())
					}
					if req.Header.Get("x-api-key") != "test-key" {
						return nil, fmt.Errorf("api key header mismatch: %q", req.Header.Get("x-api-key"))
					}
					if calls == 1 {
						return jsonResponse(529, `{"error":"overloaded"}`), nil
					}
					return jsonResponse(200, `{"content":[{"type":"text","text":"# Summary\n\nUse Claude."}]}`), nil
				}),
			},
			request: func(ctx context.Context, client *http.Client, prompt string) (string, bool, error) {
				return requestClaudeSummary(ctx, client, "test-key", "claude-test", prompt)
			},
			requestTimeout: time.Second,
			maxRetries:     1,
			retryBackoff:   time.Millisecond,
		},
	}

	got, err := analyzer.AnalyzePRandReviewPlan(context.Background(), "prompt")
	if err != nil {
		t.Fatalf("analyze: %v", err)
	}
	if !strings.Contains(got, "Use Claude.") {
		t.Fatalf("summary mismatch: %q", got)
	}
	if calls != 2 {
		t.Fatalf("call count mismatch: got %d want 2", calls)
	}
}

func TestChatGPTAnalyzerRetriesAndReturnsOutputText(t *testing.T) {
	t.Parallel()

	var calls int
	analyzer := ChatGPTAnalyzer{
		retryingAnalyzer: retryingAnalyzer{
			client: &http.Client{
				Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
					calls++
					if req.URL.Host != "api.openai.com" || req.URL.Path != "/v1/responses" {
						return nil, fmt.Errorf("unexpected request: %s", req.URL.String())
					}
					if req.Header.Get("Authorization") != "Bearer test-key" {
						return nil, fmt.Errorf("authorization header mismatch: %q", req.Header.Get("Authorization"))
					}
					if calls == 1 {
						return jsonResponse(429, `{"error":"rate limited"}`), nil
					}
					return jsonResponse(200, `{"output_text":"# Summary\n\nUse ChatGPT."}`), nil
				}),
			},
			request: func(ctx context.Context, client *http.Client, prompt string) (string, bool, error) {
				return requestChatGPTSummary(ctx, client, "test-key", "gpt-test", prompt)
			},
			requestTimeout: time.Second,
			maxRetries:     1,
			retryBackoff:   time.Millisecond,
		},
	}

	got, err := analyzer.AnalyzePRandReviewPlan(context.Background(), "prompt")
	if err != nil {
		t.Fatalf("analyze: %v", err)
	}
	if !strings.Contains(got, "Use ChatGPT.") {
		t.Fatalf("summary mismatch: %q", got)
	}
	if calls != 2 {
		t.Fatalf("call count mismatch: got %d want 2", calls)
	}
}

func TestNewAnalyzerSelectsSingleProvider(t *testing.T) {
	t.Parallel()

	claudeAnalyzer, err := NewAnalyzer(Config{Provider: ProviderClaude, APIKey: "claude-key"})
	if err != nil {
		t.Fatalf("new claude analyzer: %v", err)
	}
	if _, ok := claudeAnalyzer.(*ClaudeAnalyzer); !ok {
		t.Fatalf("expected ClaudeAnalyzer, got %T", claudeAnalyzer)
	}

	chatGPTAnalyzer, err := NewAnalyzer(Config{Provider: ProviderChatGPT, APIKey: "openai-key"})
	if err != nil {
		t.Fatalf("new chatgpt analyzer: %v", err)
	}
	if _, ok := chatGPTAnalyzer.(*ChatGPTAnalyzer); !ok {
		t.Fatalf("expected ChatGPTAnalyzer, got %T", chatGPTAnalyzer)
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

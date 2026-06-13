package llm

import (
	"context"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

type retryingAnalyzer struct {
	client         *http.Client
	request        func(context.Context, *http.Client, string) (string, bool, error)
	requestTimeout time.Duration
	maxRetries     int
	retryBackoff   time.Duration
}

func newRetryingAnalyzer(cfg Config, request func(context.Context, *http.Client, string) (string, bool, error)) retryingAnalyzer {
	timeout := cfg.RequestTimeout
	if timeout <= 0 {
		timeout = time.Minute
	}
	maxRetries := cfg.MaxRetries
	if maxRetries < 0 {
		maxRetries = 0
	}
	backoff := cfg.RetryBackoff
	if backoff <= 0 {
		backoff = 500 * time.Millisecond
	}

	return retryingAnalyzer{
		client:         newPooledHTTPClient(cfg),
		request:        request,
		requestTimeout: timeout,
		maxRetries:     maxRetries,
		retryBackoff:   backoff,
	}
}

func (a retryingAnalyzer) AnalyzePRandReviewPlan(ctx context.Context, prompt string) (string, error) {
	requestCtx, cancel := context.WithTimeout(ctx, a.requestTimeout)
	defer cancel()

	var lastErr error
	for attempt := 0; attempt <= a.maxRetries; attempt++ {
		if attempt > 0 {
			if err := sleepWithBackoffJitter(requestCtx, a.retryBackoff, attempt); err != nil {
				return "", err
			}
		}

		summary, retryable, err := a.request(requestCtx, a.client, prompt)
		if err == nil {
			return ensureTrailingNewline(summary), nil
		}
		lastErr = err
		if !retryable {
			return "", err
		}
	}

	return "", lastErr
}

func isRetryableStatus(statusCode int) bool {
	return statusCode == http.StatusTooManyRequests || statusCode >= 500
}

func ensureTrailingNewline(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	return value + "\n"
}

func sleepWithBackoffJitter(ctx context.Context, base time.Duration, attempt int) error {
	if attempt < 1 {
		attempt = 1
	}
	delay := base << (attempt - 1)
	jitterMax := delay / 2
	if jitterMax <= 0 {
		jitterMax = time.Millisecond
	}
	wait := delay + time.Duration(rand.Int63n(int64(jitterMax)+1))

	timer := time.NewTimer(wait)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

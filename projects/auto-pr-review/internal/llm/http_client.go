package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

func newPooledHTTPClient(cfg Config) *http.Client {
	maxIdleConns := cfg.MaxIdleConns
	if maxIdleConns <= 0 {
		maxIdleConns = 10
	}
	maxIdleConnsPerHost := cfg.MaxIdleConnsPerHost
	if maxIdleConnsPerHost <= 0 {
		maxIdleConnsPerHost = 10
	}
	maxConnsPerHost := cfg.MaxConnsPerHost
	if maxConnsPerHost <= 0 {
		maxConnsPerHost = 10
	}

	return &http.Client{
		Transport: &http.Transport{
			MaxIdleConns:        maxIdleConns,
			MaxIdleConnsPerHost: maxIdleConnsPerHost,
			MaxConnsPerHost:     maxConnsPerHost,
			IdleConnTimeout:     90 * time.Second,
		},
	}
}

func postJSON(ctx context.Context, client *http.Client, requestURL string, headers map[string]string, payload any, target any, provider string) (int, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return 0, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, requestURL, bytes.NewReader(data))
	if err != nil {
		return 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("%s request failed: %w", provider, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return resp.StatusCode, fmt.Errorf("%s request failed: status %d: %s", provider, resp.StatusCode, strings.TrimSpace(string(body)))
	}

	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		return resp.StatusCode, fmt.Errorf("decode %s response: %w", provider, err)
	}

	return resp.StatusCode, nil
}

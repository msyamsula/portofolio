package llm

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

const defaultClaudeModel = "claude-3-5-haiku-latest"

type ClaudeAnalyzer struct {
	retryingAnalyzer
}

type claudeMessagesRequest struct {
	Model     string          `json:"model"`
	MaxTokens int             `json:"max_tokens"`
	Messages  []claudeMessage `json:"messages"`
}

type claudeMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type claudeMessagesResponse struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
}

func NewClaudeAnalyzer(cfg Config) (*ClaudeAnalyzer, error) {
	apiKey := strings.TrimSpace(cfg.APIKey)
	if apiKey == "" {
		return nil, fmt.Errorf("ANTHROPIC_API_KEY is required")
	}
	model := strings.TrimSpace(cfg.Model)
	if model == "" {
		model = defaultClaudeModel
	}

	return &ClaudeAnalyzer{
		retryingAnalyzer: newRetryingAnalyzer(cfg, func(ctx context.Context, client *http.Client, prompt string) (string, bool, error) {
			return requestClaudeSummary(ctx, client, apiKey, model, prompt)
		}),
	}, nil
}

func requestClaudeSummary(ctx context.Context, client *http.Client, apiKey, model, prompt string) (string, bool, error) {
	var result claudeMessagesResponse
	statusCode, err := postJSON(ctx, client, "https://api.anthropic.com/v1/messages", map[string]string{
		"x-api-key":         apiKey,
		"anthropic-version": "2023-06-01",
	}, claudeMessagesRequest{
		Model:     model,
		MaxTokens: 2048,
		Messages: []claudeMessage{
			{Role: "user", Content: prompt},
		},
	}, &result, ProviderClaude)
	if err != nil {
		return "", statusCode == 0 || isRetryableStatus(statusCode), err
	}

	summary := strings.TrimSpace(extractClaudeText(result))
	if summary == "" {
		return "", false, fmt.Errorf("claude response did not contain text")
	}

	return summary, false, nil
}

func extractClaudeText(result claudeMessagesResponse) string {
	var parts []string
	for _, content := range result.Content {
		if content.Type == "text" && strings.TrimSpace(content.Text) != "" {
			parts = append(parts, content.Text)
		}
	}
	return strings.Join(parts, "\n")
}

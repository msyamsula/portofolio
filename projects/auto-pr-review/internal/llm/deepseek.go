package llm

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

const defaultDeepSeekModel = "deepseek-v4-pro"

type DeepSeekAnalyzer struct {
	retryingAnalyzer
}

type deepSeekRequest struct {
	Model    string            `json:"model"`
	Messages []deepSeekMessage `json:"messages"`
}

type deepSeekMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type deepSeekResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func NewDeepSeekAnalyzer(cfg Config) (*DeepSeekAnalyzer, error) {
	apiKey := strings.TrimSpace(cfg.APIKey)
	if apiKey == "" {
		return nil, fmt.Errorf("DEEPSEEK_API_KEY is required")
	}
	model := strings.TrimSpace(cfg.Model)
	if model == "" {
		model = defaultDeepSeekModel
	}

	return &DeepSeekAnalyzer{
		retryingAnalyzer: newRetryingAnalyzer(cfg, func(ctx context.Context, client *http.Client, prompt string) (string, bool, error) {
			return requestDeepSeekSummary(ctx, client, apiKey, model, prompt)
		}),
	}, nil
}

func requestDeepSeekSummary(ctx context.Context, client *http.Client, apiKey, model, prompt string) (string, bool, error) {
	var result deepSeekResponse
	statusCode, err := postJSON(ctx, client, "https://api.deepseek.com/v1/chat/completions", map[string]string{
		"Authorization": "Bearer " + apiKey,
	}, deepSeekRequest{
		Model: model,
		Messages: []deepSeekMessage{
			{Role: "user", Content: prompt},
		},
	}, &result, ProviderDeepSeek)
	if err != nil {
		return "", statusCode == 0 || isRetryableStatus(statusCode), err
	}

	summary := strings.TrimSpace(extractDeepSeekText(result))
	if summary == "" {
		return "", false, fmt.Errorf("deepseek response did not contain text")
	}

	return summary, false, nil
}

func extractDeepSeekText(result deepSeekResponse) string {
	var parts []string
	for _, choice := range result.Choices {
		if strings.TrimSpace(choice.Message.Content) != "" {
			parts = append(parts, choice.Message.Content)
		}
	}
	return strings.Join(parts, "\n")
}

package llm

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

const defaultChatGPTModel = "gpt-4o-mini"

type ChatGPTAnalyzer struct {
	retryingAnalyzer
}

type chatGPTResponsesRequest struct {
	Model string `json:"model"`
	Input string `json:"input"`
}

type chatGPTResponsesResponse struct {
	OutputText string `json:"output_text"`
	Output     []struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	} `json:"output"`
}

func NewChatGPTAnalyzer(cfg Config) (*ChatGPTAnalyzer, error) {
	apiKey := strings.TrimSpace(cfg.APIKey)
	if apiKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY is required")
	}
	model := strings.TrimSpace(cfg.Model)
	if model == "" {
		model = defaultChatGPTModel
	}

	return &ChatGPTAnalyzer{
		retryingAnalyzer: newRetryingAnalyzer(cfg, func(ctx context.Context, client *http.Client, prompt string) (string, bool, error) {
			return requestChatGPTSummary(ctx, client, apiKey, model, prompt)
		}),
	}, nil
}

func requestChatGPTSummary(ctx context.Context, client *http.Client, apiKey, model, prompt string) (string, bool, error) {
	var result chatGPTResponsesResponse
	statusCode, err := postJSON(ctx, client, "https://api.openai.com/v1/responses", map[string]string{
		"Authorization": "Bearer " + apiKey,
	}, chatGPTResponsesRequest{
		Model: model,
		Input: prompt,
	}, &result, ProviderChatGPT)
	if err != nil {
		return "", statusCode == 0 || isRetryableStatus(statusCode), err
	}

	summary := strings.TrimSpace(extractChatGPTOutputText(result))
	if summary == "" {
		return "", false, fmt.Errorf("chatgpt response did not contain output text")
	}

	return summary, false, nil
}

func extractChatGPTOutputText(result chatGPTResponsesResponse) string {
	if strings.TrimSpace(result.OutputText) != "" {
		return result.OutputText
	}

	var parts []string
	for _, output := range result.Output {
		for _, content := range output.Content {
			if content.Type == "output_text" && strings.TrimSpace(content.Text) != "" {
				parts = append(parts, content.Text)
			}
		}
	}

	return strings.Join(parts, "\n")
}

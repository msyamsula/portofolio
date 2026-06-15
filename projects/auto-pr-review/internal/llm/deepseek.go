package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

const defaultDeepSeekModel = "deepseek-v4-pro"

type DeepSeekAnalyzer struct {
	retryingAnalyzer
	apiKey string
	model  string
}

type deepSeekRequest struct {
	Model    string            `json:"model"`
	Messages []deepSeekMessage `json:"messages"`
}

type deepSeekMessage struct {
	Role       string               `json:"role"`
	Content    string               `json:"content,omitempty"`
	ToolCalls  []deepSeekToolCall   `json:"tool_calls,omitempty"`
	ToolCallID string               `json:"tool_call_id,omitempty"`
}

type deepSeekToolCall struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Function struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	} `json:"function"`
}

type deepSeekTool struct {
	Type     string `json:"type"`
	Function struct {
		Name        string         `json:"name"`
		Description string         `json:"description"`
		Parameters  map[string]any `json:"parameters,omitempty"`
	} `json:"function"`
}

type deepSeekChatRequest struct {
	Model    string            `json:"model"`
	Messages []deepSeekMessage `json:"messages"`
	Tools    []deepSeekTool    `json:"tools,omitempty"`
}

type deepSeekChatResponse struct {
	Choices []struct {
		Message struct {
			Content   string             `json:"content"`
			ToolCalls []deepSeekToolCall `json:"tool_calls"`
		} `json:"message"`
	} `json:"choices"`
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
		apiKey: apiKey,
		model:  model,
	}, nil
}

func (a *DeepSeekAnalyzer) Chat(ctx context.Context, messages []Message, tools []Tool) (*Response, error) {
	chatCtx, cancel := context.WithTimeout(ctx, a.requestTimeout)
	defer cancel()
	_ = ctx
	_ = chatCtx

	payload := deepSeekChatRequest{
		Model: a.model,
	}
	for _, t := range tools {
		payload.Tools = append(payload.Tools, deepSeekTool{
			Type: "function",
			Function: struct {
				Name        string         `json:"name"`
				Description string         `json:"description"`
				Parameters  map[string]any `json:"parameters,omitempty"`
			}{Name: t.Name, Description: t.Description, Parameters: t.Parameters},
		})
	}
	for _, m := range messages {
		dsm := deepSeekMessage{Role: m.Role, Content: m.Content, ToolCallID: m.ToolCallID}
		for _, tc := range m.ToolCalls {
			dsm.ToolCalls = append(dsm.ToolCalls, deepSeekToolCall{
				ID:   tc.ID,
				Type: "function",
				Function: struct {
					Name      string `json:"name"`
					Arguments string `json:"arguments"`
				}{Name: tc.Name, Arguments: mustMarshalJSON(tc.Args)},
			})
		}
		payload.Messages = append(payload.Messages, dsm)
	}

	var result deepSeekChatResponse
	statusCode, err := postJSON(chatCtx, a.client, "https://api.deepseek.com/v1/chat/completions", map[string]string{
		"Authorization": "Bearer " + a.apiKey,
	}, payload, &result, ProviderDeepSeek)
	if err != nil {
		return nil, fmt.Errorf("deepseek chat: %w", err)
	}

	return parseDeepSeekChatResponse(statusCode, &result)
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

func parseDeepSeekChatResponse(statusCode int, result *deepSeekChatResponse) (*Response, error) {
	if len(result.Choices) == 0 {
		return nil, fmt.Errorf("deepseek response has no choices")
	}

	msg := result.Choices[0].Message
	resp := &Response{Text: strings.TrimSpace(msg.Content)}

	for _, tc := range msg.ToolCalls {
		var args map[string]any
		if err := json.Unmarshal([]byte(tc.Function.Arguments), &args); err != nil {
			return nil, fmt.Errorf("parse tool call arguments: %w", err)
		}
		resp.ToolCalls = append(resp.ToolCalls, ToolCall{
			ID:   tc.ID,
			Name: tc.Function.Name,
			Args: args,
		})
	}

	return resp, nil
}

var _ Agent = (*DeepSeekAnalyzer)(nil)

func mustMarshalJSON(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}

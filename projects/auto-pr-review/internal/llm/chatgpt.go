package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

const defaultChatGPTModel = "gpt-4o-mini"

type ChatGPTAnalyzer struct {
	retryingAnalyzer
	apiKey string
	model  string
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
		apiKey: apiKey,
		model:  model,
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

func (a *ChatGPTAnalyzer) Chat(ctx context.Context, messages []Message, tools []Tool) (*Response, error) {
	chatCtx, cancel := context.WithTimeout(ctx, a.requestTimeout)
	defer cancel()

	req := chatGPTChatRequest{
		Model:     a.model,
		MaxTokens: 4096,
	}

	for _, t := range tools {
		req.Tools = append(req.Tools, chatGPTChatTool{
			Type: "function",
			Function: struct {
				Name        string         `json:"name"`
				Description string         `json:"description"`
				Parameters  map[string]any `json:"parameters,omitempty"`
			}{Name: t.Name, Description: t.Description, Parameters: t.Parameters},
		})
	}
	for _, m := range messages {
		cm := chatGPTChatMessage{Role: m.Role, Content: m.Content, ToolCallID: m.ToolCallID}
		for _, tc := range m.ToolCalls {
			cm.ToolCalls = append(cm.ToolCalls, chatGPTChatToolCall{
				ID:   tc.ID,
				Type: "function",
				Function: struct {
					Name      string `json:"name"`
					Arguments string `json:"arguments"`
				}{Name: tc.Name, Arguments: mustMarshalJSON(tc.Args)},
			})
		}
		req.Messages = append(req.Messages, cm)
	}

	var lastErr error
	for attempt := 0; attempt <= a.maxRetries; attempt++ {
		if attempt > 0 {
			if err := sleepWithBackoffJitter(chatCtx, a.retryBackoff, attempt); err != nil {
				return nil, err
			}
		}

		var result chatGPTChatResponse
		statusCode, err := postJSON(chatCtx, a.client, "https://api.openai.com/v1/chat/completions", map[string]string{
			"Authorization": "Bearer " + a.apiKey,
		}, req, &result, ProviderChatGPT)
		if err == nil {
			return parseChatGPTChatResponse(&result)
		}
		lastErr = err
		if !(statusCode == 0 || isRetryableStatus(statusCode)) {
			return nil, fmt.Errorf("chatgpt chat: %w", err)
		}
	}
	return nil, fmt.Errorf("chatgpt chat after %d retries: %w", a.maxRetries, lastErr)
}

func parseChatGPTChatResponse(result *chatGPTChatResponse) (*Response, error) {
	if len(result.Choices) == 0 {
		return nil, fmt.Errorf("chatgpt response has no choices")
	}

	msg := result.Choices[0].Message
	resp := &Response{Text: strings.TrimSpace(msg.Content)}

	for _, tc := range msg.ToolCalls {
		var args map[string]any
		if err := json.Unmarshal([]byte(tc.Function.Arguments), &args); err != nil {
			return nil, fmt.Errorf("parse chatgpt tool call arguments: %w", err)
		}
		resp.ToolCalls = append(resp.ToolCalls, ToolCall{
			ID:   tc.ID,
			Name: tc.Function.Name,
			Args: args,
		})
	}

	return resp, nil
}

type chatGPTChatRequest struct {
	Model     string               `json:"model"`
	MaxTokens int                  `json:"max_completion_tokens"`
	Messages  []chatGPTChatMessage `json:"messages"`
	Tools     []chatGPTChatTool    `json:"tools,omitempty"`
}

type chatGPTChatMessage struct {
	Role       string                `json:"role"`
	Content    string                `json:"content,omitempty"`
	ToolCalls  []chatGPTChatToolCall `json:"tool_calls,omitempty"`
	ToolCallID string                `json:"tool_call_id,omitempty"`
}

type chatGPTChatTool struct {
	Type     string `json:"type"`
	Function struct {
		Name        string         `json:"name"`
		Description string         `json:"description"`
		Parameters  map[string]any `json:"parameters,omitempty"`
	} `json:"function"`
}

type chatGPTChatToolCall struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Function struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	} `json:"function"`
}

type chatGPTChatResponse struct {
	Choices []struct {
		Message struct {
			Content   string                `json:"content"`
			ToolCalls []chatGPTChatToolCall `json:"tool_calls"`
		} `json:"message"`
	} `json:"choices"`
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

var _ Agent = (*ChatGPTAnalyzer)(nil)

package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

const defaultClaudeModel = "claude-sonnet-4-6"

type ClaudeAnalyzer struct {
	retryingAnalyzer
	apiKey string
	model  string
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
		apiKey: apiKey,
		model:  model,
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

type claudeChatTool struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	InputSchema map[string]any `json:"input_schema"`
}

type claudeChatRequest struct {
	Model     string              `json:"model"`
	MaxTokens int                 `json:"max_tokens"`
	Messages  []claudeChatMessage `json:"messages"`
	System    string              `json:"system,omitempty"`
	Tools     []claudeChatTool    `json:"tools,omitempty"`
}

type claudeChatMessage struct {
	Role    string          `json:"role"`
	Content json.RawMessage `json:"content"`
}

type claudeChatContentBlock struct {
	Type      string          `json:"type"`
	Text      string          `json:"text,omitempty"`
	ID        string          `json:"id,omitempty"`
	Name      string          `json:"name,omitempty"`
	Input     json.RawMessage `json:"input,omitempty"`
	ToolUseID string          `json:"tool_use_id,omitempty"`
	Content   string          `json:"content,omitempty"`
}

type claudeChatResponse struct {
	Content []claudeChatContentBlock `json:"content"`
}

func (a *ClaudeAnalyzer) Chat(ctx context.Context, messages []Message, tools []Tool) (*Response, error) {
	chatCtx, cancel := context.WithTimeout(ctx, a.requestTimeout)
	defer cancel()

	req := claudeChatRequest{
		Model:     a.model,
		MaxTokens: 4096,
	}

	for _, t := range tools {
		req.Tools = append(req.Tools, claudeChatTool{
			Name:        t.Name,
			Description: t.Description,
			InputSchema: t.Parameters,
		})
	}

	var systemBuf strings.Builder
	var chatMessages []claudeChatMessage

	for _, m := range messages {
		switch m.Role {
		case "system":
			systemBuf.WriteString(m.Content)
			systemBuf.WriteString("\n")
		case "user":
			raw, _ := json.Marshal(m.Content)
			chatMessages = append(chatMessages, claudeChatMessage{Role: "user", Content: raw})
		case "assistant":
			var blocks []claudeChatContentBlock
			if m.Content != "" {
				blocks = append(blocks, claudeChatContentBlock{Type: "text", Text: m.Content})
			}
			for _, tc := range m.ToolCalls {
				blocks = append(blocks, claudeChatContentBlock{
					Type:  "tool_use",
					ID:    tc.ID,
					Name:  tc.Name,
					Input: json.RawMessage(mustMarshalJSON(tc.Args)),
				})
			}
			raw, _ := json.Marshal(blocks)
			chatMessages = append(chatMessages, claudeChatMessage{Role: "assistant", Content: raw})
		case "tool":
			blocks := []claudeChatContentBlock{{
				Type:      "tool_result",
				ToolUseID: m.ToolCallID,
				Content:   m.Content,
			}}
			raw, _ := json.Marshal(blocks)
			chatMessages = append(chatMessages, claudeChatMessage{Role: "user", Content: raw})
		}
	}

	req.Messages = chatMessages
	if sys := strings.TrimSpace(systemBuf.String()); sys != "" {
		req.System = sys
	}

	var lastErr error
	for attempt := 0; attempt <= a.maxRetries; attempt++ {
		if attempt > 0 {
			if err := sleepWithBackoffJitter(chatCtx, a.retryBackoff, attempt); err != nil {
				return nil, err
			}
		}

		var result claudeChatResponse
		statusCode, err := postJSON(chatCtx, a.client, "https://api.anthropic.com/v1/messages", map[string]string{
			"x-api-key":         a.apiKey,
			"anthropic-version": "2023-06-01",
		}, req, &result, ProviderClaude)
		if err == nil {
			return parseClaudeChatResponse(&result)
		}
		lastErr = err
		if !(statusCode == 0 || isRetryableStatus(statusCode)) {
			return nil, fmt.Errorf("claude chat: %w", err)
		}
	}
	return nil, fmt.Errorf("claude chat after %d retries: %w", a.maxRetries, lastErr)
}

func parseClaudeChatResponse(result *claudeChatResponse) (*Response, error) {
	resp := &Response{}
	for _, block := range result.Content {
		switch block.Type {
		case "text":
			if strings.TrimSpace(block.Text) != "" {
				if resp.Text != "" {
					resp.Text += "\n"
				}
				resp.Text += block.Text
			}
		case "tool_use":
			var args map[string]any
			if len(block.Input) > 0 {
				if err := json.Unmarshal(block.Input, &args); err != nil {
					return nil, fmt.Errorf("parse claude tool call arguments: %w", err)
				}
			}
			resp.ToolCalls = append(resp.ToolCalls, ToolCall{
				ID:   block.ID,
				Name: block.Name,
				Args: args,
			})
		}
	}
	resp.Text = strings.TrimSpace(resp.Text)
	return resp, nil
}

var _ Agent = (*ClaudeAnalyzer)(nil)

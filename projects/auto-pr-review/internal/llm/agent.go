package llm

import "context"

type Tool struct {
	Name        string
	Description string
	Parameters  map[string]any
}

type ToolCall struct {
	ID   string
	Name string
	Args map[string]any
}

type Message struct {
	Role       string
	Content    string
	ToolCalls  []ToolCall
	ToolCallID string
}

type Response struct {
	Text      string
	ToolCalls []ToolCall
}

type Agent interface {
	Chat(ctx context.Context, messages []Message, tools []Tool) (*Response, error)
}

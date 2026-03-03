package dto

// Request represents an LLM request
type Request struct {
	// The natural language prompt
	Prompt string

	// System message for context
	SystemPrompt string

	// Optional schema context
	Schema string

	// Optional conversation history
	History []Message

	// Temperature for generation (0.0 to 1.0)
	Temperature float32

	// Maximum tokens to generate
	MaxTokens int
}

// Message represents a conversation message
type Message struct {
	Role    string // "system", "user", "assistant"
	Content string
}

// Response represents an LLM response
type Response struct {
	// Generated SQL or text
	Content string

	// Raw response from the API
	Raw any

	// Model used
	Model string

	// Tokens used
	Usage TokenUsage
}

// TokenUsage represents token usage statistics
type TokenUsage struct {
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
}

// SQLRequest is a specialized request for SQL generation
type SQLRequest struct {
	NaturalLanguage string
	Schema          string
	DatabaseContext string
}

// SQLResponse is the response for SQL generation
type SQLResponse struct {
	SQL         string
	Explanation string
	Confidence  float32
}

package service

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/msyamsula/portofolio/backend-app/infrastructure/llm/dto"
	llmprompt "github.com/msyamsula/portofolio/backend-app/infrastructure/llm/prompt"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/openai/openai-go/packages/param"
)

// llmService implements the Service interface
type llmService struct {
	client *openai.Client
	apiKey string
	model  string
}

// New creates a new LLM service
func New(apiKey string) (Service, error) {
	if apiKey == "" {
		return &llmService{
			apiKey: apiKey,
			model:  "gpt-4o-mini",
		}, nil
	}

	client := openai.NewClient(
		option.WithAPIKey(apiKey),
	)

	return &llmService{
		client: &client,
		apiKey: apiKey,
		model:  "gpt-4o-mini",
	}, nil
}

// SetAPIKey sets the OpenAI API key
func (s *llmService) SetAPIKey(apiKey string) {
	s.apiKey = apiKey
	if apiKey != "" {
		client := openai.NewClient(
			option.WithAPIKey(apiKey),
		)
		s.client = &client
	}
}

// GetAPIKey returns the current API key
func (s *llmService) GetAPIKey() string {
	return s.apiKey
}

// GenerateSQL generates SQL from natural language
func (s *llmService) GenerateSQL(ctx context.Context, req dto.SQLRequest) (dto.SQLResponse, error) {
	if s.client == nil {
		return dto.SQLResponse{}, fmt.Errorf("LLM client not initialized. Please set an API key")
	}

	systemPrompt := llmprompt.BuildSystemPrompt(req.Schema)
	userPrompt := req.NaturalLanguage

	messages := []openai.ChatCompletionMessageParamUnion{
		openai.SystemMessage(systemPrompt),
		openai.UserMessage(userPrompt),
	}

	response, err := s.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages:    messages,
		Model:       s.model,
		Temperature: param.Opt[float64]{Value: 0.3},
		MaxTokens:   param.Opt[int64]{Value: 1000},
	})
	if err != nil {
		return dto.SQLResponse{}, fmt.Errorf("failed to generate SQL: %w", err)
	}

	if len(response.Choices) == 0 {
		return dto.SQLResponse{}, fmt.Errorf("no response from LLM")
	}

	content := response.Choices[0].Message.Content
	sql := extractSQL(content)

	return dto.SQLResponse{
		SQL:         sql,
		Explanation: "",  // Can be populated with a separate explanation call
		Confidence:  0.9, // Simplified confidence
	}, nil
}

// Generate generates text from a prompt
func (s *llmService) Generate(ctx context.Context, req dto.Request) (dto.Response, error) {
	if s.client == nil {
		return dto.Response{}, fmt.Errorf("LLM client not initialized. Please set an API key")
	}

	messages := []openai.ChatCompletionMessageParamUnion{}

	if req.SystemPrompt != "" {
		messages = append(messages, openai.SystemMessage(req.SystemPrompt))
	} else if req.Schema != "" {
		messages = append(messages, openai.SystemMessage(llmprompt.BuildSystemPrompt(req.Schema)))
	}

	// Add history if provided
	for _, msg := range req.History {
		switch strings.ToLower(msg.Role) {
		case "system":
			messages = append(messages, openai.SystemMessage(msg.Content))
		case "user":
			messages = append(messages, openai.UserMessage(msg.Content))
		case "assistant":
			messages = append(messages, openai.AssistantMessage(msg.Content))
		}
	}

	// Add the current prompt
	messages = append(messages, openai.UserMessage(req.Prompt))

	temperature := 0.7
	if req.Temperature > 0 {
		temperature = float64(req.Temperature)
	}

	maxTokens := int64(1000)
	if req.MaxTokens > 0 {
		maxTokens = int64(req.MaxTokens)
	}

	response, err := s.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages:    messages,
		Model:       s.model,
		Temperature: param.Opt[float64]{Value: temperature},
		MaxTokens:   param.Opt[int64]{Value: maxTokens},
	})
	if err != nil {
		return dto.Response{}, fmt.Errorf("failed to generate: %w", err)
	}

	if len(response.Choices) == 0 {
		return dto.Response{}, fmt.Errorf("no response from LLM")
	}

	content := response.Choices[0].Message.Content

	var usage dto.TokenUsage
	usage.PromptTokens = int(response.Usage.PromptTokens)
	usage.CompletionTokens = int(response.Usage.CompletionTokens)
	usage.TotalTokens = int(response.Usage.TotalTokens)

	return dto.Response{
		Content: content,
		Raw:     content,
		Model:   s.model,
		Usage:   usage,
	}, nil
}

// extractSQL extracts SQL code from a response that may contain markdown formatting
func extractSQL(content string) string {
	// Remove markdown code blocks if present
	content = strings.TrimSpace(content)

	// Check for ```sql or ``` code blocks
	sqlBlock := regexp.MustCompile("(?i)```(?:sql)?\\s*([\\s\\S]*?)\\s*```")
	matches := sqlBlock.FindStringSubmatch(content)
	if len(matches) > 1 {
		content = matches[1]
	}

	// Remove leading/trailing whitespace
	content = strings.TrimSpace(content)

	// Remove common SQL comments
	commentPattern := regexp.MustCompile(`--.*?\n`)
	content = commentPattern.ReplaceAllString(content, "")

	return strings.TrimSpace(content)
}

// GetModel returns the current model being used
func (s *llmService) GetModel() string {
	return s.model
}

// SetModel sets the model to use
func (s *llmService) SetModel(model string) {
	s.model = model
}

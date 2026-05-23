package service

import (
	"context"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

// llmService implements the Service interface
type llmService struct {
	modelClient *openai.Client
	model       string
}

// NewOpenAI creates a new llmService with an OpenAI client
func NewOpenAI(ctx context.Context, apiKey string, model string) (*llmService, error) {
	client := openai.NewClient(option.WithAPIKey(apiKey))

	return &llmService{
		modelClient: &client,
		model:       model,
	}, nil
}

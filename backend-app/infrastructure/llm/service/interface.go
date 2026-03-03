package service

import (
	"context"

	"github.com/msyamsula/portofolio/backend-app/infrastructure/llm/dto"
)

// Service defines the interface for LLM operations
//
//go:generate mockgen -source=interface.go -destination=../../../mock/llm_service_mock.go -package=mock -mock_names Service=MockLLMService
type Service interface {
	// GenerateSQL generates SQL from natural language
	GenerateSQL(ctx context.Context, req dto.SQLRequest) (dto.SQLResponse, error)

	// Generate generates text from a prompt
	Generate(ctx context.Context, req dto.Request) (dto.Response, error)

	// SetAPIKey sets the OpenAI API key
	SetAPIKey(apiKey string)

	// GetAPIKey returns the current API key
	GetAPIKey() string
}

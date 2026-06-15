package llm

import (
	"context"
	"fmt"
	"strings"
	"time"
)

const (
	ProviderClaude   = "claude"
	ProviderChatGPT  = "chatgpt"
	ProviderDeepSeek = "deepseek"
)

type Analyzer interface {
	AnalyzePRandReviewPlan(ctx context.Context, prompt string) (string, error)
}

type Config struct {
	Provider            string
	APIKey              string
	Model               string
	RequestTimeout      time.Duration
	MaxRetries          int
	RetryBackoff        time.Duration
	MaxIdleConns        int
	MaxIdleConnsPerHost int
	MaxConnsPerHost     int
}

func NewAnalyzer(cfg Config) (Analyzer, error) {
	switch strings.ToLower(strings.TrimSpace(cfg.Provider)) {
	case "", ProviderClaude:
		return NewClaudeAnalyzer(cfg)
	case ProviderChatGPT:
		return NewChatGPTAnalyzer(cfg)
	case ProviderDeepSeek:
		return NewDeepSeekAnalyzer(cfg)
	default:
		return nil, fmt.Errorf("unsupported LLM provider %q", cfg.Provider)
	}
}

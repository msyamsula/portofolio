package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	llmdto "github.com/msyamsula/portofolio/backend-app/infrastructure/llm/dto"
)

func TestNewService(t *testing.T) {
	svc, err := New("")
	require.NoError(t, err)
	require.NotNil(t, svc)

	assert.Empty(t, svc.GetAPIKey())
}

func TestNewServiceWithAPIKey(t *testing.T) {
	svc, err := New("test-key")
	require.NoError(t, err)
	require.NotNil(t, svc)

	assert.Equal(t, "test-key", svc.GetAPIKey())
}

func TestSetAPIKey(t *testing.T) {
	svc, err := New("")
	require.NoError(t, err)

	svc.SetAPIKey("new-key")
	assert.Equal(t, "new-key", svc.GetAPIKey())
}

func TestGenerateSQLWithoutAPIKey(t *testing.T) {
	svc, err := New("")
	require.NoError(t, err)

	req := llmdto.SQLRequest{
		NaturalLanguage: "show me all users",
	}
	_, err = svc.GenerateSQL(context.Background(), req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "LLM client not initialized")
}

func TestGenerateSQLWithEmptySchema(t *testing.T) {
	svc, _ := New("test-key")

	req := llmdto.SQLRequest{
		NaturalLanguage: "show me all users",
		Schema:          "",
	}

	// This will fail because we're not actually calling the API
	_, err := svc.GenerateSQL(context.Background(), req)
	assert.Error(t, err)
}

func TestGenerateWithoutAPIKey(t *testing.T) {
	svc, err := New("")
	require.NoError(t, err)

	req := llmdto.Request{
		Prompt: "test prompt",
	}

	_, err = svc.Generate(context.Background(), req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "LLM client not initialized")
}

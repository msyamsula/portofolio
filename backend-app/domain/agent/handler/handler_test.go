package handler

import (
	"context"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/msyamsula/portofolio/backend-app/domain/agent/dto"
	llmservice "github.com/msyamsula/portofolio/backend-app/infrastructure/llm/service"
	"github.com/msyamsula/portofolio/backend-app/infrastructure/repl/readline"
	"github.com/msyamsula/portofolio/backend-app/pkg/parser"
	"github.com/stretchr/testify/assert"
)

// mockAgentService is a simple mock for testing
type mockAgentService struct{}

func (m *mockAgentService) Connect(ctx context.Context, cfg dto.ConnectionConfig) (*sqlx.DB, error) {
	return nil, nil
}

func (m *mockAgentService) ProcessInput(ctx context.Context, input string) (dto.QueryResponse, error) {
	return dto.QueryResponse{}, nil
}

func (m *mockAgentService) GenerateSQL(ctx context.Context, naturalLang string) (string, error) {
	return "SELECT * FROM test", nil
}

func (m *mockAgentService) ExecuteSQL(ctx context.Context, sql string) (dto.QueryResponse, error) {
	return dto.QueryResponse{}, nil
}

func (m *mockAgentService) GetConnectionInfo(ctx context.Context) (dto.ConnectionInfo, error) {
	return dto.ConnectionInfo{Database: "testdb", User: "testuser"}, nil
}

func (m *mockAgentService) GetTables(ctx context.Context) ([]dto.Table, error) {
	return []dto.Table{{Name: "users"}}, nil
}

func (m *mockAgentService) DescribeTable(ctx context.Context, tableName string) ([]dto.Column, error) {
	return []dto.Column{{Name: "id", Type: "int"}}, nil
}

func (m *mockAgentService) GetFullSchema(ctx context.Context) (dto.Schema, error) {
	return dto.Schema{}, nil
}

func (m *mockAgentService) ExplainQuery(ctx context.Context, sql string) (dto.ExplainResult, error) {
	return dto.ExplainResult{Plan: "test plan"}, nil
}

func (m *mockAgentService) BeginTransaction(ctx context.Context) error {
	return nil
}

func (m *mockAgentService) CommitTransaction(ctx context.Context) error {
	return nil
}

func (m *mockAgentService) RollbackTransaction(ctx context.Context) error {
	return nil
}

func (m *mockAgentService) InTransaction() bool {
	return false
}

func (m *mockAgentService) SetLLMService(svc llmservice.Service) {}

func (m *mockAgentService) SetReadline(rl *readline.Readline) {}

func (m *mockAgentService) GetSchemaForPrompt(ctx context.Context) (string, error) {
	return "Table users (id int, name text)", nil
}

func (m *mockAgentService) ParseInput(input string) parser.Result {
	return parser.Result{Type: parser.TypeSQL, SQL: "SELECT * FROM test"}
}

func (m *mockAgentService) IsDestructive(sql string) bool {
	return false
}

func TestNewHandler(t *testing.T) {
	mockSvc := &mockAgentService{}
	h := New(mockSvc)

	assert.NotNil(t, h)
	assert.NotNil(t, h.ui)
	assert.NotNil(t, h.prompt)
}

func TestSetDatabase(t *testing.T) {
	mockSvc := &mockAgentService{}
	h := New(mockSvc)

	h.SetDatabase("testdb")
	// The prompt should have been updated
	assert.NotNil(t, h.prompt)
}

func TestSetUser(t *testing.T) {
	mockSvc := &mockAgentService{}
	h := New(mockSvc)

	h.SetUser("testuser")
	assert.NotNil(t, h.prompt)
}

func TestSetTransaction(t *testing.T) {
	mockSvc := &mockAgentService{}
	h := New(mockSvc)

	h.SetTransaction(true)
	assert.NotNil(t, h.prompt)

	h.SetTransaction(false)
	assert.NotNil(t, h.prompt)
}

func TestPluralize(t *testing.T) {
	tests := []struct {
		count    int64
		expected string
	}{
		{0, "s"},
		{1, ""},
		{2, "s"},
		{10, "s"},
	}

	for _, tt := range tests {
		t.Run(string(rune(tt.count)), func(t *testing.T) {
			result := pluralize(tt.count)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSimplePrompt(t *testing.T) {
	result, err := simplePrompt("test> ")
	assert.NoError(t, err)
	assert.NotEmpty(t, result)
}

func TestPluralizeEdgeCases(t *testing.T) {
	tests := []struct {
		count    int64
		expected string
	}{
		{-1, "s"},
		{0, "s"},
		{1, ""},
		{100, "s"},
		{999, "s"},
	}

	for _, tt := range tests {
		t.Run(string(rune(tt.count)), func(t *testing.T) {
			result := pluralize(tt.count)
			assert.Equal(t, tt.expected, result)
		})
	}
}

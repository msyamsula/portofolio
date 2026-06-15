package test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"auto-pr-review/internal/config"
	"auto-pr-review/internal/prreviewagent"
	"auto-pr-review/internal/llm"
)

func TestApplyFixes_BedtimeStory(t *testing.T) {
	cfg, err := config.LoadPRReviewer()
	if err != nil {
		t.Skipf("config load failed (integration test): %v", err)
	}

	agent, err := llm.NewAnalyzer(cfg.LLM)
	if err != nil {
		t.Fatalf("create llm agent: %v", err)
	}

	deepseekAgent, ok := agent.(*llm.DeepSeekAnalyzer)
	if !ok {
		t.Fatalf("expected deepseek agent, got %T", agent)
	}

	workspace := filepath.Join("..", "..", "..", "bed-time-story")
	if err := os.RemoveAll(workspace); err != nil {
		t.Fatalf("clean workspace: %v", err)
	}
	if err := os.MkdirAll(workspace, 0o755); err != nil {
		t.Fatalf("create workspace: %v", err)
	}

	ex := &prreviewagent.Executor{}
	ex.SetAgent(deepseekAgent)

	info := &prreviewagent.PRInfo{
		WorkspaceDir: workspace,
		Plan:         "Write a bedtime story for children aged 1-5. The story should be simple, warm, and help children fall asleep. Write the story as a file named 'story.md' directly in the current directory. The story should mention a sleepy moon, a friendly bear, and soft rain.",
	}

	ctx := context.Background()
	if err := ex.ApplyFixes(ctx, info); err != nil {
		t.Fatalf("apply fixes: %v", err)
	}

	t.Logf("report: %s", info.Report)

	storyPath := filepath.Join(workspace, "story.md")
	data, err := os.ReadFile(storyPath)
	if err != nil {
		t.Fatalf("read story file: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("story file is empty")
	}
	t.Logf("story:\n%s", string(data))
}

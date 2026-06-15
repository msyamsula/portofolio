package aitools

import (
	"bytes"
	"context"
	"os/exec"
	"strings"
)

type GitDiff struct {
	workspace string
}

func NewGitDiff(workspace string) *GitDiff {
	return &GitDiff{workspace: workspace}
}

func (t *GitDiff) Name() string        { return "git_diff" }
func (t *GitDiff) Description() string { return "Show staged and unstaged changes in the workspace." }

func (t *GitDiff) Parameters() map[string]any {
	return map[string]any{"type": "object", "properties": map[string]any{}}
}

func (t *GitDiff) Execute(ctx context.Context, _ map[string]any) (string, error) {
	cmd := exec.CommandContext(ctx, "git", "diff")
	cmd.Dir = t.workspace
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	if len(out) == 0 {
		return "(no changes)", nil
	}
	return strings.TrimSpace(string(out)), nil
}

type GitStatus struct {
	workspace string
}

func NewGitStatus(workspace string) *GitStatus {
	return &GitStatus{workspace: workspace}
}

func (t *GitStatus) Name() string        { return "git_status" }
func (t *GitStatus) Description() string { return "Show the working tree status." }

func (t *GitStatus) Parameters() map[string]any {
	return map[string]any{"type": "object", "properties": map[string]any{}}
}

func (t *GitStatus) Execute(ctx context.Context, _ map[string]any) (string, error) {
	cmd := exec.CommandContext(ctx, "git", "status", "--porcelain")
	cmd.Dir = t.workspace
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	if len(out) == 0 {
		return "(clean)", nil
	}
	return strings.TrimSpace(string(out)), nil
}

type RunCommand struct {
	workspace string
}

func NewRunCommand(workspace string) *RunCommand {
	return &RunCommand{workspace: workspace}
}

func (t *RunCommand) Name() string        { return "run_command" }
func (t *RunCommand) Description() string { return "Run a shell command in the workspace directory." }

func (t *RunCommand) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"cmd": map[string]any{"type": "string", "description": "The shell command to execute."},
		},
		"required": []string{"cmd"},
	}
}

func (t *RunCommand) Execute(ctx context.Context, args map[string]any) (string, error) {
	cmd, _ := args["cmd"].(string)
	if cmd == "" {
		return "", nil
	}
	c := exec.CommandContext(ctx, "sh", "-c", cmd)
	c.Dir = t.workspace
	var out, stderr bytes.Buffer
	c.Stdout = &out
	c.Stderr = &stderr
	_ = c.Run()
	output := strings.TrimSpace(out.String())
	if stderr.Len() > 0 {
		output += "\n" + strings.TrimSpace(stderr.String())
	}
	if output == "" {
		return "(empty)", nil
	}
	return output, nil
}

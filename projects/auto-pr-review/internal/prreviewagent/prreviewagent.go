package prreviewagent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"auto-pr-review/internal/aitools"
	"auto-pr-review/internal/githubprdescriptor"
	"auto-pr-review/internal/llm"
)

const systemPrompt = `You are an autonomous PR fix agent. You have a git worktree on the feature branch, already merged with the latest develop.

Phase 1 — Survey:
  1. Read the PR description below to understand what this PR does.
  2. Use list_files to survey the project structure.
  3. Read relevant files that the review comments reference.
  4. Use run_command to check the current build/test state.

Phase 2 — Fix:
  Apply each actionable item from the PR discussion. Use write_file with the full updated file content.
  After each batch of changes, use run_command to verify the build/tests still pass.
  Use git_diff to verify your changes are minimal and targeted.

Phase 3 — Validate (REQUIRED before reporting):
  You MUST run these commands before producing your final report:
  - Use the run_command tool to execute: make test
  - Use the run_command tool to execute: make lint-gci-new
  If a command is not available, you MUST explicitly note this in your report.
  If a command fails, you MUST read the error output, fix the failures, and re-run until it passes.
  DO NOT skip this step. DO NOT produce a report without running these commands first.

Phase 4 — Report:
  When all changes are applied AND validation passes, produce a Markdown report.
  The first line MUST be: COMMIT_MSG: <your commit message>
  Then include the full report describing what was fixed and what was left open.
  At the end, include a ## Diff section summarizing the changes (file paths and a brief description of each).
  If there is no diff, state: "No changes in this round of review."

Rules:
- Be surgical. Only change what is needed. Never refactor or beautify unrelated code.
- Always read a file with read_file before editing it.
- Never include credentials, API keys, or secrets in any file.
- Run make test after every change. Do not report done if tests fail.
- If you cannot fix something, explain why in your report.
- Make multiple independent tool calls in a single response to reduce turns.`

type Config struct {
	GitHubClient     *githubprdescriptor.Client
	GitHubPool       githubprdescriptor.PoolConfig
	GitHubTokenPath  string
	Agent            llm.Agent
	WorkspaceRoot    string
	DevelopBranch    string
	OutputDir        string
	TraceDir         string
	GitUser          string
	GitEmail         string
	MaxAgentTurns    int
	SystemPromptPath string
}

type Executor struct {
	githubClient *githubprdescriptor.Client
	agent        llm.Agent
	cfg          Config
	logger       *slog.Logger
	systemPrompt string
}

type PRInfo struct {
	Ref          githubprdescriptor.PRRef
	Discussion   githubprdescriptor.PRDiscussion
	URL          string
	WorkspaceDir string
	Plan         string
	Report       string
	Branch       string
}

func New(cfg Config) (*Executor, error) {
	githubClient := cfg.GitHubClient
	if githubClient == nil {
		githubClient = githubprdescriptor.NewClient(nil, cfg.GitHubTokenPath, cfg.GitHubPool)
	}

	agent := cfg.Agent
	if agent == nil {
		return nil, fmt.Errorf("agent is required")
	}

	e := &Executor{
		githubClient: githubClient,
		agent:        agent,
		cfg:          cfg,
		logger:       slog.Default(),
		systemPrompt: loadSystemPrompt(cfg.SystemPromptPath),
	}
	e.logger.Info("prreviewagent initialized",
		slog.String("workspace_root", cfg.WorkspaceRoot),
		slog.String("develop_branch", cfg.DevelopBranch))

	return e, nil
}

func (e *Executor) SetAgent(a llm.Agent) { e.agent = a }
func (e *Executor) SetLogger(l *slog.Logger)  { e.logger = l }

func loadSystemPrompt(path string) string {
	if path == "" {
		return systemPrompt
	}
	data, err := os.ReadFile(path)
	if err != nil {
		slog.Warn("failed to read system prompt file, using built-in default",
			slog.String("path", path),
			slog.String("error", err.Error()))
		return systemPrompt
	}
	trimmed := strings.TrimSpace(string(data))
	if trimmed == "" {
		slog.Warn("system prompt file is empty, using built-in default",
			slog.String("path", path))
		return systemPrompt
	}
	return trimmed
}

func (e *Executor) Execute(ctx context.Context, prLink string) (string, error) {
	if e.cfg.WorkspaceRoot == "" {
		return "", fmt.Errorf("workspace root is required")
	}

	pr, err := githubprdescriptor.ParseURL(prLink)
	if err != nil {
		return "", fmt.Errorf("parse PR link: %w", err)
	}

	discussion, err := e.githubClient.FetchDiscussion(ctx, pr)
	if err != nil {
		return "", fmt.Errorf("fetch PR discussion: %w", err)
	}

	if discussion.Branch == "" {
		return "", fmt.Errorf("PR branch not found in GitHub response")
	}

	info := &PRInfo{
		Ref:        pr,
		Discussion: discussion,
		URL:        prLink,
		Branch:     discussion.Branch,
	}

	worktree, err := e.prepareWorktree(ctx, info)
	if err != nil {
		return "", fmt.Errorf("prepare worktree: %w", err)
	}
	defer e.cleanupWorktree(worktree)

	if err := e.mergeDevelop(ctx, worktree, info); err != nil {
		return "", err
	}

	if err := e.ApplyFixes(ctx, info); err != nil {
		return "", fmt.Errorf("apply fixes: %w", err)
	}

	if info.Report == "" {
		return "", fmt.Errorf("agent returned empty report")
	}

	if err := e.checkCredentials(ctx, worktree); err != nil {
		return "", fmt.Errorf("credential check: %w", err)
	}

	msg := extractCommitMsg(info.Report)
	if err := e.commit(ctx, worktree, msg); err != nil {
		return "", fmt.Errorf("commit: %w", err)
	}

	return e.writeReport(ctx, info)
}

func (e *Executor) prepareWorktree(ctx context.Context, info *PRInfo) (string, error) {
	worktree, err := os.MkdirTemp("", "agent-wt-*")
	if err != nil {
		return "", fmt.Errorf("create worktree dir: %w", err)
	}
	info.WorkspaceDir = worktree

	cmds := [][]string{
		{"pull", "origin", e.cfg.DevelopBranch},
		{"fetch", "origin"},
		{"worktree", "prune"},
	}
	for _, args := range cmds {
		if err := runGitCmd(ctx, e.cfg.WorkspaceRoot, args); err != nil {
			e.logger.Warn("git prep failed", slog.String("error", err.Error()))
		}
	}

	if stale := findExistingWorktree(ctx, e.cfg.WorkspaceRoot, info.Branch); stale != "" {
		e.logger.Warn("removing stale worktree for branch", slog.String("branch", info.Branch), slog.String("path", stale))
		_ = runGitCmd(ctx, e.cfg.WorkspaceRoot, []string{"worktree", "remove", "--force", stale})
	}

	_ = runGitCmd(ctx, e.cfg.WorkspaceRoot, []string{"fetch", "origin", "+refs/heads/" + info.Branch + ":refs/heads/" + info.Branch})

	if err := runGitCmd(ctx, e.cfg.WorkspaceRoot, []string{"worktree", "add", worktree, info.Branch}); err != nil {
		_ = runGitCmd(ctx, e.cfg.WorkspaceRoot, []string{"worktree", "remove", "--force", worktree})
		if err := runGitCmd(ctx, e.cfg.WorkspaceRoot, []string{"worktree", "add", worktree, info.Branch}); err != nil {
			return "", fmt.Errorf("worktree add %s: %w", info.Branch, err)
		}
	}

	setupCmds := [][]string{
		{"config", "user.name", e.cfg.GitUser},
		{"config", "user.email", e.cfg.GitEmail},
	}
	for _, args := range setupCmds {
		if err := retryGitCmd(ctx, worktree, args, 3, 100*time.Millisecond); err != nil {
			return "", fmt.Errorf("configure worktree: %w", err)
		}
	}

	return worktree, nil
}

func (e *Executor) mergeDevelop(ctx context.Context, worktree string, info *PRInfo) error {
	if err := runGitCmd(ctx, worktree, []string{"merge", "origin/" + e.cfg.DevelopBranch}); err != nil {
		e.logger.Warn("merge develop failed, discarding worktree",
			slog.String("branch", info.Branch),
			slog.String("error", err.Error()))
		e.cleanupWorktree(worktree)
		return fmt.Errorf("merge conflict with %s on branch %s", e.cfg.DevelopBranch, info.Branch)
	}
	return nil
}

func (e *Executor) cleanupWorktree(worktree string) {
	_ = runGitCmd(context.Background(), e.cfg.WorkspaceRoot, []string{"worktree", "remove", "--force", worktree})
	_ = os.RemoveAll(worktree)
}

func (e *Executor) checkCredentials(ctx context.Context, worktree string) error {
	diff, err := captureGitCmd(ctx, worktree, []string{"diff", "--staged"})
	if err != nil {
		return err
	}

	patterns := []string{
		`(?:api_key|apikey|api-key|secret|password|token|credential)[\s]*[=:][\s]*['"]?[A-Za-z0-9._\-]{16,}['"]?`,
		`-----BEGIN (?:RSA |EC |DSA |OPENSSH )?PRIVATE KEY-----`,
		`sk-[A-Za-z0-9]{32,}`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(`(?i)` + pattern)
		if re.MatchString(diff) {
			return fmt.Errorf("credential pattern detected in staged changes: %s", pattern)
		}
	}

	return nil
}

func (e *Executor) commit(ctx context.Context, worktree, msg string) error {
	files, err := captureGitCmd(ctx, worktree, []string{"status", "--porcelain"})
	if err != nil {
		return err
	}
	if strings.TrimSpace(files) == "" {
		return nil
	}

	cmds := [][]string{
		{"add", "-A"},
		{"commit", "--no-verify", "-m", msg},
	}
	for _, args := range cmds {
		if err := runGitCmd(ctx, worktree, args); err != nil {
			return err
		}
	}

	e.logger.Info("changes committed", slog.String("branch", e.cfg.WorkspaceRoot))
	return nil
}

func extractCommitMsg(report string) string {
	report = strings.TrimSpace(report)
	if !strings.HasPrefix(report, "COMMIT_MSG:") {
		return "apply PR review fixes"
	}
	idx := strings.Index(report, "\n")
	if idx < 0 {
		return strings.TrimPrefix(report, "COMMIT_MSG:")
	}
	return strings.TrimSpace(report[len("COMMIT_MSG:"):idx])
}

func (e *Executor) ApplyFixes(ctx context.Context, info *PRInfo) error {
	registry := aitools.NewRegistry(
		aitools.NewReadFile(info.WorkspaceDir),
		aitools.NewWriteFile(info.WorkspaceDir),
		aitools.NewListFiles(info.WorkspaceDir),
		aitools.NewGitDiff(info.WorkspaceDir),
		aitools.NewGitStatus(info.WorkspaceDir),
		aitools.NewRunCommand(info.WorkspaceDir),
	)

	prompt := buildDiscussionPrompt(info.Discussion)

	messages := []llm.Message{
		{Role: "system", Content: e.systemPrompt},
		{Role: "user", Content: prompt},
	}

	traceFile := traceFilePath(e.cfg.TraceDir, info.URL)
	prevTrace := readTrace(traceFile)
	if prevTrace != "" {
		prevTrace = stripTracePaths(prevTrace, info.WorkspaceDir)
		messages = append(messages, llm.Message{
			Role: "user",
			Content: fmt.Sprintf("CONTEXT: This is a retry. Your workspace is %s. File paths below use this directory. Here is the trace from the previous attempt. Use this to avoid repeating work that was already done. IMPORTANT: do NOT copy old /tmp/agent-wt-* paths from the trace — always use paths relative to your current workspace.\n\n%s", info.WorkspaceDir, prevTrace),
		})
		e.logger.Info("resuming from previous trace", slog.String("github_link", info.URL))
	}

	maxTurns := e.cfg.MaxAgentTurns
	if maxTurns <= 0 {
		maxTurns = 50
	}

	for turn := 0; turn < maxTurns; turn++ {
		e.logger.Info("agent turn", slog.Int("turn", turn+1), slog.Int("max", maxTurns), slog.String("github_link", info.URL))
		resp, err := e.agent.Chat(ctx, messages, registry.List())
		if err != nil {
			return fmt.Errorf("agent chat: %w", err)
		}

		if resp.Text != "" {
			e.logger.Info("agent response", slog.String("text", resp.Text), slog.Int("turn", turn+1))
		}

		if len(resp.ToolCalls) == 0 {
			e.logger.Info("agent finished", slog.Int("turns", turn+1), slog.String("github_link", info.URL))
			info.Report = resp.Text
			return nil
		}

		for _, tc := range resp.ToolCalls {
			e.logger.Info("agent tool call", slog.String("tool", tc.Name), slog.Any("args", tc.Args), slog.Int("turn", turn+1))
		}

		messages = append(messages, llm.Message{Role: "assistant", Content: resp.Text, ToolCalls: resp.ToolCalls})

		for _, tc := range resp.ToolCalls {
			result, err := registry.Execute(ctx, tc.Name, tc.Args)
			if err != nil {
				result = fmt.Sprintf("error: %v", err)
			}
			e.logger.Info("agent tool result", slog.String("tool", tc.Name), slog.String("result", result), slog.Int("turn", turn+1))
			messages = append(messages, llm.Message{Role: "tool", Content: result, ToolCallID: tc.ID})
			appendTrace(traceFile, tc.Name, tc.Args, result)
		}
	}

	return fmt.Errorf("agent did not finish within %d turns", maxTurns)
}

func traceFilePath(dir, prURL string) string {
	sanitized := strings.NewReplacer(
		"://", "-", "/", "-", ".", "-", ":", "-", "?", "-", "&", "-", "=", "-",
	).Replace(prURL)
	sanitized = strings.Trim(sanitized, "-")
	return filepath.Join(dir, sanitized+".jsonl")
}

type traceEntry struct {
	Tool   string         `json:"tool"`
	Args   map[string]any `json:"args,omitempty"`
	Result string         `json:"result"`
}

func readTrace(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

func appendTrace(path, tool string, args map[string]any, result string) {
	_ = os.MkdirAll(filepath.Dir(path), 0o755)
	entry := traceEntry{Tool: tool, Args: args, Result: result}
	b, _ := json.Marshal(entry)
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	defer f.Close()
	_, _ = f.Write(b)
	_, _ = f.Write([]byte("\n"))
}

func buildDiscussionPrompt(discussion githubprdescriptor.PRDiscussion) string {
	var b strings.Builder
	b.WriteString("## PR Discussion\n\n")
	b.WriteString(fmt.Sprintf("URL: %s\n", discussion.URL))
	b.WriteString(fmt.Sprintf("Title: %s\n", discussion.Title))
	b.WriteString(fmt.Sprintf("State: %s\n", discussion.State))
	b.WriteString(fmt.Sprintf("Author: %s\n\n", discussion.Author))

	if strings.TrimSpace(discussion.Body) != "" {
		b.WriteString("### Description\n\n")
		b.WriteString(discussion.Body)
		b.WriteString("\n\n")
	}

	b.WriteString("### Issue Comments\n\n")
	for _, c := range discussion.IssueComments {
		fmt.Fprintf(&b, "- **%s**: %s\n", c.User.Login, c.Body)
	}
	if len(discussion.IssueComments) == 0 {
		b.WriteString("(none)\n")
	}

	b.WriteString("\n### Review Summaries\n\n")
	for _, r := range discussion.Reviews {
		fmt.Fprintf(&b, "- **%s** (%s): %s\n", r.User.Login, r.State, r.Body)
	}
	if len(discussion.Reviews) == 0 {
		b.WriteString("(none)\n")
	}

	b.WriteString("\n### Inline Review Comments\n\n")
	for _, rc := range discussion.ReviewComments {
		fmt.Fprintf(&b, "- **%s** on `%s`: %s\n", rc.User.Login, rc.Path, rc.Body)
	}
	if len(discussion.ReviewComments) == 0 {
		b.WriteString("(none)\n")
	}

	b.WriteString("\nApply fixes for all actionable items above.\n")
	b.WriteString("Read relevant files first, make targeted changes, then build and test.\n")
	b.WriteString("When done, provide a Markdown report starting with COMMIT_MSG: <message>\n")

	return b.String()
}

func (e *Executor) writeReport(ctx context.Context, info *PRInfo) (string, error) {
	outputDir := e.cfg.OutputDir
	if outputDir == "" {
		outputDir = "reports/agent"
	}

	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return "", fmt.Errorf("create output dir: %w", err)
	}

	reportPath := filepath.Join(outputDir, fmt.Sprintf("agent-%s-%s-%d.md",
		sanitizePart(info.Ref.Owner),
		sanitizePart(info.Ref.Repo),
		info.Ref.Number,
	))

	var b strings.Builder
	b.WriteString(fmt.Sprintf("# Agent Report\n\n"))
	b.WriteString(fmt.Sprintf("**PR:** %s\n\n", info.URL))
	b.WriteString(fmt.Sprintf("**Branch:** %s\n\n", info.Branch))
	b.WriteString(fmt.Sprintf("## Summary\n\n%s\n\n", info.Report))
	b.WriteString(fmt.Sprintf("## Discussion Stats\n\n"))
	b.WriteString(fmt.Sprintf("- Comments: %d\n", len(info.Discussion.IssueComments)))
	b.WriteString(fmt.Sprintf("- Reviews: %d\n", len(info.Discussion.Reviews)))
	b.WriteString(fmt.Sprintf("- Inline comments: %d\n", len(info.Discussion.ReviewComments)))

	if err := os.WriteFile(reportPath, []byte(b.String()), 0o644); err != nil {
		return "", fmt.Errorf("write report: %w", err)
	}

	return reportPath, nil
}

func runGitCmd(ctx context.Context, dir string, args []string) error {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func retryGitCmd(ctx context.Context, dir string, args []string, attempts int, delay time.Duration) error {
	var lastErr error
	for i := 0; i < attempts; i++ {
		cmd := exec.CommandContext(ctx, "git", args...)
		cmd.Dir = dir
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err == nil {
			return nil
		} else {
			lastErr = err
		}
		if i < attempts-1 {
			time.Sleep(delay)
		}
	}
	return lastErr
}

func captureGitCmd(ctx context.Context, dir string, args []string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		var stderr bytes.Buffer
		cmd.Stderr = &stderr
		return "", fmt.Errorf("git %s: %w\n%s", strings.Join(args, " "), err, stderr.String())
	}
	return strings.TrimSpace(string(out)), nil
}

func sanitizePart(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	return strings.NewReplacer("/", "-", " ", "-", ".", "-").Replace(s)
}

func findExistingWorktree(ctx context.Context, repoDir, branch string) string {
	out, err := captureGitCmd(ctx, repoDir, []string{"worktree", "list", "--porcelain"})
	if err != nil {
		return ""
	}
	branchRef := "refs/heads/" + branch
	var currentPath string
	for _, line := range strings.Split(out, "\n") {
		if strings.HasPrefix(line, "worktree ") {
			currentPath = strings.TrimSpace(strings.TrimPrefix(line, "worktree "))
		}
		if strings.HasPrefix(line, "branch ") && strings.TrimSpace(strings.TrimPrefix(line, "branch ")) == branchRef {
			if currentPath == repoDir {
				return "" // skip the main working tree
			}
			return currentPath
		}
	}
	return ""
}

var tracePathRe = regexp.MustCompile(`/tmp/agent-wt-\d+`)

func stripTracePaths(trace, workspaceDir string) string {
	return tracePathRe.ReplaceAllString(trace, workspaceDir)
}

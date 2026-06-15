package prreviewagent

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"auto-pr-review/internal/aitools"
	"auto-pr-review/internal/githubprdescriptor"
	"auto-pr-review/internal/llm"
)

const maxAgentTurns = 50
const systemPrompt = `You are an autonomous PR fix agent. You have a git worktree on the feature branch, already merged with the latest develop.

Phase 1 — Survey:
  1. Read the PR description below to understand what this PR does.
  2. Use list_files to survey the project structure.
  3. Read relevant files that the review comments reference.
  4. Use run_command to check the current build/test state.

Phase 2 — Fix:
  Apply each actionable item from the PR discussion. Use write_file with the full updated file content.
  After a batch of changes, run build/tests with run_command.
  Use git_diff to verify your changes are minimal and correct.

Phase 3 — Report:
  When all items are addressed or you cannot proceed further, produce a Markdown report.
  The first line MUST be: COMMIT_MSG: <your commit message>
  Then the full report describing what was fixed and what was left open.

Rules:
- Be surgical. Only change what is needed. Never refactor or beautify unrelated code.
- Always read a file with read_file before editing it.
- Never include credentials, API keys, or secrets in any file.
- If you cannot fix something, explain why in your report.`

type Config struct {
	GitHubClient    *githubprdescriptor.Client
	GitHubPool      githubprdescriptor.PoolConfig
	GitHubTokenPath string
	Agent           llm.Agent
	LLM             llm.Config
	WorkspaceRoot   string
	DevelopBranch   string
	OutputDir       string
	GitUser         string
	GitEmail        string
}

type Executor struct {
	githubClient *githubprdescriptor.Client
	agent        llm.Agent
	cfg          Config
	logger       *slog.Logger
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
		analyzer, err := llm.NewAnalyzer(cfg.LLM)
		if err != nil {
			return nil, err
		}
		var ok bool
		agent, ok = analyzer.(llm.Agent)
		if !ok {
			return nil, fmt.Errorf("llm provider %q does not implement agent interface", cfg.LLM.Provider)
		}
	}

	e := &Executor{
		githubClient: githubClient,
		agent:        agent,
		cfg:          cfg,
		logger:       slog.Default(),
	}
	e.logger.Info("prreviewagent initialized",
		slog.String("provider", cfg.LLM.Provider),
		slog.String("workspace_root", cfg.WorkspaceRoot),
		slog.String("develop_branch", cfg.DevelopBranch),
		slog.String("llm_request_timeout", cfg.LLM.RequestTimeout.String()))

	return e, nil
}

func (e *Executor) SetAgent(a llm.Agent) { e.agent = a }
func (e *Executor) SetLogger(l *slog.Logger)  { e.logger = l }

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
	}
	for _, args := range cmds {
		if err := runGitCmd(ctx, e.cfg.WorkspaceRoot, args); err != nil {
			e.logger.Warn("git prep failed", slog.String("error", err.Error()))
		}
	}

	if err := runGitCmd(ctx, e.cfg.WorkspaceRoot, []string{"worktree", "add", worktree, info.Branch}); err != nil {
		return "", fmt.Errorf("worktree add %s: %w", info.Branch, err)
	}

	setupCmds := [][]string{
		{"config", "user.name", e.cfg.GitUser},
		{"config", "user.email", e.cfg.GitEmail},
	}
	for _, args := range setupCmds {
		if err := runGitCmd(ctx, worktree, args); err != nil {
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
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: prompt},
	}

	for turn := 0; turn < maxAgentTurns; turn++ {
		e.logger.Info("agent turn", slog.Int("turn", turn+1), slog.Int("max", maxAgentTurns), slog.String("github_link", info.URL))
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
		}
	}

	return fmt.Errorf("agent did not finish within %d turns", maxAgentTurns)
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

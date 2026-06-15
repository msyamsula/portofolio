package prreviewagent

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"auto-pr-review/internal/aitools"
	"auto-pr-review/internal/githubprdescriptor"
	"auto-pr-review/internal/llm"
)

const maxAgentTurns = 20

type Config struct {
	GitHubClient    *githubprdescriptor.Client
	GitHubPool      githubprdescriptor.PoolConfig
	GitHubTokenPath string
	Agent           llm.Agent
	LLM             llm.Config
	DevelopBranch   string
	RepoPath        string
	OutputDir       string
	GitUser         string
	GitEmail        string
}

type Executor struct {
	githubClient *githubprdescriptor.Client
	agent        llm.Agent
	cfg          Config
}

type PRInfo struct {
	Ref          githubprdescriptor.PRRef
	Branch       string
	Discussion   githubprdescriptor.PRDiscussion
	URL          string
	WorkspaceDir string
	Plan         string
	Report       string
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

	return &Executor{
		githubClient: githubClient,
		agent:        agent,
		cfg:          cfg,
	}, nil
}

func (e *Executor) SetAgent(a llm.Agent) { e.agent = a }

func (e *Executor) Execute(ctx context.Context, prLink string) (string, error) {
	pr, err := githubprdescriptor.ParseURL(prLink)
	if err != nil {
		return "", fmt.Errorf("parse PR link: %w", err)
	}

	discussion, err := e.githubClient.FetchDiscussion(ctx, pr)
	if err != nil {
		return "", fmt.Errorf("fetch PR discussion: %w", err)
	}

	info := &PRInfo{
		Ref:        pr,
		Discussion: discussion,
		URL:        prLink,
	}

	if err := e.setupWorkspace(ctx, info); err != nil {
		return "", fmt.Errorf("setup workspace: %w", err)
	}
	defer os.RemoveAll(info.WorkspaceDir)

	if err := e.mergeBase(ctx, info); err != nil {
		return "", fmt.Errorf("merge base: %w", err)
	}

	plan, err := e.buildFixPlan(ctx, info)
	if err != nil {
		return "", fmt.Errorf("build fix plan: %w", err)
	}
	info.Plan = plan

	if err := e.ApplyFixes(ctx, info); err != nil {
		return "", fmt.Errorf("apply fixes: %w", err)
	}

	if err := e.checkCredentials(ctx, info); err != nil {
		return "", fmt.Errorf("credential check: %w", err)
	}

	if err := e.commitAndPush(ctx, info); err != nil {
		return "", fmt.Errorf("commit and push: %w", err)
	}

	if err := e.resolveDiscussions(ctx, info); err != nil {
		return "", fmt.Errorf("resolve discussions: %w", err)
	}

	return e.writeReport(ctx, info)
}

func (e *Executor) setupWorkspace(ctx context.Context, info *PRInfo) error {
	repoPath := e.cfg.RepoPath
	if repoPath == "" {
		return fmt.Errorf("repo path is required")
	}

	workspace, err := os.MkdirTemp("", "executor-*")
	if err != nil {
		return fmt.Errorf("create workspace: %w", err)
	}
	info.WorkspaceDir = workspace

	cmds := [][]string{
		{"clone", repoPath, workspace},
		{"config", "user.name", e.cfg.GitUser},
		{"config", "user.email", e.cfg.GitEmail},
		{"fetch", "origin"},
		{"checkout", info.Branch},
	}
	for _, args := range cmds {
		if err := runGitCmd(ctx, workspace, args); err != nil {
			return err
		}
	}

	return nil
}

func (e *Executor) mergeBase(ctx context.Context, info *PRInfo) error {
	if err := runGitCmd(ctx, info.WorkspaceDir, []string{"merge", "origin/" + e.cfg.DevelopBranch}); err != nil {
		return fmt.Errorf("merge %s: %w", e.cfg.DevelopBranch, err)
	}
	return nil
}

func (e *Executor) buildFixPlan(ctx context.Context, info *PRInfo) (string, error) {
	var b strings.Builder
	b.WriteString("Review this PR discussion and produce a concrete fix plan with numbered action items.\n")
	b.WriteString("For each item, specify the file path and exact change. Be surgical.\n\n")
	b.WriteString("## PR Discussion\n\n")
	b.WriteString(fmt.Sprintf("URL: %s\n", info.Discussion.URL))
	b.WriteString(fmt.Sprintf("Title: %s\n", info.Discussion.Title))
	b.WriteString(fmt.Sprintf("State: %s\n", info.Discussion.State))

	for _, comment := range info.Discussion.IssueComments {
		b.WriteString(fmt.Sprintf("\nComment by %s: %s\n", comment.User.Login, comment.Body))
	}
	for _, review := range info.Discussion.Reviews {
		b.WriteString(fmt.Sprintf("\nReview by %s (%s): %s\n", review.User.Login, review.State, review.Body))
	}
	for _, rc := range info.Discussion.ReviewComments {
		b.WriteString(fmt.Sprintf("\nReview comment by %s on %s: %s\n", rc.User.Login, rc.Path, rc.Body))
	}

	resp, err := e.agent.Chat(ctx, []llm.Message{
		{Role: "system", Content: "You are a fix plan generator. Produce only the plan, no tool calls."},
		{Role: "user", Content: b.String()},
	}, nil)
	if err != nil {
		return "", err
	}
	return resp.Text, nil
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

	messages := []llm.Message{
		{Role: "system", Content: `You are an autonomous coding agent. Execute this fix plan using the provided tools.
- Read files before editing them.
- Use write_file to apply changes (always provide the full file content, not diffs).
- Use run_command to build, test, and verify your changes.
- If a tool call fails, try a different approach.
- When all changes are applied and verified, respond with a summary of what you did.`},
		{Role: "user", Content: fmt.Sprintf("Fix plan:\n\n%s", info.Plan)},
	}

	for turn := 0; turn < maxAgentTurns; turn++ {
		resp, err := e.agent.Chat(ctx, messages, registry.List())
		if err != nil {
			return fmt.Errorf("agent chat: %w", err)
		}

		if len(resp.ToolCalls) == 0 {
			info.Report = resp.Text
			return nil
		}

		messages = append(messages, llm.Message{Role: "assistant", Content: resp.Text, ToolCalls: resp.ToolCalls})

		for _, tc := range resp.ToolCalls {
			result, err := registry.Execute(ctx, tc.Name, tc.Args)
			if err != nil {
				result = fmt.Sprintf("error: %v", err)
			}
			messages = append(messages, llm.Message{Role: "tool", Content: result, ToolCallID: tc.ID})
		}
	}

	return fmt.Errorf("agent did not finish within %d turns", maxAgentTurns)
}

func (e *Executor) checkCredentials(ctx context.Context, info *PRInfo) error {
	diff, err := captureGitCmd(ctx, info.WorkspaceDir, []string{"diff", "--staged"})
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

func (e *Executor) commitAndPush(ctx context.Context, info *PRInfo) error {
	files, err := captureGitCmd(ctx, info.WorkspaceDir, []string{"status", "--porcelain"})
	if err != nil {
		return err
	}
	if strings.TrimSpace(files) == "" {
		return nil
	}

	msg := fmt.Sprintf("fix: apply PR review changes for %s", info.URL)
	cmds := [][]string{
		{"add", "-A"},
		{"commit", "-m", msg},
		{"push", "origin", info.Branch},
	}
	for _, args := range cmds {
		if err := runGitCmd(ctx, info.WorkspaceDir, args); err != nil {
			return err
		}
	}

	return nil
}

func (e *Executor) resolveDiscussions(ctx context.Context, info *PRInfo) error {
	return nil
}

func (e *Executor) writeReport(ctx context.Context, info *PRInfo) (string, error) {
	outputDir := e.cfg.OutputDir
	if outputDir == "" {
		outputDir = "reports/executor"
	}

	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return "", fmt.Errorf("create output dir: %w", err)
	}

	reportPath := filepath.Join(outputDir, fmt.Sprintf("executor-%s-%s-%d.md",
		sanitizePart(info.Ref.Owner),
		sanitizePart(info.Ref.Repo),
		info.Ref.Number,
	))

	var b strings.Builder
	b.WriteString(fmt.Sprintf("# Executor Report\n\n"))
	b.WriteString(fmt.Sprintf("**PR:** %s\n\n", info.URL))
	b.WriteString(fmt.Sprintf("**Branch:** %s\n\n", info.Branch))
	b.WriteString(fmt.Sprintf("## Fix Plan\n\n%s\n\n", info.Plan))
	b.WriteString(fmt.Sprintf("## Changes Applied\n\n%s\n\n", info.Report))
	b.WriteString(fmt.Sprintf("## Discussions\n\n"))
	b.WriteString(fmt.Sprintf("- Total comments: %d\n", len(info.Discussion.IssueComments)))
	b.WriteString(fmt.Sprintf("- Total reviews: %d\n", len(info.Discussion.Reviews)))
	b.WriteString(fmt.Sprintf("- Total review comments: %d\n", len(info.Discussion.ReviewComments)))

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
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git %s: %w", strings.Join(args, " "), err)
	}
	return nil
}

func captureGitCmd(ctx context.Context, dir string, args []string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git %s: %w", strings.Join(args, " "), err)
	}
	return strings.TrimSpace(string(out)), nil
}

func sanitizePart(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	return strings.NewReplacer("/", "-", " ", "-", ".", "-").Replace(s)
}

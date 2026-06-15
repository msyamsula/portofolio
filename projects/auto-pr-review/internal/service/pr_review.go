package service

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"auto-pr-review/internal/githubprdescriptor"
	"auto-pr-review/internal/llm"
)

type PRDiscussionReviewConfig struct {
	Analyzer          llm.Analyzer
	LLM               llm.Config
	GitHubTokenPath   string
	GitHubPoolConfig  githubprdescriptor.PoolConfig
	GitHubClient      *githubprdescriptor.Client
	OutputDir         string
	HTTPClient        *http.Client
	FileHandler       FileHandler
}

type PRReviewer struct {
	githubClient *githubprdescriptor.Client
	fileHandler  FileHandler
	config       PRDiscussionReviewConfig
	analyzer     llm.Analyzer
}

type FileHandler interface {
	Replace(path string, data []byte) error
}

type AtomicFileHandler struct{}

func NewPRReviewer(cfg PRDiscussionReviewConfig) (*PRReviewer, error) {
	githubClient := cfg.GitHubClient
	if githubClient == nil {
		client := cfg.HTTPClient
		githubClient = githubprdescriptor.NewClient(client, strings.TrimSpace(cfg.GitHubTokenPath), cfg.GitHubPoolConfig)
	}

	fileHandler := cfg.FileHandler
	if fileHandler == nil {
		fileHandler = AtomicFileHandler{}
	}

	analyzer := cfg.Analyzer
	if analyzer == nil {
		var err error
		analyzer, err = llm.NewAnalyzer(cfg.LLM)
		if err != nil {
			return nil, err
		}
	}

	return &PRReviewer{
		githubClient: githubClient,
		fileHandler:  fileHandler,
		config:       cfg,
		analyzer:     analyzer,
	}, nil
}

func (r *PRReviewer) WriteReviewPlanMarkdown(ctx context.Context, githubLink, outputPath string) error {
	if strings.TrimSpace(outputPath) == "" {
		return fmt.Errorf("output path is required")
	}
	if r == nil {
		return fmt.Errorf("PR reviewer is required")
	}

	pr, err := githubprdescriptor.ParseURL(githubLink)
	if err != nil {
		return err
	}

	discussion, err := r.githubClient.FetchDiscussion(ctx, pr)
	if err != nil {
		return err
	}

	summary, err := summarizePullRequestDiscussion(ctx, r.analyzer, discussion)
	if err != nil {
		return err
	}

	if err := r.fileHandler.Replace(outputPath, []byte(summary)); err != nil {
		return fmt.Errorf("write PR summary: %w", err)
	}

	return nil
}

func (r *PRReviewer) WriteReviewPlanMarkdownToDefaultFile(ctx context.Context, githubLink string) (string, error) {
	pr, err := githubprdescriptor.ParseURL(githubLink)
	if err != nil {
		return "", err
	}

	outputDir := strings.TrimSpace(r.config.OutputDir)
	if outputDir == "" {
		outputDir = "reports"
	}
	outputPath := filepath.Join(outputDir, fmt.Sprintf("pr-review-%s-%s-%d.md", sanitizeFilePart(pr.Owner), sanitizeFilePart(pr.Repo), pr.Number))
	if err := r.WriteReviewPlanMarkdown(ctx, githubLink, outputPath); err != nil {
		return "", err
	}

	return outputPath, nil
}

func sanitizeFilePart(value string) string {
	value = strings.TrimSpace(value)
	var b strings.Builder
	for _, r := range value {
		switch {
		case r >= 'a' && r <= 'z':
			b.WriteRune(r)
		case r >= 'A' && r <= 'Z':
			b.WriteRune(r)
		case r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == '-' || r == '_' || r == '.':
			b.WriteRune(r)
		default:
			b.WriteByte('-')
		}
	}
	if b.Len() == 0 {
		return "unknown"
	}
	return b.String()
}

func summarizePullRequestDiscussion(ctx context.Context, analyzer llm.Analyzer, discussion githubprdescriptor.PRDiscussion) (string, error) {
	prompt := buildPullRequestSummaryPrompt(discussion)
	return analyzer.AnalyzePRandReviewPlan(ctx, prompt)
}

func buildPullRequestSummaryPrompt(discussion githubprdescriptor.PRDiscussion) string {
	var b strings.Builder
	b.WriteString("Review this GitHub pull request discussion and summarize what is happening.\n")
	b.WriteString("Write concise Markdown with: Overview, Key discussion points, Fix plan, Open discussion or questions, and Current status.\n")
	b.WriteString("For every clearly actionable request, include a concrete fix plan with execution steps. Leave unclear, disputed, or unanswered items in Open discussion or questions instead of pretending they are ready to execute.\n")
	b.WriteString("Focus on the discussion, reviewer feedback, disagreements, and unresolved questions. Do not invent facts.\n\n")
	fmt.Fprintf(&b, "PR: %s\nTitle: %s\nState: %s\nAuthor: %s\n\n", discussion.URL, discussion.Title, discussion.State, discussion.Author)
	if strings.TrimSpace(discussion.Body) != "" {
		fmt.Fprintf(&b, "Description:\n%s\n\n", truncateForPrompt(discussion.Body))
	}

	b.WriteString("Issue comments:\n")
	for _, comment := range discussion.IssueComments {
		fmt.Fprintf(&b, "- %s at %s: %s\n", comment.User.Login, comment.CreatedAt.Format(time.RFC3339), truncateForPrompt(comment.Body))
	}
	if len(discussion.IssueComments) == 0 {
		b.WriteString("- none\n")
	}

	b.WriteString("\nReview summaries:\n")
	for _, review := range discussion.Reviews {
		fmt.Fprintf(&b, "- %s %s at %s: %s\n", review.User.Login, review.State, review.SubmittedAt.Format(time.RFC3339), truncateForPrompt(review.Body))
	}
	if len(discussion.Reviews) == 0 {
		b.WriteString("- none\n")
	}

	b.WriteString("\nReview comments:\n")
	for _, comment := range discussion.ReviewComments {
		fmt.Fprintf(&b, "- %s on %s at %s: %s\n", comment.User.Login, comment.Path, comment.CreatedAt.Format(time.RFC3339), truncateForPrompt(comment.Body))
	}
	if len(discussion.ReviewComments) == 0 {
		b.WriteString("- none\n")
	}

	return b.String()
}

func truncateForPrompt(value string) string {
	value = strings.TrimSpace(value)
	const limit = 4000
	if len(value) <= limit {
		return value
	}
	return value[:limit] + "\n[truncated]"
}

func (AtomicFileHandler) Replace(path string, data []byte) error {
	return replaceFile(path, data)
}

func replaceFile(path string, data []byte) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create summary directory: %w", err)
	}

	tmpFile, err := os.CreateTemp(dir, "."+filepath.Base(path)+".tmp-*")
	if err != nil {
		return err
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)

	if _, err := tmpFile.Write(data); err != nil {
		_ = tmpFile.Close()
		return err
	}
	if err := tmpFile.Close(); err != nil {
		return err
	}

	return os.Rename(tmpPath, path)
}

package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"auto-pr-review/internal/config"
	"auto-pr-review/internal/service"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: go run scripts/integration-pr-review.go <github-pr-url>")
		os.Exit(2)
	}

	if err := loadDotEnv(".env"); err != nil {
		fmt.Fprintf(os.Stderr, "load .env: %v\n", err)
		os.Exit(1)
	}

	cfg, err := config.LoadPRReviewer()
	if err != nil {
		fmt.Fprintf(os.Stderr, "load config: %v\n", err)
		os.Exit(1)
	}

	reviewer, err := service.NewPRReviewer(service.PRDiscussionReviewConfig{
		LLM:         cfg.LLM,
		GitHubToken: cfg.GitHub.Token,
		OutputDir:   cfg.PRReview.OutputDir,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "create PR reviewer: %v\n", err)
		os.Exit(1)
	}

	outputPath, err := reviewer.WriteReviewPlanMarkdownToDefaultFile(context.Background(), os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "write PR review: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(outputPath)
}

func loadDotEnv(path string) error {
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if key == "" || value == "" {
			continue
		}
		if err := os.Setenv(key, value); err != nil {
			return err
		}
	}

	return scanner.Err()
}

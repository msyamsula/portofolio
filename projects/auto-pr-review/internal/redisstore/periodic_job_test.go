package redisstore

import (
	"testing"

	"auto-pr-review/internal/service"
)

func TestPeriodicJobMember(t *testing.T) {
	t.Parallel()

	got := periodicJobMember(service.PeriodicJob{
		GitHubLink: "https://github.com/octocat/Hello-World",
		Period:     "5s",
	})
	want := "https://github.com/octocat/Hello-World:5s"

	if got != want {
		t.Fatalf("member mismatch: got %q want %q", got, want)
	}
}

func TestParsePeriodicJobMember(t *testing.T) {
	t.Parallel()

	job, period, err := parsePeriodicJobMember("https://github.com/octocat/Hello-World:10m")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if job.GitHubLink != "https://github.com/octocat/Hello-World" {
		t.Fatalf("github link mismatch: got %q", job.GitHubLink)
	}
	if job.Period != "10m0s" {
		t.Fatalf("period mismatch: got %q", job.Period)
	}
	if period.String() != "10m0s" {
		t.Fatalf("duration mismatch: got %q", period)
	}
}

func TestParsePeriodicJobMemberRejectsInvalidMember(t *testing.T) {
	t.Parallel()

	_, _, err := parsePeriodicJobMember("https://github.com/octocat/Hello-World")
	if err == nil {
		t.Fatal("expected error")
	}
}

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

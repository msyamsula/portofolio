package service

import (
	"context"
	"errors"
	"testing"
	"time"
)

type periodicJobRepoStub struct {
	job PeriodicJob
	err error
}

func (r *periodicJobRepoStub) RegisterPeriodicJob(_ context.Context, job PeriodicJob) error {
	r.job = job
	return r.err
}

func TestPeriodicJobServiceRegister(t *testing.T) {
	t.Parallel()

	repo := &periodicJobRepoStub{}
	svc := NewPeriodicJobService(repo)
	svc.now = func() time.Time {
		return time.Unix(100, 0).UTC()
	}

	job, err := svc.Register(context.Background(), "https://github.com/octocat/Hello-World", "5s")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if job.GitHubLink != "https://github.com/octocat/Hello-World" {
		t.Fatalf("github link mismatch: got %q", job.GitHubLink)
	}
	if job.Period != "5s" {
		t.Fatalf("period mismatch: got %q", job.Period)
	}
	if job.NextTimestamp != 105 {
		t.Fatalf("next timestamp mismatch: got %d", job.NextTimestamp)
	}
	if repo.job != job {
		t.Fatalf("repository job mismatch: got %+v want %+v", repo.job, job)
	}
}

func TestPeriodicJobServiceRegisterRejectsInvalidPeriod(t *testing.T) {
	t.Parallel()

	svc := NewPeriodicJobService(&periodicJobRepoStub{})

	_, err := svc.Register(context.Background(), "https://github.com/octocat/Hello-World", "0s")
	if !errors.Is(err, ErrInvalidJobPeriod) {
		t.Fatalf("expected ErrInvalidJobPeriod, got %v", err)
	}
}

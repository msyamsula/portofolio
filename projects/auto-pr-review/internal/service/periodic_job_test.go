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

func (r *periodicJobRepoStub) SetPeriodicJob(_ context.Context, githubURL, period string) (PeriodicJob, error) {
	r.job = PeriodicJob{GitHubLink: githubURL, Period: period}
	return r.job, r.err
}

func (r *periodicJobRepoStub) ListPeriodicJobs(_ context.Context) ([]PeriodicJob, error) {
	return []PeriodicJob{r.job}, r.err
}

func (r *periodicJobRepoStub) ClearPeriodicJob(_ context.Context, githubURL, period string) error {
	r.job = PeriodicJob{GitHubLink: githubURL, Period: period}
	return r.err
}

type periodicJobSchedulerStub struct {
	job PeriodicJob
	err error
}

func (s *periodicJobSchedulerStub) RegisterPeriodicJob(_ context.Context, job PeriodicJob) error {
	s.job = job
	return s.err
}

func (s *periodicJobSchedulerStub) DeregisterPeriodicJob(_ context.Context, job PeriodicJob) error {
	s.job = job
	return s.err
}

func (r *periodicJobRepoStub) RegisterPeriodicJob(_ context.Context, job PeriodicJob) error {
	r.job = job
	return r.err
}

func TestPeriodicJobServiceRegister(t *testing.T) {
	t.Parallel()

	repo := &periodicJobRepoStub{}
	scheduler := &periodicJobSchedulerStub{}
	svc := NewPeriodicJobService(repo, scheduler)
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
	wantRepoJob := PeriodicJob{GitHubLink: job.GitHubLink, Period: job.Period}
	if repo.job != wantRepoJob {
		t.Fatalf("repository job mismatch: got %+v want %+v", repo.job, wantRepoJob)
	}
	if scheduler.job != job {
		t.Fatalf("scheduler job mismatch: got %+v want %+v", scheduler.job, job)
	}
}

func TestPeriodicJobServiceRegisterRejectsInvalidPeriod(t *testing.T) {
	t.Parallel()

	svc := NewPeriodicJobService(&periodicJobRepoStub{}, &periodicJobSchedulerStub{})

	_, err := svc.Register(context.Background(), "https://github.com/octocat/Hello-World", "0s")
	if !errors.Is(err, ErrInvalidJobPeriod) {
		t.Fatalf("expected ErrInvalidJobPeriod, got %v", err)
	}
}

package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

var ErrInvalidJobPeriod = errors.New("invalid job period")

type PeriodicJob struct {
	GitHubLink     string `json:"github_link"`
	Period         string `json:"period"`
	NextTimestamp  int64  `json:"next_timestamp"`
	IdempotencyKey string `json:"idempotency_key,omitempty"`
}

type PeriodicJobRepository interface {
	ListPeriodicJobs(ctx context.Context) ([]PeriodicJob, error)
	SetPeriodicJob(ctx context.Context, githubURL, period string) (PeriodicJob, error)
	ClearPeriodicJob(ctx context.Context, githubURL, period string) error
}

type ScheduleManager interface {
	SetSchedule(ctx context.Context, job PeriodicJob) error
	RemoveSchedule(ctx context.Context, job PeriodicJob) error
}

type PeriodicJobService struct {
	repo      PeriodicJobRepository
	scheduler ScheduleManager
	now       func() time.Time
}

func NewPeriodicJobService(repo PeriodicJobRepository, scheduler ScheduleManager) *PeriodicJobService {
	return &PeriodicJobService{
		repo:      repo,
		scheduler: scheduler,
		now:       time.Now,
	}
}

func (s *PeriodicJobService) Register(ctx context.Context, rawURL, rawPeriod string) (PeriodicJob, error) {
	job, _, err := s.buildJob(rawURL, rawPeriod)
	if err != nil {
		return PeriodicJob{}, err
	}

	currentJobs, err := s.repo.ListPeriodicJobs(ctx)
	if err != nil {
		return PeriodicJob{}, err
	}

	storedJob, err := s.repo.SetPeriodicJob(ctx, job.GitHubLink, job.Period)
	if err != nil {
		return PeriodicJob{}, err
	}
	storedJob.NextTimestamp = job.NextTimestamp

	for _, currentJob := range currentJobs {
		if currentJob.GitHubLink == storedJob.GitHubLink && currentJob.Period != storedJob.Period {
			if err := s.scheduler.RemoveSchedule(ctx, currentJob); err != nil {
				return PeriodicJob{}, err
			}
			break
		}
	}

	if err := s.scheduler.SetSchedule(ctx, storedJob); err != nil {
		return PeriodicJob{}, err
	}

	return storedJob, nil
}

func (s *PeriodicJobService) List(ctx context.Context) ([]PeriodicJob, error) {
	return s.repo.ListPeriodicJobs(ctx)
}

func (s *PeriodicJobService) Deregister(ctx context.Context, rawURL, rawPeriod string) error {
	job, _, err := s.buildJob(rawURL, rawPeriod)
	if err != nil {
		return err
	}

	if err := s.repo.ClearPeriodicJob(ctx, job.GitHubLink, job.Period); err != nil {
		return err
	}

	return s.scheduler.RemoveSchedule(ctx, job)
}

func (s *PeriodicJobService) buildJob(rawURL, rawPeriod string) (PeriodicJob, time.Duration, error) {
	githubURL, err := NormalizeGitHubURL(rawURL)
	if err != nil {
		return PeriodicJob{}, 0, err
	}

	rawPeriod = strings.TrimSpace(rawPeriod)
	period, err := time.ParseDuration(rawPeriod)
	if err != nil || period < time.Second {
		return PeriodicJob{}, 0, fmt.Errorf("%w: must be at least 1s, like 5s, 10m, or 1h", ErrInvalidJobPeriod)
	}

	job := PeriodicJob{
		GitHubLink:    githubURL,
		Period:        rawPeriod,
		NextTimestamp: s.now().Add(period).Unix(),
	}

	return job, period, nil
}

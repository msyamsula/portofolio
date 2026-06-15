package scheduler

import (
	"context"
	"errors"
	"log/slog"

	"auto-pr-review/internal/service"
)

type PeriodicJobLookup interface {
	GetPeriodicJob(ctx context.Context, githubURL string) (service.PeriodicJob, error)
}

type PeriodicJobScheduler interface {
	RegisterPeriodicJob(ctx context.Context, job service.PeriodicJob) error
	DeregisterPeriodicJob(ctx context.Context, job service.PeriodicJob) (bool, error)
}

type Scheduler struct {
	repo   PeriodicJobLookup
	redis  PeriodicJobScheduler
	logger *slog.Logger
}

func New(repo PeriodicJobLookup, redis PeriodicJobScheduler, logger *slog.Logger) *Scheduler {
	return &Scheduler{repo: repo, redis: redis, logger: logger}
}

func (s *Scheduler) SetSchedule(ctx context.Context, job service.PeriodicJob) error {
	_, err := s.repo.GetPeriodicJob(ctx, job.GitHubLink)
	if err != nil {
		if errors.Is(err, service.ErrGitHubLinkNotFound) {
			s.logger.Info("schedule not set, job has been deregistered",
				slog.String("github_link", job.GitHubLink),
				slog.String("period", job.Period))
			return nil
		}
		return err
	}

	return s.redis.RegisterPeriodicJob(ctx, job)
}

func (s *Scheduler) RemoveSchedule(ctx context.Context, job service.PeriodicJob) error {
	removed, err := s.redis.DeregisterPeriodicJob(ctx, job)
	if err != nil {
		return err
	}

	if !removed {
		s.logger.Info("schedule not removed, job was not registered",
			slog.String("github_link", job.GitHubLink),
			slog.String("period", job.Period))
	}

	return nil
}

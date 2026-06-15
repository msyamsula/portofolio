package scheduler

import (
	"context"
	"errors"
	"log/slog"
	"time"

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
	now    func() time.Time
}

func New(repo PeriodicJobLookup, redis PeriodicJobScheduler, logger *slog.Logger) *Scheduler {
	return &Scheduler{repo: repo, redis: redis, logger: logger, now: time.Now}
}

func (s *Scheduler) SetSchedule(ctx context.Context, job service.PeriodicJob) error {
	pgJob, err := s.repo.GetPeriodicJob(ctx, job.GitHubLink)
	if err != nil {
		if errors.Is(err, service.ErrGitHubLinkNotFound) {
			s.logger.Info("schedule not set, job has been deregistered",
				slog.String("github_link", job.GitHubLink))
			return nil
		}
		return err
	}

	job.Period = pgJob.Period

	if job.NextTimestamp == 0 {
		period, err := time.ParseDuration(pgJob.Period)
		if err != nil {
			return err
		}
		job.NextTimestamp = s.now().Add(period).Unix()
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
			slog.String("github_link", job.GitHubLink))
	}

	return nil
}

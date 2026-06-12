package service

import (
	"context"
	"errors"
	"fmt"
	"time"
)

var ErrInvalidJobPeriod = errors.New("invalid job period")

type PeriodicJob struct {
	GitHubLink    string `json:"github_link"`
	Period        string `json:"period"`
	NextTimestamp int64  `json:"next_timestamp"`
}

type PeriodicJobRepository interface {
	RegisterPeriodicJob(ctx context.Context, job PeriodicJob) error
}

type PeriodicJobService struct {
	repo PeriodicJobRepository
	now  func() time.Time
}

func NewPeriodicJobService(repo PeriodicJobRepository) *PeriodicJobService {
	return &PeriodicJobService{
		repo: repo,
		now:  time.Now,
	}
}

func (s *PeriodicJobService) Register(ctx context.Context, rawURL, rawPeriod string) (PeriodicJob, error) {
	githubURL, err := NormalizeGitHubURL(rawURL)
	if err != nil {
		return PeriodicJob{}, err
	}

	period, err := time.ParseDuration(rawPeriod)
	if err != nil || period < time.Second {
		return PeriodicJob{}, fmt.Errorf("%w: must be at least 1s, like 5s, 10m, or 1h", ErrInvalidJobPeriod)
	}

	job := PeriodicJob{
		GitHubLink:    githubURL,
		Period:        period.String(),
		NextTimestamp: s.now().Add(period).Unix(),
	}
	if err := s.repo.RegisterPeriodicJob(ctx, job); err != nil {
		return PeriodicJob{}, err
	}

	return job, nil
}

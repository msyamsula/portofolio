package orchestrator

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"auto-pr-review/internal/service"
)

type periodicJobClaimerStub struct {
	jobs []service.PeriodicJob
}

func (c *periodicJobClaimerStub) ClaimDuePeriodicJob(context.Context, time.Time) (service.PeriodicJob, bool, error) {
	if len(c.jobs) == 0 {
		return service.PeriodicJob{}, false, nil
	}

	job := c.jobs[0]
	c.jobs = c.jobs[1:]
	return job, true, nil
}

func TestPeriodicJobOrchestratorAddsIdempotencyKeyPerDispatch(t *testing.T) {
	t.Parallel()

	claimer := &periodicJobClaimerStub{
		jobs: []service.PeriodicJob{
			{
				GitHubLink:    "https://github.com/octocat/Hello-World",
				Period:        "10m",
				NextTimestamp: 123,
			},
			{
				GitHubLink:    "https://github.com/octocat/Hello-World",
				Period:        "10m",
				NextTimestamp: 456,
			},
		},
	}
	keys := []string{"dispatch-key-1", "dispatch-key-2"}
	var dispatched []service.PeriodicJob
	orchestrator := NewPeriodicJobOrchestrator(
		claimer,
		time.Second,
		slog.Default(),
		func(_ context.Context, job service.PeriodicJob) error {
			dispatched = append(dispatched, job)
			return nil
		},
	)
	orchestrator.idempotencyKey = func() (string, error) {
		key := keys[0]
		keys = keys[1:]
		return key, nil
	}

	orchestrator.runOnce(context.Background())
	orchestrator.runOnce(context.Background())

	if len(dispatched) != 2 {
		t.Fatalf("dispatch count mismatch: got %d want 2", len(dispatched))
	}
	if dispatched[0].IdempotencyKey != "dispatch-key-1" {
		t.Fatalf("first idempotency key mismatch: got %q", dispatched[0].IdempotencyKey)
	}
	if dispatched[1].IdempotencyKey != "dispatch-key-2" {
		t.Fatalf("second idempotency key mismatch: got %q", dispatched[1].IdempotencyKey)
	}
}

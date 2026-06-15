package orchestrator

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"auto-pr-review/internal/service"

	"github.com/redis/go-redis/v9"
)

func evalStub(jobs []service.PeriodicJob) func(context.Context, string, []string, ...interface{}) *redis.Cmd {
	i := 0
	return func(ctx context.Context, script string, keys []string, args ...interface{}) *redis.Cmd {
		if i >= len(jobs) {
			cmd := redis.NewCmd(ctx)
			cmd.SetVal("")
			return cmd
		}
		job := jobs[i]
		i++
		cmd := redis.NewCmd(ctx)
		cmd.SetVal(service.PeriodicJobMember(job))
		return cmd
	}
}

func TestPeriodicJobOrchestratorAddsIdempotencyKeyPerDispatch(t *testing.T) {
	t.Parallel()

	jobs := []service.PeriodicJob{
		{
			GitHubLink:    "https://github.com/octocat/Hello-World",
			NextTimestamp: 123,
		},
		{
			GitHubLink:    "https://github.com/octocat/Spoon-Knife",
			NextTimestamp: 456,
		},
	}
	keys := []string{"dispatch-key-1", "dispatch-key-2"}
	var dispatched []service.PeriodicJob
	orchestrator := NewPeriodicJobOrchestrator(
		evalStub(jobs),
		"test:key",
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

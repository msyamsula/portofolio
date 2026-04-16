package reconciler

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"message/config"
	"message/domain"
	"message/repository"
	"message/service"
)

// Reconciler sweeps the event log on a fixed interval and re-drives messages
// that are stuck (never processed) or failing (repeated reconcile attempts).
//
// It is the system's convergence guarantee: even if the Kafka consumer crashes
// or a transaction partially fails, every message eventually reaches SUCCESS
// or terminal FAILED within a bounded number of reconcile cycles.
type Reconciler struct {
	eventLog     *repository.EventLogRepo
	conversation *repository.ConversationRepo
	processor    *service.Processor
	cfg          config.Config
}

// New creates a Reconciler.
func New(
	eventLog *repository.EventLogRepo,
	conversation *repository.ConversationRepo,
	processor *service.Processor,
	cfg config.Config,
) *Reconciler {
	return &Reconciler{
		eventLog:     eventLog,
		conversation: conversation,
		processor:    processor,
		cfg:          cfg,
	}
}

// Run starts the reconciler loop. Ticks at cfg.ReconcileInterval.
// Blocks until ctx is cancelled.
func (r *Reconciler) Run(ctx context.Context) {
	ticker := time.NewTicker(r.cfg.ReconcileInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := r.ReconcileOnce(ctx); err != nil {
				slog.Error("reconcile cycle error", "err", err)
			}
		}
	}
}

// ReconcileOnce executes one full reconciliation sweep.
// Per-message errors are logged and skipped — the next tick retries them.
func (r *Reconciler) ReconcileOnce(ctx context.Context) error {
	refs, err := r.eventLog.GetStuck(ctx)
	if err != nil {
		return fmt.Errorf("reconcile get_stuck: %w", err)
	}
	for _, ref := range refs {
		if err := r.reconcileOne(ctx, ref); err != nil {
			slog.Error("reconcile one failed",
				"message_id", ref.MessageID,
				"conversation_id", ref.ConversationID,
				"err", err,
			)
		}
	}
	return nil
}

// reconcileOne processes a single stuck message.
// It constructs current state, then either marks it terminal FAILED (if limits
// are exhausted) or appends a reconcile event and directly calls ProcessSuccess.
func (r *Reconciler) reconcileOne(ctx context.Context, ref repository.StuckRef) error {
	events, err := r.eventLog.GetByMessage(ctx, ref.MessageID, ref.ConversationID)
	if err != nil {
		return err
	}
	inConv, err := r.conversation.Exists(ctx, ref.MessageID, ref.ConversationID)
	if err != nil {
		return err
	}

	state := domain.BuildState(events, inConv)
	if state.Current == domain.StateSuccess || state.Current == domain.StateFailed {
		return nil // GetStuck query was slightly stale; nothing to do
	}

	if r.isExhausted(state) {
		return r.appendEvent(ctx, ref.MessageID, ref.ConversationID, state, domain.EventFailed)
	}

	return r.appendAndProcess(ctx, ref.MessageID, ref.ConversationID, state)
}

// isExhausted returns true when both retry and attempt budgets are used up.
//
// Phase 1 (stuck):  RECONCILE_STUCKED_MESSAGE × MaxRetry
// Phase 2 (failing): RECONCILE_FAILED_MESSAGE  × MaxAttempt
// Phase 3: terminal FAILED
func (r *Reconciler) isExhausted(state domain.MessageState) bool {
	return state.RetryCount >= r.cfg.MaxRetry && state.AttemptCount >= r.cfg.MaxAttempt
}

// chooseEvent selects the reconcile event name based on how many retries
// have already been attempted.
//
//   RetryCount < MaxRetry  → RECONCILE_STUCKED_MESSAGE  (initial stuck retries)
//   RetryCount >= MaxRetry → RECONCILE_FAILED_MESSAGE   (escalated failure retries)
func (r *Reconciler) chooseEvent(state domain.MessageState) domain.EventName {
	if state.RetryCount < r.cfg.MaxRetry {
		return domain.EventReconcileStuck
	}
	return domain.EventReconcileFailed
}

// appendAndProcess appends the reconcile event to the log (for audit/observability),
// then directly calls ProcessSuccess to attempt the b→d transition.
//
// The reconcile event flows to Kafka via the outbox relay asynchronously.
// We don't wait for it — the reconciler drives processing directly to
// avoid adding another round-trip through the Kafka pipeline.
func (r *Reconciler) appendAndProcess(ctx context.Context, mid, cid uuid.UUID, state domain.MessageState) error {
	eventName := r.chooseEvent(state)
	if err := r.appendEvent(ctx, mid, cid, state, eventName); err != nil {
		return err
	}
	// Re-read state after appending reconcile event and call processor.
	// ProcessSuccess re-reads events from DB internally, so we pass minimal fields.
	synth := domain.EventLog{MessageID: mid, ConversationID: cid}
	return r.processor.ProcessSuccess(ctx, synth)
}

// appendEvent writes a single event atomically with ON CONFLICT DO NOTHING.
// If another reconciler instance wins the version slot simultaneously,
// this is a silent no-op — the next tick will re-evaluate.
func (r *Reconciler) appendEvent(
	ctx context.Context,
	mid, cid uuid.UUID,
	state domain.MessageState,
	name domain.EventName,
) error {
	newVersion := state.Version + 1
	event := domain.EventLog{
		EventID:        domain.HashEventID(mid, cid, newVersion),
		MessageID:      mid,
		ConversationID: cid,
		SenderID:       state.SenderID,
		ReceiverID:     state.ReceiverID,
		Version:        newVersion,
		EventName:      name,
		Payload:        state.Payload,
	}

	tx, err := r.eventLog.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("reconciler begin tx: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	inserted, err := r.eventLog.Append(ctx, tx, event)
	if err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("reconciler commit: %w", err)
	}
	if !inserted {
		// A concurrent reconciler instance won this version slot.
		// The message is being handled; next tick will re-evaluate if still needed.
		slog.Info("reconciler version conflict (peer won)", "message_id", mid, "version", newVersion)
	}
	return nil
}

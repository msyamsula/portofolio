package domain

import "github.com/google/uuid"

// State is the current position of a message in the state machine.
// Never stored — always derived by replaying the event log.
type State int

const (
	StateUnknown State = iota

	// StateSent: SENT event exists; message not yet delivered.
	StateSent

	// StateSuccess: conversation row exists. Terminal success.
	StateSuccess

	// StateFailed: FAILED event exists. Terminal failure. No more retries.
	StateFailed
)

// MessageState is the result of replaying a message's event log.
// It carries all information needed to decide the next action.
type MessageState struct {
	Current State

	// Version is the highest version seen in the log.
	// The next append must use Version+1.
	Version int

	// RetryCount is the number of RECONCILE_STUCKED_MESSAGE events.
	// When RetryCount >= config.MaxRetry the reconciler escalates to RECONCILE_FAILED_MESSAGE.
	RetryCount int

	// AttemptCount is the number of RECONCILE_FAILED_MESSAGE events.
	// When AttemptCount >= config.MaxAttempt the reconciler writes terminal FAILED.
	AttemptCount int

	// SenderID, ReceiverID, Payload are captured from the first SENT event.
	// They are needed when the reconciler re-drives processing.
	SenderID   uuid.UUID
	ReceiverID uuid.UUID
	Payload    []byte
}

// BuildState replays events in ascending version order and applies state machine
// transitions to produce the current MessageState.
//
// inConversation is a separate DB lookup — it is the authoritative signal for
// StateSuccess. We do not rely solely on the event log because a transaction
// that wrote SUCCESS may have committed the conversation row but not had its
// Kafka message acked, causing re-delivery of an already-done message.
func BuildState(events []EventLog, inConversation bool) MessageState {
	var s MessageState
	for _, e := range events {
		if e.Version > s.Version {
			s.Version = e.Version
		}
		switch e.EventName {
		case EventSent:
			s.Current = StateSent
			s.SenderID = e.SenderID
			s.ReceiverID = e.ReceiverID
			s.Payload = e.Payload
		case EventSuccess:
			s.Current = StateSuccess
		case EventFailed:
			s.Current = StateFailed
		case EventReconcileStuck:
			s.RetryCount++
		case EventReconcileFailed:
			s.AttemptCount++
		}
	}
	if inConversation {
		s.Current = StateSuccess
	}
	return s
}

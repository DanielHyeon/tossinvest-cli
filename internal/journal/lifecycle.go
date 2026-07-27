package journal

import (
	"errors"
	"fmt"
)

// ErrIllegalTransition means the requested state change is not in the lifecycle
// table. It is not a retryable condition: it means the caller's model of the
// attempt and the journal's record disagree.
var ErrIllegalTransition = errors.New("journal: illegal attempt state transition")

// stateNone is the pseudo-state an attempt has before it is written.
const stateNone AttemptState = ""

// transitionTable is the MutationAttempt lifecycle from the order-execution spec.
//
//	(new) ──► RECORDED ──► DISPATCH_STARTED ──► ACKED ──► CONFIRMED
//	              │               │               │
//	              │               ├──► IN_DOUBT ◄─┘
//	              │               ├──► NOT_DISPATCHED
//	              ▼               └──► FAILED_CONFIRMED
//	        NOT_DISPATCHED
//
// The edges that are deliberately absent are the interesting part:
//
//   - RECORDED cannot reach CONFIRMED, IN_DOUBT or FAILED_CONFIRMED. Nothing was
//     sent, so there is nothing to confirm, doubt or have rejected. The only close
//     is NOT_DISPATCHED, which is also what a restart applies.
//   - DISPATCH_STARTED cannot reach CONFIRMED directly. Confirmation goes through
//     an ACK (or through IN_DOUBT resolution), so "confirmed" always has a
//     traceable source.
//   - IN_DOUBT cannot reach NOT_DISPATCHED. Once the request may have left, "it
//     never happened" is not something we can assert; proving absence lands on
//     FAILED_CONFIRMED, and failing to prove anything lands on
//     UNRESOLVED_IN_DOUBT.
//   - CONFIRMED, NOT_DISPATCHED and FAILED_CONFIRMED have no outgoing edges at
//     all. UNRESOLVED_IN_DOUBT keeps two, reachable only through an explicit
//     operator resolution API (task 2.7); no automatic path uses them, and Settle
//     cannot because it refuses to act on a terminal attempt.
var transitionTable = map[AttemptState][]AttemptState{
	stateNone:              {StateRecorded},
	StateRecorded:          {StateDispatchStarted, StateNotDispatched},
	StateDispatchStarted:   {StateAcked, StateInDoubt, StateNotDispatched, StateFailedConfirmed},
	StateAcked:             {StateConfirmed, StateInDoubt},
	StateInDoubt:           {StateConfirmed, StateFailedConfirmed, StateUnresolvedInDoubt},
	StateUnresolvedInDoubt: {StateConfirmed, StateFailedConfirmed},
	StateConfirmed:         nil,
	StateNotDispatched:     nil,
	StateFailedConfirmed:   nil,
}

// ValidateTransition reports whether from → to is in the lifecycle table.
func ValidateTransition(from, to AttemptState) error {
	allowed, known := transitionTable[from]
	if !known {
		return fmt.Errorf("%w: unknown state %q", ErrIllegalTransition, from)
	}
	if _, knownTo := transitionTable[to]; !knownTo || to == stateNone {
		return fmt.Errorf("%w: unknown target state %q", ErrIllegalTransition, to)
	}
	for _, candidate := range allowed {
		if candidate == to {
			return nil
		}
	}
	return fmt.Errorf("%w: %s cannot move to %s", ErrIllegalTransition, from, to)
}

// AllowedNext returns the states an attempt may move to next. Terminal states
// return nil, except UNRESOLVED_IN_DOUBT (see transitionTable).
func AllowedNext(from AttemptState) []AttemptState {
	allowed := transitionTable[from]
	if len(allowed) == 0 {
		return nil
	}
	out := make([]AttemptState, len(allowed))
	copy(out, allowed)
	return out
}

// Reason codes the lifecycle itself writes. They are stable strings: operators and
// the alerting path read them.
const (
	// ReasonRestartNotDispatched closes an attempt that never got past RECORDED.
	ReasonRestartNotDispatched = "restart_not_dispatched"
	// ReasonRestartInDoubt marks an attempt that was mid-dispatch when the
	// process died.
	ReasonRestartInDoubt = "restart_in_doubt"

	// ReasonDispatchNotSent is a provably unsent mutation.
	ReasonDispatchNotSent = "dispatch_not_sent"
	// ReasonDispatchRejected is a well-formed broker refusal.
	ReasonDispatchRejected = "dispatch_rejected_by_broker"
	// ReasonDispatchAmbiguous is the "we cannot tell" outcome.
	ReasonDispatchAmbiguous = "dispatch_outcome_unknown"
	// ReasonBrokerAcknowledged is a broker ack that named an order.
	ReasonBrokerAcknowledged = "broker_acknowledged"
	// ReasonAckWithoutOrderID is an ack we cannot track: no order number came
	// back, so the mutation may exist and be unfindable by id.
	ReasonAckWithoutOrderID = "ack_without_order_id"
	// ReasonUnknownDispatchClass is a dispatch function returning something this
	// build does not classify.
	ReasonUnknownDispatchClass = "unknown_dispatch_class"
)

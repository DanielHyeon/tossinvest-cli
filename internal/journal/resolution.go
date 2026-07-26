package journal

// resolution.go is the write side of IN_DOUBT resolution (harden-execution-base
// task 2.7). It is a new file rather than an edit to durability.go: the primitives
// there are released and their tests pin them, and everything below is additive.
//
// Two things exist here that Settle deliberately cannot do:
//
//   - confirming an attempt *and* recording the broker order id the resolution
//     discovered, in one transaction. Settle takes no order id, and a CONFIRMED
//     attempt whose order number is only in a log line is not trackable.
//   - moving out of UNRESOLVED_IN_DOUBT. Settle refuses that state on purpose
//     (see lifecycle.go); the only way out is an operator saying what happened.

import (
	"context"
	"fmt"
	"strings"
)

// Reason codes written by the resolution procedure. Stable strings: alerting and
// the Phase 2 ledger read them.
const (
	// ReasonResolvedFound: the mutation was found at the broker.
	ReasonResolvedFound = "in_doubt_resolved_found"
	// ReasonResolvedAbsent: the mutation was proven absent — stable repeated
	// observation plus an account cross-check, never a single empty list.
	ReasonResolvedAbsent = "in_doubt_resolved_absent"
	// ReasonResolutionExhausted: neither presence nor absence could be proven.
	// The attempt is parked for an operator and its symbol stays blocked.
	ReasonResolutionExhausted = "in_doubt_unresolved"
	// ReasonOperatorResolved: a human closed the attempt.
	ReasonOperatorResolved = "operator_resolved"
)

// ResolveConfirmed closes an IN_DOUBT (or stuck ACKED) attempt as CONFIRMED,
// recording the broker order id the resolution found.
//
// The order id is mandatory: "it exists but we do not know its number" is not a
// resolution, because nothing can later cancel or reconcile it.
//
// It is stored exactly as the broker sent it. `orderId` is an opaque token —
// openapi contracts no shape for it — so trimming, case-folding or any other
// normalisation on the way in would store an identifier the broker never issued,
// and a later byte comparison against the real one would silently fail to match.
// Trimming survives only in the emptiness check, where it asks "is there a name
// here at all".
func (a *Attempt) ResolveConfirmed(ctx context.Context, brokerOrderID, reasonCode, detail string) error {
	if strings.TrimSpace(brokerOrderID) == "" {
		return fmt.Errorf("%w: confirming an attempt requires the broker order id", ErrInvalidRequest)
	}
	return a.transition(ctx, StateConfirmed, transitionOpts{
		from:          []AttemptState{StateInDoubt, StateAcked},
		brokerOrderID: brokerOrderID,
		reasonCode:    firstNonEmpty(reasonCode, ReasonResolvedFound),
		detail:        detail,
		setSettled:    true,
	})
}

// ResolveFailed closes an IN_DOUBT attempt as FAILED_CONFIRMED: the mutation was
// proven not to have happened.
func (a *Attempt) ResolveFailed(ctx context.Context, reasonCode, detail string) error {
	return a.transition(ctx, StateFailedConfirmed, transitionOpts{
		from:       []AttemptState{StateInDoubt},
		reasonCode: firstNonEmpty(reasonCode, ReasonResolvedAbsent),
		detail:     detail,
		setSettled: true,
	})
}

// ResolveUnresolved parks an IN_DOUBT attempt as UNRESOLVED_IN_DOUBT.
//
// This is the fail-closed outcome, not an error path: the symbol stays blocked
// until an operator resolves it, which is the correct behaviour when a live
// account's state cannot be established.
func (a *Attempt) ResolveUnresolved(ctx context.Context, reasonCode, detail string) error {
	return a.transition(ctx, StateUnresolvedInDoubt, transitionOpts{
		from:       []AttemptState{StateInDoubt},
		reasonCode: firstNonEmpty(reasonCode, ReasonResolutionExhausted),
		detail:     detail,
		setSettled: true,
	})
}

// UnresolvedAttempts returns every attempt parked as UNRESOLVED_IN_DOUBT, oldest
// first.
//
// PendingAttempts deliberately does not include them — UNRESOLVED_IN_DOUBT is a
// terminal state — but they still block their symbol, and that block has to
// survive a restart. This is how a caller finds them again.
func (j *Journal) UnresolvedAttempts(ctx context.Context) ([]AttemptRecord, error) {
	rows, err := j.db.QueryContext(ctx, attemptSelect+
		` WHERE state = ? ORDER BY recorded_at, rowid`, string(StateUnresolvedInDoubt))
	if err != nil {
		return nil, fmt.Errorf("journal: listing unresolved attempts: %w", err)
	}
	defer rows.Close()

	var out []AttemptRecord
	for rows.Next() {
		rec, err := scanAttempt(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, rec)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("journal: listing unresolved attempts: %w", err)
	}
	return out, nil
}

// OperatorResolve is the only exit from UNRESOLVED_IN_DOUBT.
//
// It requires a terminal target (CONFIRMED or FAILED_CONFIRMED), the operator's
// identity and a note, because this is a human overriding the machine's "I cannot
// tell" about a live account — the audit trail is the point. Confirming also
// requires the broker order id the operator found.
func (j *Journal) OperatorResolve(ctx context.Context, attemptID string, to AttemptState, operator, brokerOrderID, note string) error {
	switch to {
	case StateConfirmed, StateFailedConfirmed:
	default:
		return fmt.Errorf("%w: an operator may only close an attempt as CONFIRMED or FAILED_CONFIRMED, not %s",
			ErrInvalidRequest, to)
	}
	if strings.TrimSpace(operator) == "" {
		return fmt.Errorf("%w: operator resolution requires the operator's identity", ErrInvalidRequest)
	}
	if strings.TrimSpace(note) == "" {
		return fmt.Errorf("%w: operator resolution requires a note explaining what was verified", ErrInvalidRequest)
	}
	if to == StateConfirmed && strings.TrimSpace(brokerOrderID) == "" {
		return fmt.Errorf("%w: confirming by hand requires the broker order id the operator found", ErrInvalidRequest)
	}

	attempt, err := j.Resume(ctx, attemptID)
	if err != nil {
		return err
	}
	// Verbatim again: the operator typed or pasted what the broker shows, and it
	// has to compare byte-for-byte with what the broker returns later.
	return attempt.transition(ctx, to, transitionOpts{
		from:          []AttemptState{StateUnresolvedInDoubt},
		brokerOrderID: brokerOrderID,
		reasonCode:    ReasonOperatorResolved,
		detail:        fmt.Sprintf("operator %s: %s", strings.TrimSpace(operator), strings.TrimSpace(note)),
		setSettled:    true,
	})
}

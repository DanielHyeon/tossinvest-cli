package journal

import (
	"context"
	"fmt"
)

// BlockedSymbol is a symbol that must not accept new mutations until an unresolved
// attempt is settled.
type BlockedSymbol struct {
	Market    string
	Symbol    string
	AttemptID string
	State     AttemptState
}

// RecoveryReport is what a restart found and did.
type RecoveryReport struct {
	// NotDispatched lists attempts this pass safely closed (RECORDED only).
	NotDispatched []string
	// InDoubt lists attempts this pass moved to IN_DOUBT (were DISPATCH_STARTED).
	InDoubt []string
	// Blocked lists every symbol that is still blocked after this pass, including
	// attempts that were already IN_DOUBT or ACKED before it ran.
	Blocked []BlockedSymbol
}

// PendingAttempts returns every attempt that is not in a terminal state, oldest
// first.
func (j *Journal) PendingAttempts(ctx context.Context) ([]AttemptRecord, error) {
	rows, err := j.db.QueryContext(ctx, attemptSelect+
		` WHERE state IN (?,?,?,?) ORDER BY recorded_at, rowid`,
		string(StateRecorded), string(StateDispatchStarted), string(StateAcked), string(StateInDoubt))
	if err != nil {
		return nil, fmt.Errorf("journal: listing pending attempts: %w", err)
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
		return nil, fmt.Errorf("journal: listing pending attempts: %w", err)
	}
	return out, nil
}

// Resume returns a handle on an attempt recorded earlier, so a restart or an
// operator flow can continue its lifecycle.
func (j *Journal) Resume(ctx context.Context, attemptID string) (*Attempt, error) {
	rec, err := j.LookupAttempt(ctx, attemptID)
	if err != nil {
		return nil, err
	}
	return j.handleFor(rec), nil
}

func (j *Journal) handleFor(rec AttemptRecord) *Attempt {
	return &Attempt{
		j:         j,
		id:        rec.ID,
		intentID:  rec.IntentID,
		kind:      rec.Kind,
		attemptNo: rec.AttemptNo,
		state:     rec.State,
	}
}

// RecoverPending applies the restart rules from the order-execution spec:
//
//   - RECORDED only: nothing was ever sent, so the attempt is closed as
//     NOT_DISPATCHED and blocks nothing. This is the whole payoff of writing the
//     intent before dispatching — the ambiguous window is bounded by one fsync.
//   - DISPATCH_STARTED: the request may have reached the broker, so the attempt
//     becomes IN_DOUBT and its symbol is blocked until resolution finishes.
//   - ACKED and IN_DOUBT: left as they are, still blocking. An ACK that was never
//     confirmed and an unresolved doubt both need the resolution pass, not a
//     state change from the recovery pass.
//
// It is idempotent: running it again neither re-transitions nor loses the blocks.
func (j *Journal) RecoverPending(ctx context.Context) (RecoveryReport, error) {
	pending, err := j.PendingAttempts(ctx)
	if err != nil {
		return RecoveryReport{}, err
	}

	report := RecoveryReport{}
	for _, rec := range pending {
		switch rec.State {
		case StateRecorded:
			handle := j.handleFor(rec)
			if err := handle.Settle(ctx, StateNotDispatched, ReasonRestartNotDispatched,
				"found at startup with no dispatch recorded"); err != nil {
				return report, fmt.Errorf("journal: closing attempt %s at startup: %w", rec.ID, err)
			}
			report.NotDispatched = append(report.NotDispatched, rec.ID)

		case StateDispatchStarted:
			handle := j.handleFor(rec)
			if err := handle.MarkInDoubt(ctx, ReasonRestartInDoubt,
				"process stopped after dispatch started; outcome unknown"); err != nil {
				return report, fmt.Errorf("journal: marking attempt %s in doubt at startup: %w", rec.ID, err)
			}
			report.InDoubt = append(report.InDoubt, rec.ID)
			blocked, err := j.blockedFor(ctx, rec, StateInDoubt)
			if err != nil {
				return report, err
			}
			report.Blocked = append(report.Blocked, blocked)

		case StateAcked, StateInDoubt:
			blocked, err := j.blockedFor(ctx, rec, rec.State)
			if err != nil {
				return report, err
			}
			report.Blocked = append(report.Blocked, blocked)
		}
	}
	return report, nil
}

// blockedFor resolves the symbol a pending attempt blocks, from its immutable
// intent.
func (j *Journal) blockedFor(ctx context.Context, rec AttemptRecord, state AttemptState) (BlockedSymbol, error) {
	intent, err := j.LookupIntent(ctx, rec.IntentID)
	if err != nil {
		return BlockedSymbol{}, fmt.Errorf("journal: resolving the symbol of attempt %s: %w", rec.ID, err)
	}
	return BlockedSymbol{
		Market:    intent.Market,
		Symbol:    intent.Symbol,
		AttemptID: rec.ID,
		State:     state,
	}, nil
}

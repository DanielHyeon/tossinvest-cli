package journal

// apply_hook.go is the atomic apply point: the place where a fill, the position
// projection it implies and the exit state it moves all commit together
// (design D7's "원자 apply hook", position-ledger "체결 반영의 원자성").
//
// # The requirement, and why a hook rather than a call afterwards
//
// The snapshot update, the projection update and the exit-state update must be
// one transaction (SHALL). A second commit after RecordFill returns does not
// satisfy that: a crash between the two leaves a fill the projection never saw,
// and the projection is what the reconciliation, the exit policy and the entry
// gate all read. So the transaction has to reach the domain code.
//
// It reaches it by injection and not by ownership. The journal does not own the
// position state machine or the exit policy — those are internal/position and
// internal/exitpolicy, they have their own rules and their own tests, and a
// storage layer that decided when a position becomes CLOSED would be a storage
// layer nobody can audit. What the journal owns is the atomic point. So the
// domain hands the journal two functions, and the fill transaction calls them
// inside its own scope, before it commits.
//
// # The public API
//
//	journal.ApplyHooks{Project, Exit}   the two injected functions
//	(*Journal).SetApplyHooks            bound once, at wiring time
//	journal.ApplyFunc                   func(ctx, *ApplyTx, AppliedFill) error
//	journal.AppliedFill                 what the transaction just recorded
//	journal.ApplyTx                     the tx-scoped write handle
//
// Contract, in order:
//
//  1. Both functions run inside the fill's BEGIN IMMEDIATE, after the snapshot
//     has advanced and after any terminal release, and before the commit.
//     Project runs first, so Exit sees the projection this fill produced.
//  2. Returning an error from either aborts the whole transaction. The fill
//     snapshot does not advance, no fill event is appended, no reservation is
//     released, and RecordFill returns the error. There is no partial state to
//     reconcile, which is the entire point.
//  3. They are NOT called for a refused (fail-closed) snapshot or for a
//     byte-identical re-observation. Nothing was applied, so there is nothing
//     to apply; a hook that fired on a refusal would be applying a fill the
//     ledger just declined to believe.
//  4. A hook MUST NOT call an exported *Journal method. The journal holds a
//     single connection (SetMaxOpenConns(1)) and the caller is holding it, so a
//     hook that opened its own statement would wait for a connection its own
//     transaction owns and deadlock until the context expired. Everything a
//     hook may write is on *ApplyTx, which carries the live transaction.
//  5. The *ApplyTx handle is valid only for the duration of the call. Keeping
//     one and using it later returns an error rather than writing into a
//     transaction that has since committed or rolled back.
//
// # Why the guarded columns live here and nowhere else
//
// `exit_states.taken_ratio_total` moves at the moment a partial take-profit
// fills, and the pending proposal is resolved by the same event (exit-policy:
// 체결 시점 필드 — 체결 반영 트랜잭션의 원자 apply hook에서만). Both are
// therefore written by this file only: there is no exported setter for either,
// and the only way to reach the writers is to be inside a hook, because
// *ApplyTx cannot be obtained anywhere else.
//
// The rule is enforced structurally rather than by convention, and
// TestGuardedExitColumnsAreWrittenOnlyByTheApplyHook fails if any other
// production file in this package so much as names those columns. When the
// proposal path (task 7.3) needs to *arm* a pending proposal — a different
// event, in a different transaction — its writer belongs in this file too, so
// that the one test keeps naming every writer of these columns.
//
// # Broker-behaviour claims
//
// None. AppliedFill carries what RecordFill already derived from the broker's
// cumulative snapshot; this file adds no interpretation of it.

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"sync/atomic"

	"github.com/JungHoonGhae/tossinvest-cli/internal/riskcalc"
)

// ErrExitStateNotFound means no exit_states row carries that position id. A
// position with no exit state is not an error in itself — an externally
// acquired position has none, because it has no entry stop to build a baseline
// from — so the hook decides what to do about it rather than the journal.
var ErrExitStateNotFound = errors.New("journal: no exit state for that position")

// ErrApplyTxClosed means an *ApplyTx was used outside the hook call it was
// handed to. The transaction it referred to is over; writing through it would
// either do nothing or land in somebody else's transaction.
var ErrApplyTxClosed = errors.New("journal: the apply handle is not inside a fill transaction")

// AppliedFill is what the fill transaction recorded, handed to the injected
// functions so they do not have to re-read it.
//
// Delta is the *new* quantity and not the cumulative one: the projection adds
// it, and adding a cumulative quantity is the double count the snapshot table
// exists to prevent. It is "0" for an observation that moved no quantity —
// which still reaches the hooks when it was a correction or a terminal
// transition, because both change what the position is even though neither
// changes how much of it there is.
type AppliedFill struct {
	OrderID    string
	AccountRef string
	Symbol     string
	Market     string
	// State is the caller's derived broker state, and Terminal its verdict that
	// the order can no longer change.
	State    string
	Terminal bool
	// Delta and CumulativeQuantity are decimal strings.
	Delta              string
	CumulativeQuantity string
	AveragePrice       string
	FilledAmount       string
	// Corrected reports an EXECUTION_CORRECTION: the quantity did not move but
	// the average price or the amount did, so the position's cost basis has.
	Corrected bool
	// BrokerVisibleAt is when the broker made this observable; CommittedAt is
	// this transaction's instant, both RFC3339 UTC.
	BrokerVisibleAt string
	CommittedAt     string
}

// ApplyFunc is one injected apply function. It runs inside the fill
// transaction; returning an error rolls the whole thing back.
type ApplyFunc func(ctx context.Context, tx *ApplyTx, fill AppliedFill) error

// ApplyHooks are the domain's two apply functions. Either may be nil, which is
// the state the journal ships in: a journal with no hooks records fills exactly
// as it did before this change.
type ApplyHooks struct {
	// Project updates the position projection from the fill. It runs first.
	Project ApplyFunc
	// Exit updates the exit state — resolving the pending proposal a fill
	// answered, moving taken_ratio_total when a partial take-profit fills. It
	// runs second, so it sees the projection this fill produced.
	Exit ApplyFunc
}

// SetApplyHooks binds the domain's apply functions. It is called once, at
// wiring time, before the fill detector runs.
//
// Rebinding is refused rather than allowed to win: two different projections of
// the same fill stream is not a configuration, it is a bug, and the second
// binding would silently become the truth from an arbitrary fill onwards.
func (j *Journal) SetApplyHooks(hooks ApplyHooks) error {
	if hooks.Project == nil && hooks.Exit == nil {
		return fmt.Errorf("%w: SetApplyHooks was given no apply function; "+
			"leave the hooks unbound instead of binding nothing", ErrInvalidRequest)
	}
	j.applyMu.Lock()
	defer j.applyMu.Unlock()
	if j.applyHooks.Project != nil || j.applyHooks.Exit != nil {
		return fmt.Errorf("%w: the fill apply hooks are already bound; "+
			"a second projection of the same fill stream is not a configuration", ErrInvalidRequest)
	}
	j.applyHooks = hooks
	return nil
}

// runApplyHooks calls the injected functions inside the caller's transaction.
//
// The handle it builds is invalidated when this returns, which is what makes
// rule 5 above true: a hook that stored the pointer holds a closed handle, not
// a way into the next transaction.
func (j *Journal) runApplyHooks(ctx context.Context, tx *sql.Tx, fill AppliedFill) error {
	j.applyMu.RLock()
	hooks := j.applyHooks
	j.applyMu.RUnlock()
	if hooks.Project == nil && hooks.Exit == nil {
		return nil
	}

	handle := &ApplyTx{tx: tx, now: fill.CommittedAt}
	defer handle.invalidate()

	if hooks.Project != nil {
		if err := hooks.Project(ctx, handle, fill); err != nil {
			return fmt.Errorf("journal: projecting the fill of %s: %w", fill.OrderID, err)
		}
	}
	if hooks.Exit != nil {
		if err := hooks.Exit(ctx, handle, fill); err != nil {
			return fmt.Errorf("journal: applying the exit state for the fill of %s: %w",
				fill.OrderID, err)
		}
	}
	return nil
}

// ApplyTx is the write handle a hook gets, valid for the duration of that hook
// call and carrying the fill's own transaction.
//
// It is deliberately narrow. A hook could be handed the raw *sql.Tx, and then
// "taken_ratio_total moves only at the apply point" would be a sentence in a
// document; with these methods it is a property of the type, because the
// statements that write those columns exist nowhere else.
//
// Every field is unexported, so a value constructed outside this package
// (journal.ApplyTx{}) carries no transaction and every method on it refuses.
type ApplyTx struct {
	tx     *sql.Tx
	now    string
	closed atomic.Bool
}

func (a *ApplyTx) invalidate() { a.closed.Store(true) }

// live reports whether the handle still refers to an open transaction.
func (a *ApplyTx) live() error {
	if a == nil || a.tx == nil {
		return fmt.Errorf("%w: this handle was not produced by a fill transaction", ErrApplyTxClosed)
	}
	if a.closed.Load() {
		return fmt.Errorf("%w: the fill transaction it belonged to has already ended", ErrApplyTxClosed)
	}
	return nil
}

// Now is the transaction's instant, RFC3339 UTC, from the journal's injected
// clock. A hook stamps its rows with this rather than with its own clock, so
// everything one fill wrote carries one timestamp.
func (a *ApplyTx) Now() string { return a.now }

// PendingState is the part of exit_states this apply point owns: what has been
// taken, and what proposal is outstanding.
//
// The rest of the row — the baseline, the high-water mark, the ratchet level —
// belongs to the exit policy's own judgement path and is not readable or
// writable from here, because a fill is not an observation and must not move a
// protection level.
type PendingState struct {
	PositionID string
	// TakenRatioTotal is the cumulative taken fraction of the *initial*
	// quantity, a decimal string in [0, 1].
	TakenRatioTotal string
	// The pending triple, all empty when nothing is outstanding. Level carries a
	// ratchet level under RATCHET and a rung index under LADDER.
	PendingAction   string
	PendingLevel    string
	PendingIntentID string
	Completed       bool
}

// Pending reports whether a proposal is outstanding.
func (p PendingState) Pending() bool { return p.PendingAction != "" }

// PendingState reads the fill-time fields of one position's exit state.
func (a *ApplyTx) PendingState(ctx context.Context, positionID string) (PendingState, error) {
	if err := a.live(); err != nil {
		return PendingState{}, err
	}
	id := strings.TrimSpace(positionID)
	var (
		state                          PendingState
		action, level, intentID        sql.NullString
		completed                      int
		takenRatioTotal, scannedPosnID string
	)
	err := a.tx.QueryRowContext(ctx,
		`SELECT position_id, taken_ratio_total, pending_action, pending_level,
		        pending_intent_id, completed
		   FROM exit_states WHERE position_id = ?`, id).
		Scan(&scannedPosnID, &takenRatioTotal, &action, &level, &intentID, &completed)
	if errors.Is(err, sql.ErrNoRows) {
		return PendingState{}, fmt.Errorf("%w: position %s", ErrExitStateNotFound, id)
	}
	if err != nil {
		return PendingState{}, fmt.Errorf("journal: reading the exit state of %s: %w", id, err)
	}
	state.PositionID = scannedPosnID
	state.TakenRatioTotal = takenRatioTotal
	state.PendingAction = action.String
	state.PendingLevel = level.String
	state.PendingIntentID = intentID.String
	state.Completed = completed != 0
	return state, nil
}

// MoveTakenRatioTotal advances the cumulative taken fraction.
//
// The value is the domain's arithmetic — the denominator rule (cumulative
// against the initial quantity, each proposal against the remaining one) is the
// exit policy's, not the journal's. What is enforced here is what the journal
// can know without owning that rule: the fraction is a decimal in [0, 1] and it
// never goes backwards. A decrease would mean the position un-took a
// take-profit, which is not an event that exists; allowing it would let a
// re-processed fill re-arm a proposal that was already answered.
func (a *ApplyTx) MoveTakenRatioTotal(ctx context.Context, positionID, total string) error {
	if err := a.live(); err != nil {
		return err
	}
	id := strings.TrimSpace(positionID)
	next, err := riskcalc.CanonicalDecimal(total)
	if err != nil {
		return fmt.Errorf("%w: taken ratio %q for position %s: %v", ErrInvalidRequest, total, id, err)
	}
	negative, err := riskcalc.IsNegativeDecimal(next)
	if err != nil {
		return fmt.Errorf("%w: taken ratio for position %s: %v", ErrInvalidRequest, id, err)
	}
	if negative {
		return fmt.Errorf("%w: taken ratio %s for position %s is negative", ErrInvalidRequest, next, id)
	}
	if cmp, err := riskcalc.CompareDecimal(next, "1"); err != nil {
		return fmt.Errorf("%w: taken ratio for position %s: %v", ErrInvalidRequest, id, err)
	} else if cmp > 0 {
		return fmt.Errorf("%w: taken ratio %s for position %s exceeds the whole position",
			ErrInvalidRequest, next, id)
	}

	current, err := a.PendingState(ctx, id)
	if err != nil {
		return err
	}
	cmp, err := riskcalc.CompareDecimal(next, current.TakenRatioTotal)
	if err != nil {
		return fmt.Errorf("%w: comparing the taken ratio of position %s: %v", ErrInvalidRequest, id, err)
	}
	if cmp < 0 {
		return fmt.Errorf(
			"%w: taken ratio for position %s would move from %s back to %s; a taken fraction does not un-take",
			ErrInvalidRequest, id, current.TakenRatioTotal, next)
	}

	if _, err := a.tx.ExecContext(ctx,
		`UPDATE exit_states SET taken_ratio_total = ?, updated_at = ? WHERE position_id = ?`,
		next, a.now, id); err != nil {
		return fmt.Errorf("journal: moving the taken ratio of position %s: %w", id, err)
	}
	return nil
}

// ResolvePending clears the outstanding proposal, which is what a fill answering
// it means. Clearing all three columns together is the whole resolution: a
// proposal half-cleared would be re-proposed at the next observation or never
// again, depending on which column the reader looked at.
//
// Resolving when nothing is pending is not an error. A fill can arrive for an
// order whose proposal was already resolved by an earlier partial, and making
// the hook check first would only move the race one statement earlier.
func (a *ApplyTx) ResolvePending(ctx context.Context, positionID string) error {
	if err := a.live(); err != nil {
		return err
	}
	id := strings.TrimSpace(positionID)
	if _, err := a.tx.ExecContext(ctx,
		`UPDATE exit_states
		    SET pending_action = NULL, pending_level = NULL, pending_intent_id = NULL,
		        updated_at = ?
		  WHERE position_id = ?`, a.now, id); err != nil {
		return fmt.Errorf("journal: resolving the pending proposal of position %s: %w", id, err)
	}
	return nil
}

// Exec runs one statement inside the fill transaction, for the projection rows
// this apply point does not own the shape of — positions and
// position_adjustments, whose state machine is internal/position's (task 6.1).
//
// It is not an escape hatch for the guarded exit columns. Those have their own
// methods above, and a hook reaching them through here would be writing SQL
// that TestGuardedExitColumnsAreWrittenOnlyByTheApplyHook cannot see; the
// statement is therefore refused.
func (a *ApplyTx) Exec(ctx context.Context, query string, args ...any) (sql.Result, error) {
	if err := a.live(); err != nil {
		return nil, err
	}
	if column, found := guardedColumn(query); found {
		return nil, fmt.Errorf(
			"%w: this statement writes %s, which moves only through the ApplyTx method for it",
			ErrInvalidRequest, column)
	}
	return a.tx.ExecContext(ctx, query, args...)
}

// Query reads inside the fill transaction, so a projection sees the rows this
// transaction has written and nobody else's uncommitted work.
//
// There is no QueryRow counterpart on purpose: *sql.Row carries its error until
// Scan, so a closed handle would surface as a scan failure that a caller could
// mistake for "no such row". Every read here returns its refusal directly.
func (a *ApplyTx) Query(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	if err := a.live(); err != nil {
		return nil, err
	}
	return a.tx.QueryContext(ctx, query, args...)
}

// guardedColumns are the exit_states fields this file is the sole writer of.
// The list is duplicated in the static test on purpose: one copy is the rule
// the code enforces at runtime, the other is the rule the package layout is
// checked against, and they are meant to be compared by a human when either
// changes.
var guardedColumns = []string{
	"taken_ratio_total", "pending_action", "pending_level", "pending_intent_id",
}

func guardedColumn(query string) (string, bool) {
	lowered := strings.ToLower(query)
	for _, column := range guardedColumns {
		if strings.Contains(lowered, column) {
			return column, true
		}
	}
	return "", false
}

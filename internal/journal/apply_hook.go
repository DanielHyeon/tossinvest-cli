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
// Task 7.3 took that instruction literally. The second half of this file is the
// proposal lifecycle: arming, attaching the intent, resolving a refusal or a
// cancellation, reading the state back, and the Exit apply function itself. All
// of it is here and none of it is in exit_state.go, which owns the rest of the
// row and is forbidden from spelling these four column names at all. The seam
// between the two files is deliberately awkward, because the awkwardness is what
// makes a second writer impossible to add without reading this paragraph.
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

	"github.com/JungHoonGhae/tossinvest-cli/internal/costs"
	"github.com/JungHoonGhae/tossinvest-cli/internal/exitpolicy"
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
	// OrderedQuantity is the quantity the order was placed for — 원주문 수량,
	// which is what the projection judges "this entry is complete" against
	// (internal/position). It is the broker's own `quantity` from this snapshot.
	OrderedQuantity string
	// PrevCumulativeQuantity and PrevAveragePrice are the snapshot as it stood
	// *before* this observation. They are carried rather than re-read because by
	// the time a hook runs the row has already been advanced and the previous
	// values no longer exist anywhere.
	//
	// They are what makes a cost basis exact. An order's contribution to a
	// position is `cumulative filled × average price`, so what one observation
	// contributes is the change in that product — and without the previous pair
	// the only available approximation is "the delta cost the order's *current*
	// average", which is wrong for every order that filled in more than one
	// piece at more than one price. Both are "" on an order's first observation.
	PrevCumulativeQuantity string
	PrevAveragePrice       string
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
	// Costs is the shared cost model the frozen trade outcome is priced with
	// (task 8.1). It is here rather than on the journal because it is the same
	// injection the other two are: the journal owns the atomic point and refuses
	// to own the domain's numbers.
	//
	// An unconfigured model does not fail anything. It means no outcome row is
	// written at the close, which is recoverable through BackfillTradeOutcome —
	// whereas pricing a round trip as free would report every trade as better
	// than it was, permanently, in a row nothing rewrites.
	Costs costs.Model
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

// ProjectionBound reports whether a Project hook is bound.
//
// It exists for a wiring guard and not for control flow, and the thing it
// guards is a fail-*open*: reconciliation's local state is the position
// projection (internal/reconcile LocalStateFromJournal), and the projection is
// produced by this hook. A process that records fills with nothing bound
// believes it holds nothing, every broker holding then classifies as an external
// position rather than a disagreement, and the entry gate is not raised. So a
// profile that is going to reconcile has to be able to ask whether the thing
// that produces its beliefs is connected.
//
// The question is asked of the binding and never inferred from the data. "There
// are fills but no positions" is a legitimate state for a journal migrated from
// v5, whose fills predate the projection entirely, and a heuristic that refused
// that would refuse every upgraded install.
func (j *Journal) ProjectionBound() bool {
	j.applyMu.RLock()
	defer j.applyMu.RUnlock()
	return j.applyHooks.Project != nil
}

// ExitApplierBound reports whether an Exit hook is bound.
//
// The same question about the other half, asked for a smaller but real reason
// (task 7.3): the exit applier is what resolves a pending proposal, and a
// process that records fills without it leaves every proposal outstanding, which
// suppresses every proposal after the first. That is a fail-*closed* silence
// rather than a fail-open one, so it does not carry the projection's argument for
// refusing to start — but it is still a wiring bug that nothing else would
// notice, because a system that stops proposing looks exactly like a system with
// nothing to propose.
func (j *Journal) ExitApplierBound() bool {
	j.applyMu.RLock()
	defer j.applyMu.RUnlock()
	return j.applyHooks.Exit != nil
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

	handle := &ApplyTx{tx: tx, now: fill.CommittedAt, costs: hooks.Costs}
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
	costs  costs.Model
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

// --- the proposal lifecycle (task 7.3) ----------------------------------------
//
// Everything below writes or reads one of the guarded four, which is why it is
// in this file. The lifecycle it implements is exit-policy's:
//
//	arm      — one proposal per level or rung, recorded before it is submitted
//	attach   — the intent id, once the issuance path has minted one
//	resolve  — a fill (in the apply hook), or a refusal or cancellation (not)
//	restore  — after a crash, the armed proposal is still armed
//
// The three of those that are not fills happen in their own transactions, so
// they are *Journal methods rather than *ApplyTx ones. That is the shape
// issues.md predicted for this task and it is the right one: arming is a
// judgement's write, not a fill's.

// ErrProposalPending means a proposal is already outstanding for the position.
// Arming a second one is the duplicate the pending lifecycle exists to prevent.
var ErrProposalPending = errors.New("journal: a proposal is already outstanding for that position")

// ExitState is one position's whole protection state.
type ExitState struct {
	PositionID string
	PolicyKind string
	PolicyID   string
	// PolicyIdentity is populated by new state creation and by the engine's
	// runtime compatibility resolver. Rows read from the pre-a042 schema leave
	// it zero rather than inventing a version or digest from today's registry.
	PolicyIdentity exitpolicy.PolicyIdentity
	// EntryPrice, InitialStop and InitialRisk are frozen at t0. InitialRisk is
	// the denominator of every R the position is judged by and it does not move
	// for a partial take-profit or an adjustment.
	EntryPrice  string
	InitialStop string
	InitialRisk string
	// Baseline and HighWater are decimal strings, both monotone non-decreasing.
	Baseline  string
	HighWater string
	// RatchetLevel is the ratchet's step; ActiveRung is the ladder's, and is
	// exitpolicy.NoRung when none has been reached.
	RatchetLevel string
	ActiveRung   int
	// TakenRatioTotal is the cumulative taken fraction of the initial quantity.
	TakenRatioTotal string
	// The outstanding proposal, all empty when there is none. Level carries a
	// ratchet level under RATCHET and a rung index under LADDER.
	PendingAction   string
	PendingLevel    string
	PendingIntentID string
	Completed       bool
	UpdatedAt       string
}

// Pending reports whether a proposal is outstanding.
func (s ExitState) Pending() bool { return s.PendingAction != "" }

const exitStateSelect = `SELECT position_id, policy_kind, coalesce(policy_id,''), entry_price, initial_stop, initial_risk,
	baseline_price, high_water, ratchet_level, active_rung, taken_ratio_total,
	coalesce(pending_action,''), coalesce(pending_level,''), coalesce(pending_intent_id,''),
	completed, updated_at FROM exit_states`

func scanExitState(row rowScanner) (ExitState, error) {
	var (
		s    ExitState
		rung sql.NullInt64
		done int
	)
	err := row.Scan(&s.PositionID, &s.PolicyKind, &s.PolicyID, &s.EntryPrice, &s.InitialStop, &s.InitialRisk,
		&s.Baseline, &s.HighWater, &s.RatchetLevel, &rung, &s.TakenRatioTotal,
		&s.PendingAction, &s.PendingLevel, &s.PendingIntentID, &done, &s.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return ExitState{}, ErrExitStateNotFound
	}
	if err != nil {
		return ExitState{}, fmt.Errorf("journal: reading an exit state: %w", err)
	}
	s.ActiveRung = -1
	if s.PolicyKind == ExitPolicyLadder && s.PolicyID == "" {
		s.PolicyID = "default_v1"
	}
	if rung.Valid {
		s.ActiveRung = int(rung.Int64)
	}
	s.Completed = done != 0
	return s, nil
}

// ExitState returns one position's protection state.
func (j *Journal) ExitState(ctx context.Context, positionID string) (ExitState, error) {
	return scanExitState(j.db.QueryRowContext(ctx,
		exitStateSelect+" WHERE position_id = ?", strings.TrimSpace(positionID)))
}

// OpenExitStates returns the positions whose exit policy is still running, on
// one account.
//
// It is two things at once, and deliberately not two functions. It is the
// observation loop's working set (task 7.4), and it is the crash restore: a
// process that starts with an armed proposal in this list re-reads it before it
// evaluates, which is what makes "발의 기록 후 제출 전에 크래시" resume without
// proposing the same level twice. A separate "restore" query would be a second
// definition of what is outstanding, and the two could disagree.
func (j *Journal) OpenExitStates(ctx context.Context, accountRef string) ([]ExitState, error) {
	account := strings.TrimSpace(accountRef)
	rows, err := j.db.QueryContext(ctx, `
		SELECT e.position_id, e.policy_kind, coalesce(e.policy_id,''), e.entry_price, e.initial_stop, e.initial_risk,
		       e.baseline_price, e.high_water, e.ratchet_level, e.active_rung, e.taken_ratio_total,
		       coalesce(e.pending_action,''), coalesce(e.pending_level,''),
		       coalesce(e.pending_intent_id,''), e.completed, e.updated_at
		  FROM exit_states e
		  JOIN positions p ON p.id = e.position_id
		 WHERE p.account_ref = ? AND e.completed = 0
		 ORDER BY p.market, p.symbol, p.instance_seq`, account)
	if err != nil {
		return nil, fmt.Errorf("journal: listing the exit states of %s: %w", account, err)
	}
	defer rows.Close()

	var out []ExitState
	for rows.Next() {
		state, err := scanExitState(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, state)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("journal: listing the exit states of %s: %w", account, err)
	}
	return out, nil
}

// armExitProposalTx records a proposal inside the judgement's own transaction.
//
// It is the arming half of the crash contract: the row is written before the
// order is submitted, so a process that dies in between restarts holding the
// proposal rather than a clean state that would make it again.
//
// A second proposal while one is outstanding is refused rather than overwritten.
// Overwriting would lose the identity of the order that is still working, and
// the fill that eventually arrives would then resolve nothing.
func armExitProposalTx(ctx context.Context, tx *sql.Tx, positionID string,
	proposal ExitProposal, now string) error {
	var action sql.NullString
	err := tx.QueryRowContext(ctx,
		`SELECT pending_action FROM exit_states WHERE position_id = ?`, positionID).Scan(&action)
	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("%w: position %s", ErrExitStateNotFound, positionID)
	}
	if err != nil {
		return fmt.Errorf("journal: reading the proposal of %s: %w", positionID, err)
	}
	if strings.TrimSpace(action.String) != "" {
		return fmt.Errorf("%w: %s holds %s", ErrProposalPending, positionID, action.String)
	}
	if _, err := tx.ExecContext(ctx, `
		UPDATE exit_states
		   SET pending_action = ?, pending_level = ?, pending_intent_id = ?, updated_at = ?
		 WHERE position_id = ?`,
		proposal.Action, proposal.Level, nullableString(proposal.IntentID), now, positionID); err != nil {
		return fmt.Errorf("journal: arming the proposal of %s: %w", positionID, err)
	}
	return nil
}

// AttachExitIntent records which intent carries an already-armed proposal.
//
// The two steps are separate because the order they happen in is the crash
// contract: arm, mint the intent, attach, submit. An armed proposal with no
// intent id is a recoverable state — the restart re-proposes nothing and the
// operator sees an outstanding proposal with no order — whereas an intent
// submitted before anything was armed is the duplicate this whole lifecycle is
// about.
func (j *Journal) AttachExitIntent(ctx context.Context, positionID, intentID string) error {
	id := strings.TrimSpace(positionID)
	intent := strings.TrimSpace(intentID)
	if intent == "" {
		return fmt.Errorf("%w: an intent id is required to attach one", ErrInvalidRequest)
	}
	now := j.nowString()

	tx, err := j.db.BeginTx(ctx, nil) // BEGIN IMMEDIATE
	if err != nil {
		return fmt.Errorf("journal: attaching the intent of %s: %w", id, err)
	}
	defer tx.Rollback()

	var action, existing sql.NullString
	err = tx.QueryRowContext(ctx,
		`SELECT pending_action, pending_intent_id FROM exit_states WHERE position_id = ?`, id).
		Scan(&action, &existing)
	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("%w: position %s", ErrExitStateNotFound, id)
	}
	if err != nil {
		return fmt.Errorf("journal: reading the proposal of %s: %w", id, err)
	}
	if strings.TrimSpace(action.String) == "" {
		return fmt.Errorf("%w: %s has no outstanding proposal to attach an intent to",
			ErrInvalidRequest, id)
	}
	if current := strings.TrimSpace(existing.String); current != "" && current != intent {
		return fmt.Errorf(
			"%w: the proposal of %s already carries intent %s; a second intent for one proposal is two orders",
			ErrInvalidRequest, id, current)
	}

	if _, err := tx.ExecContext(ctx,
		`UPDATE exit_states SET pending_intent_id = ?, updated_at = ? WHERE position_id = ?`,
		intent, now, id); err != nil {
		return fmt.Errorf("journal: attaching the intent of %s: %w", id, err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("journal: attaching the intent of %s: %w", id, err)
	}
	return nil
}

// ProposalResolution is how a proposal ended without filling.
type ProposalResolution string

const (
	// ProposalRefused is a proposal Guardian or the gateway declined to submit.
	ProposalRefused ProposalResolution = "REFUSED"
	// ProposalCancelled is a submitted order that was cancelled or expired.
	ProposalCancelled ProposalResolution = "CANCELLED"
)

// ResolveExitProposal clears a proposal that ended without filling, which re-arms
// its level (exit-policy: 거부·취소 시 해당 레벨은 재발의 가능해진다).
//
// Under LADDER it also rolls the rung back, because the rung index is what that
// policy de-duplicates on and a rung left promoted could never be proposed again.
// The baseline is not rolled back with it: a lock once granted is protection, and
// §0.9 permits no movement in the other direction.
//
// Clearing when nothing is pending is not an error, for the same reason
// ApplyTx.ResolvePending's is not: a retry of the resolution must converge rather
// than fail, and checking first would only move the race one statement earlier.
//
// The reason for the refusal is not recorded here. `exit_events` has no free-text
// column (D7), and the subsystem that refused has its own record; the join
// between them is the intent id this event carries.
func (j *Journal) ResolveExitProposal(ctx context.Context, positionID string,
	resolution ProposalResolution) error {
	id := strings.TrimSpace(positionID)
	action := ExitEventProposalRefused
	switch resolution {
	case ProposalRefused:
	case ProposalCancelled:
		action = ExitEventProposalCancelled
	default:
		return fmt.Errorf("%w: %q is neither %s nor %s", ErrInvalidRequest,
			string(resolution), ProposalRefused, ProposalCancelled)
	}
	now := j.nowString()

	tx, err := j.db.BeginTx(ctx, nil) // BEGIN IMMEDIATE
	if err != nil {
		return fmt.Errorf("journal: resolving the proposal of %s: %w", id, err)
	}
	defer tx.Rollback()

	var kind string
	var pendingAction, pendingLevel, pendingIntent sql.NullString
	err = tx.QueryRowContext(ctx, `
		SELECT policy_kind, pending_action, pending_level, pending_intent_id
		  FROM exit_states WHERE position_id = ?`, id).
		Scan(&kind, &pendingAction, &pendingLevel, &pendingIntent)
	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("%w: position %s", ErrExitStateNotFound, id)
	}
	if err != nil {
		return fmt.Errorf("journal: reading the proposal of %s: %w", id, err)
	}
	if strings.TrimSpace(pendingAction.String) == "" {
		return nil
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE exit_states
		   SET pending_action = NULL, pending_level = NULL, pending_intent_id = NULL, updated_at = ?
		 WHERE position_id = ?`, now, id); err != nil {
		return fmt.Errorf("journal: resolving the proposal of %s: %w", id, err)
	}
	if kind == ExitPolicyLadder {
		if rung, err := exitpolicy.RungIndex(strings.TrimSpace(pendingLevel.String)); err == nil {
			if err := rollBackRungTx(ctx, tx, id, rung-1, now); err != nil {
				return err
			}
		}
	}
	if err := appendExitEventTx(ctx, tx, exitEventRow{
		PositionID: id, LevelAfter: strings.TrimSpace(pendingLevel.String),
		Action: action, ProposedIntentID: strings.TrimSpace(pendingIntent.String), CreatedAt: now,
	}); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("journal: resolving the proposal of %s: %w", id, err)
	}
	return nil
}

// ApplyExitFill is the Exit apply hook: it moves the exit state that a fill has
// answered, inside the fill's own transaction.
//
// Wiring is the second half of one line at startup:
//
//	j.SetApplyHooks(journal.ApplyHooks{Project: …, Exit: journal.ApplyExitFill})
//
// It runs after Project, which is what lets it read the quantity this fill
// produced rather than re-deriving it. Three things happen, in order:
//
//  1. **The taken fraction moves**, for a SELL. It is reconstructed from the
//     quantities rather than from a stored initial quantity
//     (exitpolicy.TakenAfterFill), because `exit_states` has no such column and
//     adding one would have been a second authority on the same number. Every
//     SELL counts, not only a proposal's own: a manual or flatten sale really did
//     take part of the position, and a cumulative fraction that ignored it would
//     under-report what is gone and re-propose a take-profit against quantity
//     that is not there.
//
//     A BUY moves nothing. A scale-in changes the initial quantity the fraction
//     is against, and this change has no rule for that; the fraction is left
//     alone rather than silently reinterpreted (see issues.md).
//
//  2. **The proposal resolves**, when this fill's intent is the one the proposal
//     was armed with and the order is terminal. Terminal, because a partially
//     filled order that is still working has not answered the proposal — clearing
//     it there would re-arm the level while the order is live and let a second one
//     be proposed on top of it.
//
//  3. **The state completes**, when the position reached CLOSED. A completed
//     state leaves the observation loop's working set, which is what stops a
//     closed position being judged forever.
//
// Nothing here decides anything about protection. A fill is not an observation
// and must not move a baseline, a watermark or a level — those reach the row only
// through RecordExitJudgement, and *ApplyTx cannot write them at all.
func ApplyExitFill(ctx context.Context, tx *ApplyTx, fill AppliedFill) error {
	where, found, err := exitContextForFill(ctx, tx, fill)
	if err != nil {
		return err
	}
	if !found {
		// An external order, or a symbol the projection holds no instance of.
		// Neither is an error and neither has an exit state to move.
		return nil
	}

	state, err := tx.PendingState(ctx, where.PositionID)
	if errors.Is(err, ErrExitStateNotFound) {
		// A position the exit policy does not manage — an external position, or
		// one whose state has not been opened yet.
		return nil
	}
	if err != nil {
		return err
	}

	if strings.EqualFold(strings.TrimSpace(where.Side), "SELL") {
		next, err := exitpolicy.TakenAfterFill(
			state.TakenRatioTotal, orZero(fill.Delta), orZero(where.QuantityAfter))
		if err != nil {
			return fmt.Errorf("%w: the taken fraction of position %s: %v",
				ErrInvalidRequest, where.PositionID, err)
		}
		if cmp, err := riskcalc.CompareDecimal(next, state.TakenRatioTotal); err != nil {
			return fmt.Errorf("%w: the taken fraction of position %s: %v",
				ErrInvalidRequest, where.PositionID, err)
		} else if cmp > 0 {
			if err := tx.MoveTakenRatioTotal(ctx, where.PositionID, next); err != nil {
				return err
			}
		}
	}

	if fill.Terminal && state.Pending() && state.PendingIntentID != "" &&
		state.PendingIntentID == where.IntentID {
		if err := tx.ResolvePending(ctx, where.PositionID); err != nil {
			return err
		}
		if err := appendExitEventFromHook(ctx, tx, exitEventRow{
			PositionID: where.PositionID, LevelAfter: state.PendingLevel,
			Action: ExitEventProposalFilled, ProposedIntentID: state.PendingIntentID,
			CreatedAt: tx.Now(),
		}); err != nil {
			return err
		}
	}

	if where.PositionState == PositionClosed && !state.Completed {
		if err := markExitCompletedTx(ctx, tx, where.PositionID); err != nil {
			return err
		}
		// 4. **The outcome freezes** (task 8.1). Last, and deliberately unable to
		//    fail: trade-analytics isolates the analytics path from the order path
		//    (SHALL NOT — 분석 작업의 실패·지연이 체결 반영·청산 처리를 …), so a
		//    trade this cannot price leaves no row rather than rolling back a fill
		//    the broker has already reported. BackfillTradeOutcome recovers it from
		//    data that, the position being CLOSED, cannot move any more.
		freezeTradeOutcomeTx(ctx, tx, where.PositionID, tx.costs)
	}
	return nil
}

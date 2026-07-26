package journal

// reconcile_states.go is the durable RECONCILE state (design D6, task 4.1;
// order-execution "RECONCILE 상태").
//
// # The journal is the authority
//
// Before this file, RECONCILE lived in two in-memory places: the entry gate's
// latches and reconcile.Tracker's block set. Both are lost on restart, which
// means a process that stopped while the account and the engine disagreed came
// back believing they agreed. The spec closes that: RECONCILE 상태는 journal에
// 영속된다(SHALL) … 기존 in-memory 차단 기계는 기동 시 journal에서 재구성되는
// 투영이다(SHALL — 재시작이 차단을 잃어서는 안 된다).
//
// So the rows here are the truth and the latches are a view of them. Entering
// and releasing go through this file; the projections are rebuilt from
// ActiveReconcileStates.
//
// # Why the cause is validated here rather than by a CHECK
//
// D9 wrote "cause enum NOT NULL" without enumerating the values, and a CHECK
// constraint cannot be relaxed in SQLite without rebuilding the table — which
// schema.go's additive rules forbid. Task 0.1 therefore left the column
// unconstrained and recorded the decision in issues.md; the Manager approved it
// **on the condition that the storage layer validates the cause against the Go
// constant set at write time and refuses an unknown one**. That is what
// ValidReconcileCause and the refusals below are. A cause nobody enumerated is
// a state nobody can write a release procedure for, so it fails closed.
//
// # Scope
//
// A row is either account-wide (symbol NULL) or one symbol. One active row per
// scope, enforced by the two partial UNIQUE indexes from task 0.1 — re-entering
// an already-active scope is a no-op that keeps the first observation, because
// a state whose "since" advances on every poll looks new forever.
//
// # Broker-behaviour claims
//
// None. The causes are TossOS-internal judgements about what was observed; the
// two that come from broker shape (IDENTIFIER_CONFLICT, ATTRIBUTION_FAILED) are
// documented at their constants in execution_contract.go, and their concrete
// broker form is [미측정 — 2b 2.1].

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"
)

// Release causes. A release names why the state is over, and the spec requires
// that naming: 해제는 재조회 일치와 원인 기록을 요구한다(SHALL).
const (
	// ReconcileReleaseRecheckMatched — a fresh comparison agreed. This is the
	// only automatic exit, and it exists because the condition it closes was a
	// disagreement that a later read can disprove.
	ReconcileReleaseRecheckMatched = "RECHECK_MATCHED"
	// ReconcileReleaseOperator — a human closed it, with evidence.
	ReconcileReleaseOperator = "OPERATOR"
)

// ErrReconcileStateNotFound means no matching reconcile_states row exists.
var ErrReconcileStateNotFound = errors.New("journal: reconcile state not found")

// ValidReconcileCause reports whether a cause is one this build writes.
//
// It is the write-time constraint the column does not carry. Adding a value
// here is a deliberate act; discovering one at runtime is not.
func ValidReconcileCause(cause string) bool {
	switch cause {
	case ReconcileCauseSnapshotUnavailable, ReconcileCauseSnapshotStale,
		ReconcileCauseQuantityMismatch, ReconcileCauseIdentifierConflict,
		ReconcileCauseAttributionFailed:
		return true
	default:
		return false
	}
}

// ValidReconcileReleaseCause reports whether a release cause is one this build
// writes.
func ValidReconcileReleaseCause(cause string) bool {
	switch cause {
	case ReconcileReleaseRecheckMatched, ReconcileReleaseOperator:
		return true
	default:
		return false
	}
}

// ReconcileState is a stored reconcile_states row.
type ReconcileState struct {
	ID         string
	AccountRef string
	// Symbol is empty for an account-wide state.
	Symbol string
	Cause  string
	// Evidence is what was observed, for the operator who has to act on it.
	Evidence     string
	EnteredAt    time.Time
	ReleasedAt   time.Time
	ReleaseCause string
}

// Active reports whether the state still blocks.
func (s ReconcileState) Active() bool { return s.ReleaseCause == "" && s.ReleasedAt.IsZero() }

// AccountWide reports whether the state covers the whole account.
func (s ReconcileState) AccountWide() bool { return strings.TrimSpace(s.Symbol) == "" }

// EnterReconcileRequest asks the journal to enter a RECONCILE state.
type EnterReconcileRequest struct {
	// ID is optional. An empty one is derived from the scope, cause and
	// instant, so a producer that has no id scheme of its own does not have to
	// invent one.
	ID         string
	AccountRef string
	// Symbol empty means account-wide.
	Symbol string
	// Cause must be one of the ReconcileCause* constants.
	Cause string
	// Evidence is required: a state an operator cannot act on is a state that
	// stays forever.
	Evidence string
}

// EnterReconcile records a RECONCILE state, or reports the one already active
// for that scope.
//
// The bool is true when this call entered the state and false when it was
// already active. Re-entering keeps the original row — its cause, its evidence
// and its entered_at — because the first observation is the one an operator
// should be looking at, and the partial UNIQUE indexes would refuse a second
// active row anyway.
func (j *Journal) EnterReconcile(ctx context.Context, req EnterReconcileRequest) (ReconcileState, bool, error) {
	account := strings.TrimSpace(req.AccountRef)
	symbol := strings.ToUpper(strings.TrimSpace(req.Symbol))
	cause := strings.TrimSpace(req.Cause)
	evidence := strings.TrimSpace(req.Evidence)
	switch {
	case account == "":
		return ReconcileState{}, false, fmt.Errorf(
			"%w: a RECONCILE state is scoped to an account; none was named", ErrInvalidRequest)
	case !ValidReconcileCause(cause):
		return ReconcileState{}, false, fmt.Errorf(
			"%w: RECONCILE cause %q is not one this build writes; an unenumerated cause has no release procedure",
			ErrInvalidRequest, req.Cause)
	case evidence == "":
		return ReconcileState{}, false, fmt.Errorf(
			"%w: a RECONCILE state requires the evidence that raised it", ErrInvalidRequest)
	}

	now := j.clk.Now().UTC()
	nowText := formatJournalTime(now)

	tx, err := j.db.BeginTx(ctx, nil)
	if err != nil {
		return ReconcileState{}, false, fmt.Errorf("journal: starting the RECONCILE transaction: %w", err)
	}
	defer tx.Rollback()

	existing, err := scanReconcileState(tx.QueryRowContext(ctx,
		reconcileSelect+activeScopeWhere(symbol), scopeArgs(account, symbol)...))
	switch {
	case err == nil:
		return existing, false, nil
	case errors.Is(err, ErrReconcileStateNotFound):
	default:
		return ReconcileState{}, false, err
	}

	id := strings.TrimSpace(req.ID)
	if id == "" {
		id = reconcileStateID(account, symbol, cause, nowText)
	}
	if _, err := tx.ExecContext(ctx,
		`INSERT INTO reconcile_states
		   (id, account_ref, symbol, cause, evidence, entered_at, released_at, release_cause)
		 VALUES (?,?,?,?,?,?,NULL,NULL)`,
		id, account, nullableString(symbol), cause, evidence, nowText); err != nil {
		return ReconcileState{}, false, fmt.Errorf(
			"journal: entering RECONCILE for %s/%s: %w", account, symbolLabel(symbol), err)
	}
	if err := tx.Commit(); err != nil {
		return ReconcileState{}, false, fmt.Errorf(
			"journal: committing the RECONCILE state for %s/%s: %w", account, symbolLabel(symbol), err)
	}
	return ReconcileState{
		ID: id, AccountRef: account, Symbol: symbol, Cause: cause,
		Evidence: evidence, EnteredAt: now,
	}, true, nil
}

// ReleaseReconcileRequest closes an active RECONCILE state.
type ReleaseReconcileRequest struct {
	AccountRef string
	// Symbol empty means the account-wide state.
	Symbol string
	// Cause is why it is over: RECHECK_MATCHED or OPERATOR.
	Cause string
	// Evidence is what established that. Required — "해제는 재조회 일치와 원인
	// 기록을 요구한다".
	Evidence string
	// ExpectCause releases only a state entered for that cause. It is how a
	// producer releases *its own* states and not somebody else's: a clean
	// quantity comparison says nothing about an identifier conflict, and
	// clearing one with the other is how a block silently disappears.
	ExpectCause string
}

// ReleaseReconcile closes the active state for a scope.
//
// The bool is false when there was nothing active to close, or when
// ExpectCause did not match — neither is an error, because a producer that has
// nothing to release is the normal case on every clean pass.
func (j *Journal) ReleaseReconcile(ctx context.Context, req ReleaseReconcileRequest) (ReconcileState, bool, error) {
	account := strings.TrimSpace(req.AccountRef)
	symbol := strings.ToUpper(strings.TrimSpace(req.Symbol))
	cause := strings.TrimSpace(req.Cause)
	evidence := strings.TrimSpace(req.Evidence)
	switch {
	case account == "":
		return ReconcileState{}, false, fmt.Errorf(
			"%w: a RECONCILE release is scoped to an account; none was named", ErrInvalidRequest)
	case !ValidReconcileReleaseCause(cause):
		return ReconcileState{}, false, fmt.Errorf(
			"%w: RECONCILE release cause %q is not one this build writes", ErrInvalidRequest, req.Cause)
	case evidence == "":
		return ReconcileState{}, false, fmt.Errorf(
			"%w: releasing a RECONCILE state requires the evidence that it is over", ErrInvalidRequest)
	}

	now := j.clk.Now().UTC()
	nowText := formatJournalTime(now)

	tx, err := j.db.BeginTx(ctx, nil)
	if err != nil {
		return ReconcileState{}, false, fmt.Errorf("journal: starting the RECONCILE release: %w", err)
	}
	defer tx.Rollback()

	active, err := scanReconcileState(tx.QueryRowContext(ctx,
		reconcileSelect+activeScopeWhere(symbol), scopeArgs(account, symbol)...))
	if errors.Is(err, ErrReconcileStateNotFound) {
		return ReconcileState{}, false, nil
	}
	if err != nil {
		return ReconcileState{}, false, err
	}
	if expect := strings.TrimSpace(req.ExpectCause); expect != "" && active.Cause != expect {
		return active, false, nil
	}

	if _, err := tx.ExecContext(ctx,
		`UPDATE reconcile_states SET released_at = ?, release_cause = ?, evidence = ?
		 WHERE id = ? AND released_at IS NULL`,
		nowText, cause, active.Evidence+" | released: "+evidence, active.ID); err != nil {
		return ReconcileState{}, false, fmt.Errorf("journal: releasing RECONCILE state %s: %w", active.ID, err)
	}
	if err := tx.Commit(); err != nil {
		return ReconcileState{}, false, fmt.Errorf("journal: committing the release of %s: %w", active.ID, err)
	}

	active.ReleasedAt = now
	active.ReleaseCause = cause
	return active, true, nil
}

// ActiveReconcileStates returns every state still blocking, oldest first. It is
// what the in-memory projections are rebuilt from at startup.
func (j *Journal) ActiveReconcileStates(ctx context.Context) ([]ReconcileState, error) {
	return j.queryReconcileStates(ctx,
		reconcileSelect+" WHERE released_at IS NULL ORDER BY entered_at, rowid")
}

// ReconcileStateHistory returns every state an account has ever been in,
// oldest first.
func (j *Journal) ReconcileStateHistory(ctx context.Context, accountRef string) ([]ReconcileState, error) {
	return j.queryReconcileStates(ctx,
		reconcileSelect+" WHERE account_ref = ? ORDER BY entered_at, rowid", strings.TrimSpace(accountRef))
}

func (j *Journal) queryReconcileStates(ctx context.Context, query string, args ...any) ([]ReconcileState, error) {
	rows, err := j.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("journal: listing RECONCILE states: %w", err)
	}
	defer rows.Close()

	var out []ReconcileState
	for rows.Next() {
		state, err := scanReconcileState(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, state)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("journal: listing RECONCILE states: %w", err)
	}
	return out, nil
}

const reconcileSelect = `SELECT id, account_ref, symbol, cause, evidence, entered_at,
	released_at, release_cause FROM reconcile_states`

// activeScopeWhere selects the active row for one scope. The account-wide scope
// needs `symbol IS NULL` rather than `symbol = ”`: SQLite's NULL is what the
// partial UNIQUE indexes are written against.
func activeScopeWhere(symbol string) string {
	if symbol == "" {
		return " WHERE account_ref = ? AND symbol IS NULL AND released_at IS NULL"
	}
	return " WHERE account_ref = ? AND symbol = ? AND released_at IS NULL"
}

func scopeArgs(account, symbol string) []any {
	if symbol == "" {
		return []any{account}
	}
	return []any{account, symbol}
}

func symbolLabel(symbol string) string {
	if symbol == "" {
		return "(account-wide)"
	}
	return symbol
}

func scanReconcileState(row rowScanner) (ReconcileState, error) {
	var (
		state                    ReconcileState
		symbol, evidence         sql.NullString
		releasedAt, releaseCause sql.NullString
		enteredAt                string
	)
	err := row.Scan(&state.ID, &state.AccountRef, &symbol, &state.Cause, &evidence,
		&enteredAt, &releasedAt, &releaseCause)
	if errors.Is(err, sql.ErrNoRows) {
		return ReconcileState{}, ErrReconcileStateNotFound
	}
	if err != nil {
		return ReconcileState{}, fmt.Errorf("journal: reading a RECONCILE state: %w", err)
	}
	state.Symbol = symbol.String
	state.Evidence = evidence.String
	state.ReleaseCause = releaseCause.String
	if state.EnteredAt, err = parseJournalTime(enteredAt); err != nil {
		return ReconcileState{}, fmt.Errorf("journal: RECONCILE state %s entered_at: %w", state.ID, err)
	}
	if releasedAt.Valid && releasedAt.String != "" {
		if state.ReleasedAt, err = parseJournalTime(releasedAt.String); err != nil {
			return ReconcileState{}, fmt.Errorf("journal: RECONCILE state %s released_at: %w", state.ID, err)
		}
	}
	return state, nil
}

// reconcileStateID derives an id for a producer that has none, the same way
// correctionID does: a length-prefixed hash, so ("ab","c") and ("a","bc") are
// different keys.
func reconcileStateID(account, symbol, cause, enteredAt string) string {
	h := sha256.New()
	for _, part := range []string{account, symbol, cause, enteredAt} {
		fmt.Fprintf(h, "%d:%s|", len(part), part)
	}
	return "rec-" + hex.EncodeToString(h.Sum(nil))[:24]
}

package journal

// flatten.go is the durable record of a flatten saga (harden-execution-base task
// 4.4, engine-safety "Flatten Saga").
//
// # Why the saga needs its own record
//
// The spec requires flatten to resume safely after a crash. Everything the saga
// does is already journalled as intents and attempts — but that record answers
// "what mutations happened", not "how far through exiting the account are we".
// Those differ in exactly the case that matters: a process that died after
// cancelling four of seven orders leaves four settled attempts and no statement
// anywhere that three more were supposed to follow.
//
// So the saga writes its own plan: one row per saga, one row per target, updated
// as each is settled. Resuming is then a read rather than a re-derivation, and a
// re-run cannot decide the remaining work differently from the run it is
// resuming.
//
// # Idempotence is on the target, not on the attempt
//
// A step's unique key is (saga, kind, symbol, target order). Re-adding a step the
// saga already knows about returns the existing row rather than creating a second
// one, so a resumed enumeration that sees the same open order again does not
// queue a second cancel for it.

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

// schemaV4 adds the flatten saga tables. Additive, per schema.go's rules.
const schemaV4 = `
-- One flatten-all run. There is at most one unfinished row at a time; the engine
-- enforces that, and a resumed run reuses the row rather than starting another.
CREATE TABLE flatten_sagas (
	id          TEXT PRIMARY KEY,        -- caller-supplied stable id
	account_ref TEXT NOT NULL,
	phase       TEXT NOT NULL,
	reason      TEXT NOT NULL DEFAULT '', -- why somebody flattened
	operator    TEXT NOT NULL DEFAULT '',
	dry_run     INTEGER NOT NULL DEFAULT 0,
	started_at  TEXT NOT NULL,
	updated_at  TEXT NOT NULL,
	finished_at TEXT,
	detail      TEXT NOT NULL DEFAULT ''
) STRICT;

CREATE INDEX idx_flatten_phase ON flatten_sagas(phase);

-- One thing the saga has to do: cancel this order, or liquidate this holding.
CREATE TABLE flatten_steps (
	id              INTEGER PRIMARY KEY AUTOINCREMENT,
	saga_id         TEXT NOT NULL REFERENCES flatten_sagas(id),
	kind            TEXT NOT NULL,             -- 'CANCEL' | 'LIQUIDATE'
	market          TEXT NOT NULL DEFAULT '',
	symbol          TEXT NOT NULL,
	target_order_id TEXT NOT NULL DEFAULT '',  -- CANCEL: the order being cancelled
	side            TEXT NOT NULL DEFAULT '',
	quantity        TEXT NOT NULL DEFAULT '',  -- decimal string
	price           TEXT NOT NULL DEFAULT '',  -- decimal string
	currency        TEXT NOT NULL DEFAULT '',
	state           TEXT NOT NULL,             -- see the FlattenStep* constants
	intent_id       TEXT NOT NULL DEFAULT '',
	attempt_id      TEXT NOT NULL DEFAULT '',
	reason_code     TEXT NOT NULL DEFAULT '',
	detail          TEXT NOT NULL DEFAULT '',
	created_at      TEXT NOT NULL,
	updated_at      TEXT NOT NULL,
	UNIQUE(saga_id, kind, symbol, target_order_id)
) STRICT;

CREATE INDEX idx_flatten_steps_saga ON flatten_steps(saga_id, state);
`

// Saga phases. They are ordered, and the engine only moves forward.
const (
	// FlattenPhaseBlocking — new entries are being shut off.
	FlattenPhaseBlocking = "BLOCKING"
	// FlattenPhaseCancelling — outstanding orders are being cancelled.
	FlattenPhaseCancelling = "CANCELLING"
	// FlattenPhaseCancelled — every cancel is settled (task 4.4 ends here).
	FlattenPhaseCancelled = "CANCELLED"
	// FlattenPhaseStabilising — waiting for the account to stop moving (4.5).
	FlattenPhaseStabilising = "STABILISING"
	// FlattenPhaseLiquidating — reduce-only exits are being submitted (4.5).
	FlattenPhaseLiquidating = "LIQUIDATING"
	// FlattenPhaseVerifying — re-reconciling for a remainder (4.5).
	FlattenPhaseVerifying = "VERIFYING"
	// FlattenPhaseComplete — nothing is left.
	FlattenPhaseComplete = "COMPLETE"
	// FlattenPhaseStalled — the saga could not finish and needs an operator.
	FlattenPhaseStalled = "STALLED"
)

// Step kinds.
const (
	FlattenStepCancel    = "CANCEL"
	FlattenStepLiquidate = "LIQUIDATE"
)

// Step states.
const (
	// FlattenStepPending — planned, not yet acted on.
	FlattenStepPending = "PENDING"
	// FlattenStepDone — the mutation settled successfully.
	FlattenStepDone = "DONE"
	// FlattenStepInDoubt — the mutation's outcome is unresolved. A symbol with a
	// step in this state must not be liquidated: see the oversell rule in
	// internal/flatten.
	FlattenStepInDoubt = "IN_DOUBT"
	// FlattenStepFailed — the broker definitively refused, or the target was
	// already gone. Not an error for the saga: the goal is that the order is not
	// live, and it is not.
	FlattenStepFailed = "FAILED"
	// FlattenStepHeld — deliberately not acted on yet, because acting would risk
	// an oversell.
	FlattenStepHeld = "HELD"
)

// ErrFlattenNotFound means no saga has that id.
var ErrFlattenNotFound = errors.New("journal: no such flatten saga")

// FlattenSaga is one run.
type FlattenSaga struct {
	ID         string
	AccountRef string
	Phase      string
	Reason     string
	Operator   string
	DryRun     bool
	StartedAt  string
	UpdatedAt  string
	FinishedAt string
	Detail     string
}

// Active reports whether the saga is still running.
func (s FlattenSaga) Active() bool {
	switch s.Phase {
	case FlattenPhaseComplete, FlattenPhaseStalled:
		return false
	default:
		return true
	}
}

// FlattenStep is one target.
type FlattenStep struct {
	ID            int64
	SagaID        string
	Kind          string
	Market        string
	Symbol        string
	TargetOrderID string
	Side          string
	Quantity      string
	Price         string
	Currency      string
	State         string
	IntentID      string
	AttemptID     string
	ReasonCode    string
	Detail        string
}

// StartFlatten opens a saga, or returns the unfinished one that already exists.
//
// Returning the existing saga rather than refusing is what makes a re-run a
// resume: the operator who re-invokes flatten after a crash is not starting a
// second exit, they are continuing the first one, and treating it as new would
// lose the record of what was already cancelled.
func (j *Journal) StartFlatten(ctx context.Context, s FlattenSaga) (FlattenSaga, error) {
	if strings.TrimSpace(s.ID) == "" {
		return FlattenSaga{}, errors.New("journal: a flatten saga needs an id")
	}
	if strings.TrimSpace(s.AccountRef) == "" {
		return FlattenSaga{}, errors.New("journal: a flatten saga needs an account reference")
	}
	if existing, err := j.ActiveFlatten(ctx); err == nil {
		return existing, nil
	} else if !errors.Is(err, ErrFlattenNotFound) {
		return FlattenSaga{}, err
	}

	now := RFC3339(j.clk.Now())
	phase := s.Phase
	if phase == "" {
		phase = FlattenPhaseBlocking
	}
	if _, err := j.db.ExecContext(ctx,
		`INSERT INTO flatten_sagas(id, account_ref, phase, reason, operator, dry_run, started_at, updated_at)
		 VALUES(?, ?, ?, ?, ?, ?, ?, ?)`,
		s.ID, s.AccountRef, phase, s.Reason, s.Operator, boolToInt(s.DryRun), now, now); err != nil {
		return FlattenSaga{}, fmt.Errorf("journal: starting flatten saga %s: %w", s.ID, err)
	}
	return j.LookupFlatten(ctx, s.ID)
}

// ActiveFlatten returns the unfinished saga, if there is one.
func (j *Journal) ActiveFlatten(ctx context.Context) (FlattenSaga, error) {
	rows, err := j.db.QueryContext(ctx, flattenSelect+
		` WHERE phase NOT IN (?, ?) ORDER BY started_at DESC LIMIT 1`,
		FlattenPhaseComplete, FlattenPhaseStalled)
	if err != nil {
		return FlattenSaga{}, fmt.Errorf("journal: looking for an active flatten saga: %w", err)
	}
	defer rows.Close()
	sagas, err := scanFlattenSagas(rows)
	if err != nil {
		return FlattenSaga{}, err
	}
	if len(sagas) == 0 {
		return FlattenSaga{}, ErrFlattenNotFound
	}
	return sagas[0], nil
}

// LookupFlatten reads one saga.
func (j *Journal) LookupFlatten(ctx context.Context, id string) (FlattenSaga, error) {
	rows, err := j.db.QueryContext(ctx, flattenSelect+` WHERE id = ?`, id)
	if err != nil {
		return FlattenSaga{}, fmt.Errorf("journal: reading flatten saga %s: %w", id, err)
	}
	defer rows.Close()
	sagas, err := scanFlattenSagas(rows)
	if err != nil {
		return FlattenSaga{}, err
	}
	if len(sagas) == 0 {
		return FlattenSaga{}, fmt.Errorf("%w: %s", ErrFlattenNotFound, id)
	}
	return sagas[0], nil
}

// SetFlattenPhase advances the saga.
func (j *Journal) SetFlattenPhase(ctx context.Context, id, phase, detail string) error {
	now := RFC3339(j.clk.Now())
	var finished any
	if phase == FlattenPhaseComplete || phase == FlattenPhaseStalled {
		finished = now
	}
	res, err := j.db.ExecContext(ctx,
		`UPDATE flatten_sagas SET phase = ?, detail = ?, updated_at = ?, finished_at = ? WHERE id = ?`,
		phase, detail, now, finished, id)
	if err != nil {
		return fmt.Errorf("journal: moving flatten saga %s to %s: %w", id, phase, err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("%w: %s", ErrFlattenNotFound, id)
	}
	return nil
}

// AddFlattenStep records a target, or returns the existing row for it.
func (j *Journal) AddFlattenStep(ctx context.Context, step FlattenStep) (FlattenStep, error) {
	if strings.TrimSpace(step.SagaID) == "" || strings.TrimSpace(step.Symbol) == "" {
		return FlattenStep{}, errors.New("journal: a flatten step needs a saga and a symbol")
	}
	if step.Kind != FlattenStepCancel && step.Kind != FlattenStepLiquidate {
		return FlattenStep{}, fmt.Errorf("journal: %q is not a flatten step kind", step.Kind)
	}
	state := step.State
	if state == "" {
		state = FlattenStepPending
	}
	now := RFC3339(j.clk.Now())

	if existing, err := j.lookupFlattenStep(ctx, step.SagaID, step.Kind, step.Symbol, step.TargetOrderID); err == nil {
		return existing, nil
	} else if !errors.Is(err, ErrFlattenNotFound) {
		return FlattenStep{}, err
	}

	if _, err := j.db.ExecContext(ctx,
		`INSERT INTO flatten_steps(saga_id, kind, market, symbol, target_order_id, side,
		                           quantity, price, currency, state, created_at, updated_at)
		 VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		step.SagaID, step.Kind, step.Market, step.Symbol, step.TargetOrderID, step.Side,
		step.Quantity, step.Price, step.Currency, state, now, now); err != nil {
		return FlattenStep{}, fmt.Errorf("journal: recording a flatten step for %s: %w", step.Symbol, err)
	}
	return j.lookupFlattenStep(ctx, step.SagaID, step.Kind, step.Symbol, step.TargetOrderID)
}

// UpdateFlattenStep records a step's outcome.
func (j *Journal) UpdateFlattenStep(ctx context.Context, id int64, state, intentID, attemptID, reasonCode, detail string) error {
	now := RFC3339(j.clk.Now())
	res, err := j.db.ExecContext(ctx,
		`UPDATE flatten_steps
		    SET state = ?, intent_id = ?, attempt_id = ?, reason_code = ?, detail = ?, updated_at = ?
		  WHERE id = ?`,
		state, intentID, attemptID, reasonCode, detail, now, id)
	if err != nil {
		return fmt.Errorf("journal: updating flatten step %d: %w", id, err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("%w: step %d", ErrFlattenNotFound, id)
	}
	return nil
}

// FlattenSteps lists a saga's steps in creation order.
func (j *Journal) FlattenSteps(ctx context.Context, sagaID string) ([]FlattenStep, error) {
	rows, err := j.db.QueryContext(ctx, flattenStepSelect+` WHERE saga_id = ? ORDER BY id`, sagaID)
	if err != nil {
		return nil, fmt.Errorf("journal: listing flatten steps for %s: %w", sagaID, err)
	}
	defer rows.Close()
	return scanFlattenSteps(rows)
}

// UnsettledCancelSymbols lists the symbols whose cancel step has not settled.
//
// This is the oversell guard's input: a symbol whose cancel outcome we cannot
// account for may still have a live order against it, and selling the full
// holding while that order is also live is how a flatten becomes a short
// position.
func (j *Journal) UnsettledCancelSymbols(ctx context.Context, sagaID string) (map[string]string, error) {
	rows, err := j.db.QueryContext(ctx,
		`SELECT symbol, state FROM flatten_steps
		  WHERE saga_id = ? AND kind = ? AND state IN (?, ?, ?)`,
		sagaID, FlattenStepCancel, FlattenStepPending, FlattenStepInDoubt, FlattenStepHeld)
	if err != nil {
		return nil, fmt.Errorf("journal: listing unsettled cancels for %s: %w", sagaID, err)
	}
	defer rows.Close()

	out := map[string]string{}
	for rows.Next() {
		var symbol, state string
		if err := rows.Scan(&symbol, &state); err != nil {
			return nil, fmt.Errorf("journal: scanning an unsettled cancel: %w", err)
		}
		out[strings.ToUpper(symbol)] = state
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("journal: reading unsettled cancels: %w", err)
	}
	return out, nil
}

// --- internals --------------------------------------------------------------

const flattenSelect = `SELECT id, account_ref, phase, reason, operator, dry_run,
       started_at, updated_at, finished_at, detail FROM flatten_sagas`

const flattenStepSelect = `SELECT id, saga_id, kind, market, symbol, target_order_id, side,
       quantity, price, currency, state, intent_id, attempt_id, reason_code, detail
  FROM flatten_steps`

func (j *Journal) lookupFlattenStep(ctx context.Context, sagaID, kind, symbol, targetOrderID string) (FlattenStep, error) {
	rows, err := j.db.QueryContext(ctx, flattenStepSelect+
		` WHERE saga_id = ? AND kind = ? AND symbol = ? AND target_order_id = ?`,
		sagaID, kind, symbol, targetOrderID)
	if err != nil {
		return FlattenStep{}, fmt.Errorf("journal: reading a flatten step: %w", err)
	}
	defer rows.Close()
	steps, err := scanFlattenSteps(rows)
	if err != nil {
		return FlattenStep{}, err
	}
	if len(steps) == 0 {
		return FlattenStep{}, ErrFlattenNotFound
	}
	return steps[0], nil
}

func scanFlattenSagas(rows *sql.Rows) ([]FlattenSaga, error) {
	var out []FlattenSaga
	for rows.Next() {
		var (
			s        FlattenSaga
			dryRun   int
			finished sql.NullString
		)
		if err := rows.Scan(&s.ID, &s.AccountRef, &s.Phase, &s.Reason, &s.Operator, &dryRun,
			&s.StartedAt, &s.UpdatedAt, &finished, &s.Detail); err != nil {
			return nil, fmt.Errorf("journal: scanning a flatten saga: %w", err)
		}
		s.DryRun = dryRun != 0
		if finished.Valid {
			s.FinishedAt = finished.String
		}
		out = append(out, s)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("journal: reading flatten sagas: %w", err)
	}
	return out, nil
}

func scanFlattenSteps(rows *sql.Rows) ([]FlattenStep, error) {
	var out []FlattenStep
	for rows.Next() {
		var s FlattenStep
		if err := rows.Scan(&s.ID, &s.SagaID, &s.Kind, &s.Market, &s.Symbol, &s.TargetOrderID,
			&s.Side, &s.Quantity, &s.Price, &s.Currency, &s.State, &s.IntentID, &s.AttemptID,
			&s.ReasonCode, &s.Detail); err != nil {
			return nil, fmt.Errorf("journal: scanning a flatten step: %w", err)
		}
		out = append(out, s)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("journal: reading flatten steps: %w", err)
	}
	return out, nil
}

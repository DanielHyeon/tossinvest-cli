package journal

// fills.go is the durable side of cumulative-snapshot fill detection
// (harden-execution-base task 3.2).
//
// # Why a snapshot table and not a fill log
//
// The official API has no per-fill identifier. It reports a *cumulative*
// filledQuantity and an average price, and nothing else (design D4). So there is
// no event to deduplicate by id, and the only way to know what is new is to keep
// the last observation and subtract.
//
// That makes the last observation a piece of durable state, not a cache: lose it
// and the next poll after a restart reports the whole cumulative quantity as a
// brand-new fill, doubling the engine's idea of its position. Hence a table, in
// the same database and the same transaction discipline as the intents.
//
// # The rules, and why they live inside the transaction
//
//	delta > 0   → a new fill: the snapshot advances and one event is appended
//	delta == 0  → nothing happened: identical snapshots are a no-op
//	delta < 0   → UNKNOWN_BROKER_STATE, fail closed, snapshot NOT advanced
//	older broker timestamp with a different quantity → same, fail closed
//
// All four are evaluated inside the BEGIN IMMEDIATE transaction that writes the
// result. Computing the delta outside it would let two cycles read the same prior
// snapshot and each append "their" fill, which is exactly the double-count the
// table exists to prevent.
//
// A shrinking cumulative fill is refused rather than corrected because there is no
// safe correction. Either the broker is reporting a different order than we think,
// or our record is wrong; both mean the engine's model of a live account is
// broken, and trading through that is worse than stopping.
//
// # Corrections (extend-execution-contract task 5.3, design D7)
//
// A fifth case sits between "nothing happened" and "a new fill":
//
//	delta == 0 and a different average price or filled amount → EXECUTION_CORRECTION
//
// The broker's execution model is cumulative and carries no per-fill identifier,
// and `OrderExecution.averageFilledPrice` is documented as "부분 체결 시 체결된
// 건의 평균" — it moves with every partial fill — while `filledAmount` is the
// total executed amount (docs/migration/openapi.latest.json, both nullable
// decimal strings). So a restatement at an unchanged quantity is a real event
// with no quantity delta: it belongs in its own table, not in fill_events.
//
// It is detected and written here, in the same BEGIN IMMEDIATE, because this is
// the only place the previous snapshot still exists. Comparing outside the
// transaction would let two cycles each decide "the average changed" from the
// same prior and write the correction twice.
//
// # The atomic apply point (add-core-domain task 0.3, design D7)
//
// The transaction below is also where the position projection and the exit
// state move. RecordFill calls the injected apply functions inside its own
// BEGIN IMMEDIATE, after the snapshot has advanced and before the commit, so a
// fill and everything it implies land together. The injection, the handle those
// functions get and the rules they run under are in apply_hook.go; what this
// file contributes is the transaction and the point inside it.
//
// # Opaque identifiers
//
// `orderId` is an opaque token — openapi contracts no shape for it — so it is
// stored and looked up exactly as received. Trimming survives only in the
// emptiness check, where it asks "is there a name here at all"
// (order-execution: 저장 SHALL 원문, 비교 SHALL 바이트 동일).

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

// Reason codes written by the fill ledger. Stable strings: they reach the entry
// gate, the alerting path and the Phase 2 ledger import.
const (
	// ReasonFillDecreased: a cumulative filled quantity went backwards.
	ReasonFillDecreased = "fill_snapshot_decreased"
	// ReasonFillOutOfOrder: a snapshot the broker timestamped earlier than the
	// one already recorded carries a different quantity.
	ReasonFillOutOfOrder = "fill_snapshot_out_of_order"
	// ReasonFillStateUnknown: the caller's own derivation failed closed.
	ReasonFillStateUnknown = "fill_snapshot_state_unknown"
)

// ErrFillNotFound means no snapshot has been recorded for that order.
var ErrFillNotFound = errors.New("journal: no fill snapshot for that order")

// FillObservation is one cumulative snapshot offered to the ledger.
//
// Quantities and the average price are decimal strings for the same reason the
// intent's are: this is an audit record, and a broker's decimal quantity has no
// exact binary float form.
type FillObservation struct {
	OrderID string
	Symbol  string
	Market  string
	// State is the caller's derived broker state (internal/brokerstate).
	State string
	// Terminal reports that the order can no longer change at the broker.
	Terminal bool
	// FailClosed is the caller's own verdict that the payload is not
	// trustworthy. It is recorded, and it stops the snapshot advancing, exactly
	// like a decrease found here.
	FailClosed bool
	// Reason and Detail explain a caller-side FailClosed.
	Reason string
	Detail string

	Quantity       string // ordered quantity, decimal string
	FilledQuantity string // cumulative filled quantity, decimal string
	AveragePrice   string // decimal string, "" when the broker gave none
	// FilledAmount is `execution.filledAmount` — the total executed amount in the
	// order's native currency (openapi: nullable decimal string). "" when the
	// broker gave none, the same convention AveragePrice uses. Without it an
	// amount-only restatement is invisible.
	FilledAmount string

	// AccountRef is the account the order belongs to. Optional: when it is empty
	// the journal derives it from the confirmed attempt that named the order,
	// which is the only account dimension a detector Snapshot has.
	AccountRef string

	// BrokerVisibleAt is when the broker made this state observable, RFC3339
	// UTC. Empty means the payload carried no usable timestamp.
	BrokerVisibleAt string
	// ObservedAt is when the poller read it, RFC3339 UTC.
	ObservedAt string
}

// FillResult is what recording one observation did.
type FillResult struct {
	OrderID string
	// Delta is the newly filled quantity as a decimal string. "0" when nothing
	// is new.
	Delta string
	// DeltaQuantity is the same number as a float, for arithmetic.
	DeltaQuantity float64
	// Changed reports whether the durable row was modified at all.
	Changed bool
	// FailClosed reports a refused snapshot. The stored snapshot is unchanged.
	FailClosed bool
	Reason     string
	Detail     string
	// Corrected reports an EXECUTION_CORRECTION: the cumulative quantity did not
	// move but the average price or the filled amount did. Delta stays "0".
	//
	// A replay the UNIQUE key absorbs still reports true. The observation *is* a
	// correction; whether its audit row was already on disk is the table's
	// business, and reporting false would make a crash look like a no-op.
	Corrected bool
	// CommittedAt is when the transaction committed, RFC3339 UTC. It is the far
	// end of the fill-detection SLO measurement.
	CommittedAt string

	// ReleasedReservations are the risk reservations this observation freed,
	// because it derived the order as terminal (design D5, task 3.2). The
	// release happened inside the same transaction as the snapshot.
	ReleasedReservations []ReservationRelease
	// ReservationAlerts are holds this observation could *not* free and that
	// only an operator can. A fail-closed snapshot produces them: the hold is
	// correct, and somebody has to be told it exists.
	ReservationAlerts []ReservationAlert
}

// FillSnapshotRecord is a stored snapshot.
type FillSnapshotRecord struct {
	OrderID         string
	Symbol          string
	Market          string
	State           string
	Terminal        bool
	FailClosed      bool
	Quantity        string
	FilledQuantity  string
	AveragePrice    string
	FilledAmount    string
	BrokerVisibleAt string
	ObservedAt      string
	CommittedAt     string
	ReasonCode      string
	Detail          string
}

// ExecutionCorrection is one recorded restatement of an already-observed
// execution: same cumulative quantity, a different average price or amount.
//
// It is not a fill and carries no delta. The previous values travel with it
// because the point of the record is what changed, and the snapshot only keeps
// the latest.
type ExecutionCorrection struct {
	ID         string
	AccountRef string
	OrderID    string
	// Prev* and New* are decimal strings; "" means the broker reported none.
	PrevAveragePrice string
	NewAveragePrice  string
	PrevFilledAmount string
	NewFilledAmount  string
	// CumulativeQuantity is unchanged by definition — it is what makes this a
	// correction rather than a fill.
	CumulativeQuantity string
	ObservedAt         string
}

// FillEvent is one appended positive delta.
type FillEvent struct {
	ID                 int64
	OrderID            string
	Symbol             string
	Market             string
	DeltaQuantity      string
	CumulativeQuantity string
	AveragePrice       string
	BrokerVisibleAt    string
	CommittedAt        string
}

// TrackedFillOrder is an order the detector must keep reading.
type TrackedFillOrder struct {
	OrderID string
	Symbol  string
	Market  string
	// SuccessorOrderID is the order an amendment created to carry this order's
	// remainder, empty when there is none.
	SuccessorOrderID string
}

// RecordFill applies one cumulative snapshot.
//
// It is idempotent by construction: the same observation applied twice moves
// nothing the second time, whether the repeat came from the poll loop, an SSE
// re-fetch or the first cycle after a restart.
func (j *Journal) RecordFill(ctx context.Context, obs FillObservation) (FillResult, error) {
	if strings.TrimSpace(obs.OrderID) == "" {
		return FillResult{}, fmt.Errorf("%w: a fill snapshot needs an order id", ErrInvalidRequest)
	}
	filled, err := strconv.ParseFloat(strings.TrimSpace(orZero(obs.FilledQuantity)), 64)
	if err != nil {
		return FillResult{}, fmt.Errorf("%w: filled quantity %q is not a decimal",
			ErrInvalidRequest, obs.FilledQuantity)
	}
	if math.IsNaN(filled) || math.IsInf(filled, 0) {
		return FillResult{}, fmt.Errorf("%w: filled quantity %q is not finite",
			ErrInvalidRequest, obs.FilledQuantity)
	}

	now := j.nowString()
	// Verbatim: the identifier is opaque, so the row is keyed by what the broker
	// sent. TrimSpace above judged emptiness and nothing else.
	orderID := obs.OrderID
	res := FillResult{OrderID: orderID, Delta: "0", CommittedAt: now}

	tx, err := j.db.BeginTx(ctx, nil) // BEGIN IMMEDIATE
	if err != nil {
		return res, fmt.Errorf("journal: starting the fill transaction for %s: %w", orderID, err)
	}
	defer tx.Rollback()

	prev, err := scanFillSnapshot(tx.QueryRowContext(ctx, fillSelect+" WHERE order_id = ?", orderID))
	switch {
	case err == nil:
	case errors.Is(err, ErrFillNotFound):
		prev = FillSnapshotRecord{}
	default:
		return res, err
	}
	hadPrev := prev.OrderID != ""

	prevFilled := 0.0
	if hadPrev {
		prevFilled, _ = strconv.ParseFloat(strings.TrimSpace(orZero(prev.FilledQuantity)), 64)
	}

	// 1. The caller's own fail-closed verdict, and the two this table detects.
	if refusal := classifyFillRefusal(obs, prev, hadPrev, filled, prevFilled); refusal != nil {
		res.FailClosed = true
		res.Reason = refusal.reason
		res.Detail = refusal.detail
		// The refusal itself is durable — an operator has to be able to see what
		// was rejected — but the snapshot is not advanced, so the next consistent
		// observation still measures its delta from the last trusted quantity.
		if err := markFillRefused(ctx, tx, orderID, obs, refusal, now, hadPrev); err != nil {
			return res, err
		}
		// Nothing is released on a refusal — that is the point. A snapshot the
		// derivation could not explain (CLOSED, nothing filled, no cancellation:
		// what an expiry would look like, [미측정 — 2b 2.1]) leaves the hold
		// standing, and the operator is told which holds it is standing on
		// (design D5: 만료 추정으로 해제하지 않는다).
		alerts, err := alertsForOrder(ctx, tx, orderID, ReasonFillStateUnknown, refusal.detail)
		if err != nil {
			return res, err
		}
		res.ReservationAlerts = alerts
		if err := tx.Commit(); err != nil {
			return res, fmt.Errorf("journal: committing the refused snapshot of %s: %w", orderID, err)
		}
		res.Changed = true
		return res, nil
	}

	delta := filled - prevFilled
	if nearlyZero(delta, filled) {
		delta = 0
	}

	// 2. A byte-identical re-observation touches nothing at all. That is what
	//    makes "the same snapshot arrived from the poller and from an SSE
	//    re-fetch" provably a no-op rather than merely harmless.
	if hadPrev && delta == 0 && sameSnapshot(obs, prev) {
		if err := tx.Commit(); err != nil {
			return res, fmt.Errorf("journal: closing the no-op snapshot of %s: %w", orderID, err)
		}
		res.CommittedAt = prev.CommittedAt
		return res, nil
	}

	// 3. A correction: the cumulative quantity did not move, but the average price
	//    or the filled amount did. It is decided *before* the upsert below, which
	//    is the last moment the previous values exist.
	correction := hadPrev && delta == 0 &&
		(obs.AveragePrice != prev.AveragePrice || obs.FilledAmount != prev.FilledAmount)

	// 4. Advance. The average price is *replaced*, never accumulated: it is a
	//    property of the whole filled quantity, so adding one in would double
	//    count the fills that produced it. The amount follows the same rule.
	if _, err := tx.ExecContext(ctx,
		`INSERT INTO fill_snapshots
		   (order_id, symbol, market, state, terminal, fail_closed, quantity, filled_quantity,
		    average_price, filled_amount, broker_visible_at, observed_at, committed_at,
		    reason_code, detail)
		 VALUES (?,?,?,?,?,0,?,?,?,?,?,?,?,'','')
		 ON CONFLICT(order_id) DO UPDATE SET
		    symbol = excluded.symbol, market = excluded.market, state = excluded.state,
		    terminal = excluded.terminal, fail_closed = 0, quantity = excluded.quantity,
		    filled_quantity = excluded.filled_quantity, average_price = excluded.average_price,
		    filled_amount = excluded.filled_amount,
		    broker_visible_at = excluded.broker_visible_at, observed_at = excluded.observed_at,
		    committed_at = excluded.committed_at, reason_code = '', detail = ''`,
		orderID, obs.Symbol, obs.Market, obs.State, boolToInt(obs.Terminal),
		orZero(obs.Quantity), orZero(obs.FilledQuantity), obs.AveragePrice, obs.FilledAmount,
		obs.BrokerVisibleAt, obs.ObservedAt, now); err != nil {
		return res, fmt.Errorf("journal: recording the fill snapshot of %s: %w", orderID, err)
	}

	if correction {
		if err := recordExecutionCorrection(ctx, tx, ExecutionCorrection{
			AccountRef:         obs.AccountRef,
			OrderID:            orderID,
			PrevAveragePrice:   prev.AveragePrice,
			NewAveragePrice:    obs.AveragePrice,
			PrevFilledAmount:   prev.FilledAmount,
			NewFilledAmount:    obs.FilledAmount,
			CumulativeQuantity: orZero(obs.FilledQuantity),
			ObservedAt:         firstNonEmpty(obs.ObservedAt, now),
		}); err != nil {
			return res, err
		}
		res.Corrected = true
	}

	if delta > 0 {
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO fill_events
			   (order_id, symbol, market, delta_quantity, cumulative_quantity,
			    average_price, broker_visible_at, committed_at)
			 VALUES (?,?,?,?,?,?,?,?)`,
			orderID, obs.Symbol, obs.Market, decimalString(delta), orZero(obs.FilledQuantity),
			obs.AveragePrice, obs.BrokerVisibleAt, now); err != nil {
			return res, fmt.Errorf("journal: appending the fill of %s: %w", orderID, err)
		}
		res.Delta = decimalString(delta)
		res.DeltaQuantity = delta
	}

	// 5. A *derived* terminal state frees the decision's holds, in this same
	//    transaction (design D5). The verdict is the caller's derivation
	//    (internal/brokerstate), never a guess made here: this branch is
	//    unreachable for a fail-closed observation, which returned above.
	if obs.Terminal {
		released, err := releaseReservationsForOrder(ctx, tx, orderID, ReleaseReasonBrokerTerminal,
			fmt.Sprintf("order %s derived terminal as %s", orderID, obs.State), now)
		if err != nil {
			return res, err
		}
		res.ReleasedReservations = released
	}

	// 6. The atomic apply point (apply_hook.go, design D7). The injected
	//    projection and exit functions run here, inside this transaction, so the
	//    snapshot, the position and the exit state commit together or not at all.
	//    A hook error aborts everything above: the deferred rollback leaves the
	//    snapshot un-advanced, and the next observation measures its delta from
	//    the last quantity the projection actually saw.
	if err := j.runApplyHooks(ctx, tx, AppliedFill{
		OrderID:            orderID,
		AccountRef:         obs.AccountRef,
		Symbol:             obs.Symbol,
		Market:             obs.Market,
		State:              obs.State,
		Terminal:           obs.Terminal,
		Delta:              res.Delta,
		CumulativeQuantity: orZero(obs.FilledQuantity),
		AveragePrice:       obs.AveragePrice,
		FilledAmount:       obs.FilledAmount,
		OrderedQuantity:    obs.Quantity,
		// `prev` is the row this transaction has just overwritten. It is read
		// here and nowhere else because this is the last scope in which it
		// exists (add-core-domain task 6.1: the projection's cost basis is the
		// change in the order's `filled × average`, which needs both ends).
		PrevCumulativeQuantity: prev.FilledQuantity,
		PrevAveragePrice:       prev.AveragePrice,
		Corrected:              res.Corrected,
		BrokerVisibleAt:        obs.BrokerVisibleAt,
		CommittedAt:            now,
	}); err != nil {
		return res, err
	}

	if err := tx.Commit(); err != nil {
		return res, fmt.Errorf("journal: committing the fill snapshot of %s: %w", orderID, err)
	}
	res.Changed = true
	return res, nil
}

type fillRefusal struct {
	reason string
	detail string
}

// classifyFillRefusal applies the three fail-closed rules, in the order their
// evidence is strongest.
func classifyFillRefusal(obs FillObservation, prev FillSnapshotRecord, hadPrev bool,
	filled, prevFilled float64) *fillRefusal {
	if obs.FailClosed {
		reason := obs.Reason
		if reason == "" {
			reason = ReasonFillStateUnknown
		}
		detail := obs.Detail
		if detail == "" {
			detail = "the caller could not derive a trustworthy state for order " + obs.OrderID
		}
		return &fillRefusal{reason: reason, detail: detail}
	}
	if !hadPrev {
		return nil
	}
	if filled < prevFilled && !nearlyZero(filled-prevFilled, prevFilled) {
		return &fillRefusal{
			reason: ReasonFillDecreased,
			detail: fmt.Sprintf(
				"order %s reported %s filled after %s: a cumulative quantity cannot shrink",
				obs.OrderID, orZero(obs.FilledQuantity), orZero(prev.FilledQuantity)),
		}
	}
	// An older broker timestamp carrying a different quantity means the two
	// observations cannot both be true. Equal quantities are just a re-read of the
	// same state and are fine.
	if obs.BrokerVisibleAt != "" && prev.BrokerVisibleAt != "" &&
		obs.BrokerVisibleAt < prev.BrokerVisibleAt &&
		!nearlyZero(filled-prevFilled, prevFilled) {
		return &fillRefusal{
			reason: ReasonFillOutOfOrder,
			detail: fmt.Sprintf(
				"order %s reported %s filled as of %s, older than the %s already recorded as of %s",
				obs.OrderID, orZero(obs.FilledQuantity), obs.BrokerVisibleAt,
				orZero(prev.FilledQuantity), prev.BrokerVisibleAt),
		}
	}
	return nil
}

// markFillRefused records the refusal without letting it move the trusted
// quantity. A first-ever observation that fails closed still gets a row, so the
// order is visible to the operator rather than invisible until it behaves.
func markFillRefused(ctx context.Context, tx *sql.Tx, orderID string, obs FillObservation,
	refusal *fillRefusal, now string, hadPrev bool) error {
	if hadPrev {
		if _, err := tx.ExecContext(ctx,
			`UPDATE fill_snapshots
			    SET fail_closed = 1, reason_code = ?, detail = ?, observed_at = ?, committed_at = ?
			  WHERE order_id = ?`,
			refusal.reason, refusal.detail, obs.ObservedAt, now, orderID); err != nil {
			return fmt.Errorf("journal: recording the refused snapshot of %s: %w", orderID, err)
		}
		return nil
	}
	if _, err := tx.ExecContext(ctx,
		`INSERT INTO fill_snapshots
		   (order_id, symbol, market, state, terminal, fail_closed, quantity, filled_quantity,
		    average_price, broker_visible_at, observed_at, committed_at, reason_code, detail)
		 VALUES (?,?,?,?,0,1,?,'0','','',?,?,?,?)`,
		orderID, obs.Symbol, obs.Market, obs.State, orZero(obs.Quantity),
		obs.ObservedAt, now, refusal.reason, refusal.detail); err != nil {
		return fmt.Errorf("journal: recording the refused snapshot of %s: %w", orderID, err)
	}
	return nil
}

// sameSnapshot reports whether an observation says exactly what the stored row
// already says about the broker's state. Timestamps of *our* reading are excluded
// on purpose: when we looked does not change what is true.
func sameSnapshot(obs FillObservation, prev FillSnapshotRecord) bool {
	return orZero(obs.FilledQuantity) == orZero(prev.FilledQuantity) &&
		obs.AveragePrice == prev.AveragePrice &&
		// The amount is part of "what the broker says" for the same reason the
		// average is: leave it out and an amount-only correction takes the
		// no-op path and is never recorded at all.
		obs.FilledAmount == prev.FilledAmount &&
		obs.State == prev.State &&
		obs.Terminal == prev.Terminal &&
		orZero(obs.Quantity) == orZero(prev.Quantity) &&
		!prev.FailClosed
}

const fillSelect = `SELECT order_id, symbol, market, state, terminal, fail_closed, quantity,
	filled_quantity, average_price, filled_amount, broker_visible_at, observed_at, committed_at,
	reason_code, detail FROM fill_snapshots`

func scanFillSnapshot(row rowScanner) (FillSnapshotRecord, error) {
	var (
		rec                  FillSnapshotRecord
		terminal, failClosed int
	)
	err := row.Scan(&rec.OrderID, &rec.Symbol, &rec.Market, &rec.State, &terminal, &failClosed,
		&rec.Quantity, &rec.FilledQuantity, &rec.AveragePrice, &rec.FilledAmount,
		&rec.BrokerVisibleAt, &rec.ObservedAt, &rec.CommittedAt, &rec.ReasonCode, &rec.Detail)
	if errors.Is(err, sql.ErrNoRows) {
		return FillSnapshotRecord{}, ErrFillNotFound
	}
	if err != nil {
		return FillSnapshotRecord{}, fmt.Errorf("journal: reading a fill snapshot: %w", err)
	}
	rec.Terminal = terminal != 0
	rec.FailClosed = failClosed != 0
	return rec, nil
}

// LookupFill returns the stored snapshot of one order.
//
// The identifier is used as given: the rows are keyed by what the broker sent,
// so normalising the lookup would miss them.
func (j *Journal) LookupFill(ctx context.Context, orderID string) (FillSnapshotRecord, error) {
	return scanFillSnapshot(j.db.QueryRowContext(ctx, fillSelect+" WHERE order_id = ?", orderID))
}

// FillEvents returns the appended fills of one order, oldest first.
func (j *Journal) FillEvents(ctx context.Context, orderID string) ([]FillEvent, error) {
	rows, err := j.db.QueryContext(ctx,
		`SELECT id, order_id, symbol, market, delta_quantity, cumulative_quantity,
		        average_price, broker_visible_at, committed_at
		   FROM fill_events WHERE order_id = ? ORDER BY id`, orderID)
	if err != nil {
		return nil, fmt.Errorf("journal: reading the fills of %s: %w", orderID, err)
	}
	defer rows.Close()

	var out []FillEvent
	for rows.Next() {
		var e FillEvent
		if err := rows.Scan(&e.ID, &e.OrderID, &e.Symbol, &e.Market, &e.DeltaQuantity,
			&e.CumulativeQuantity, &e.AveragePrice, &e.BrokerVisibleAt, &e.CommittedAt); err != nil {
			return nil, fmt.Errorf("journal: reading the fills of %s: %w", orderID, err)
		}
		out = append(out, e)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("journal: reading the fills of %s: %w", orderID, err)
	}
	return out, nil
}

// recordExecutionCorrection writes one correction inside the caller's fill
// transaction.
//
// The id is derived from the dedup key rather than minted, so a re-observation
// of the same restatement collides on the primary key and the UNIQUE alike, and
// `ON CONFLICT DO NOTHING` absorbs it. That is what makes a crash between
// observing and committing safe: the replay inserts nothing new instead of
// duplicating an audit record (D9 — "crash 재관측의 이중 삽입 방지").
//
// Unobserved values are written as the empty string and never as NULL. SQLite
// treats NULLs in an index as distinct, so a NULL here would disable the
// double-insert protection exactly when the broker reported no average price
// (issues.md, Manager 판정 (4)).
func recordExecutionCorrection(ctx context.Context, tx *sql.Tx, c ExecutionCorrection) error {
	accountRef := c.AccountRef
	if accountRef == "" {
		derived, err := accountRefForOrder(ctx, tx, c.OrderID)
		if err != nil {
			return err
		}
		// Still possibly "": an order no local intent claims is an external one,
		// and inventing an account for it would be worse than recording that we
		// could not attribute it. The column is NOT NULL, not non-empty.
		accountRef = derived
	}
	if _, err := tx.ExecContext(ctx,
		`INSERT INTO execution_corrections
		   (id, account_ref, order_id, prev_avg_price, new_avg_price,
		    prev_filled_amount, new_filled_amount, cumulative_qty, observed_at)
		 VALUES (?,?,?,?,?,?,?,?,?)
		 ON CONFLICT DO NOTHING`,
		correctionID(c), accountRef, c.OrderID, c.PrevAveragePrice, c.NewAveragePrice,
		c.PrevFilledAmount, c.NewFilledAmount, c.CumulativeQuantity, c.ObservedAt); err != nil {
		return fmt.Errorf("journal: recording the execution correction of %s: %w", c.OrderID, err)
	}
	return nil
}

// correctionID hashes exactly D9's UNIQUE key. Nothing else may enter it: adding
// a timestamp or the previous values would mint a fresh id for a replay and
// leave the UNIQUE as the only defence.
func correctionID(c ExecutionCorrection) string {
	h := sha256.New()
	for _, part := range []string{c.OrderID, c.CumulativeQuantity, c.NewAveragePrice, c.NewFilledAmount} {
		// The length prefix keeps ("ab","c") and ("a","bc") different keys.
		fmt.Fprintf(h, "%d:%s|", len(part), part)
	}
	return "corr-" + hex.EncodeToString(h.Sum(nil))[:32]
}

// accountRefForOrder finds the account an order belongs to through the attempt
// that named it. A detector Snapshot has no account dimension, and the journal
// does: the intent behind the confirmed attempt is the authority.
func accountRefForOrder(ctx context.Context, tx *sql.Tx, orderID string) (string, error) {
	var ref string
	err := tx.QueryRowContext(ctx, `
		SELECT i.account_ref
		  FROM mutation_attempts a
		  JOIN intents i ON i.id = a.intent_id
		 WHERE a.broker_order_id = ?
		 ORDER BY a.recorded_at DESC, a.rowid DESC
		 LIMIT 1`, orderID).Scan(&ref)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("journal: resolving the account of order %s: %w", orderID, err)
	}
	return ref, nil
}

// ExecutionCorrections returns the recorded restatements of one order, oldest
// first.
func (j *Journal) ExecutionCorrections(ctx context.Context, orderID string) ([]ExecutionCorrection, error) {
	rows, err := j.db.QueryContext(ctx, `
		SELECT id, account_ref, order_id, prev_avg_price, new_avg_price,
		       prev_filled_amount, new_filled_amount, cumulative_qty, observed_at
		  FROM execution_corrections WHERE order_id = ? ORDER BY observed_at, id`, orderID)
	if err != nil {
		return nil, fmt.Errorf("journal: reading the corrections of %s: %w", orderID, err)
	}
	defer rows.Close()

	var out []ExecutionCorrection
	for rows.Next() {
		var (
			c                   ExecutionCorrection
			prevAvg, prevAmount sql.NullString
		)
		if err := rows.Scan(&c.ID, &c.AccountRef, &c.OrderID, &prevAvg, &c.NewAveragePrice,
			&prevAmount, &c.NewFilledAmount, &c.CumulativeQuantity, &c.ObservedAt); err != nil {
			return nil, fmt.Errorf("journal: reading the corrections of %s: %w", orderID, err)
		}
		c.PrevAveragePrice = prevAvg.String
		c.PrevFilledAmount = prevAmount.String
		out = append(out, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("journal: reading the corrections of %s: %w", orderID, err)
	}
	return out, nil
}

// FilledQuantities returns the cumulative filled quantity per symbol, for the
// reconciliation comparison.
func (j *Journal) FilledQuantities(ctx context.Context) (map[string]string, error) {
	rows, err := j.db.QueryContext(ctx,
		`SELECT symbol, filled_quantity FROM fill_snapshots WHERE fail_closed = 0`)
	if err != nil {
		return nil, fmt.Errorf("journal: reading filled quantities: %w", err)
	}
	defer rows.Close()

	totals := map[string]float64{}
	for rows.Next() {
		var symbol, filled string
		if err := rows.Scan(&symbol, &filled); err != nil {
			return nil, fmt.Errorf("journal: reading filled quantities: %w", err)
		}
		v, err := strconv.ParseFloat(strings.TrimSpace(orZero(filled)), 64)
		if err != nil {
			continue
		}
		totals[strings.ToUpper(strings.TrimSpace(symbol))] += v
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("journal: reading filled quantities: %w", err)
	}
	out := make(map[string]string, len(totals))
	for symbol, v := range totals {
		out[symbol] = decimalString(v)
	}
	return out, nil
}

// NetPositions returns the engine's net quantity per symbol, as decimal strings.
//
// Net, not gross. FilledQuantities above answers "how much did we trade", which
// is a volume question; a position question needs the side, because a sell fill
// reduces exposure. Comparing gross fills against an account's holdings would
// report a mismatch on every completed round trip.
//
// The side comes from the intent, through the confirmed attempt that named the
// broker order. An order with no local intent contributes nothing: that is the
// definition of an external order, and the reconciliation classifies it as such
// rather than folding it into the engine's own belief.
func (j *Journal) NetPositions(ctx context.Context) (map[string]string, error) {
	rows, err := j.db.QueryContext(ctx, `
		SELECT f.symbol, i.side, f.filled_quantity
		  FROM fill_snapshots f
		  JOIN mutation_attempts a ON a.broker_order_id = f.order_id AND a.state = ?
		  JOIN intents i ON i.id = a.intent_id
		 WHERE f.fail_closed = 0`, string(StateConfirmed))
	if err != nil {
		return nil, fmt.Errorf("journal: reading net positions: %w", err)
	}
	defer rows.Close()

	totals := map[string]float64{}
	for rows.Next() {
		var symbol, side, filled string
		if err := rows.Scan(&symbol, &side, &filled); err != nil {
			return nil, fmt.Errorf("journal: reading net positions: %w", err)
		}
		v, perr := strconv.ParseFloat(strings.TrimSpace(orZero(filled)), 64)
		if perr != nil {
			continue
		}
		if strings.EqualFold(strings.TrimSpace(side), "SELL") {
			v = -v
		}
		totals[strings.ToUpper(strings.TrimSpace(symbol))] += v
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("journal: reading net positions: %w", err)
	}

	out := make(map[string]string, len(totals))
	for symbol, v := range totals {
		out[symbol] = decimalString(v)
	}
	return out, nil
}

// TrackedFillOrders returns every order the detector must keep reading: the
// snapshots that have not reached a terminal state, plus the orders confirmed
// attempts named that have not been observed yet.
//
// The second half matters more than it looks: an order that is acknowledged and
// then fills entirely before the first poll would otherwise never appear in the
// open list *or* in the snapshot table, and would be tracked by nothing.
func (j *Journal) TrackedFillOrders(ctx context.Context) ([]TrackedFillOrder, error) {
	rows, err := j.db.QueryContext(ctx, `
		SELECT order_id, symbol, market FROM fill_snapshots WHERE terminal = 0
		UNION
		SELECT a.broker_order_id, i.symbol, i.market
		  FROM mutation_attempts a
		  JOIN intents i ON i.id = a.intent_id
		 WHERE a.state = ? AND a.broker_order_id <> ''
		   AND a.broker_order_id NOT IN (SELECT order_id FROM fill_snapshots WHERE terminal = 1)
		ORDER BY 1`, string(StateConfirmed))
	if err != nil {
		return nil, fmt.Errorf("journal: listing tracked fill orders: %w", err)
	}
	defer rows.Close()

	var out []TrackedFillOrder
	for rows.Next() {
		var t TrackedFillOrder
		if err := rows.Scan(&t.OrderID, &t.Symbol, &t.Market); err != nil {
			return nil, fmt.Errorf("journal: listing tracked fill orders: %w", err)
		}
		out = append(out, t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("journal: listing tracked fill orders: %w", err)
	}

	// Lineage is a second read rather than a join: LineageChildren is already the
	// tested reader of that table, and the tracked set is small (one entry per
	// live order) so the extra queries are not a budget concern.
	for i := range out {
		children, err := j.LineageChildren(ctx, out[i].OrderID)
		if err != nil {
			return nil, err
		}
		for _, edge := range children {
			if edge.Relation == RelationReplaces {
				out[i].SuccessorOrderID = edge.ChildOrderID
				break
			}
		}
	}
	return out, nil
}

// LiveOrder is one order the engine believes is still working at the broker,
// carrying the terms a cancel has to name.
//
// The terms come from the intent rather than from the broker: a cancel names an
// order id, and everything else on the request is what the journal records about
// what was cancelled. After an amendment the id is the successor's and the terms
// are the original's — the engine has no amend path today, and the mismatch is
// recorded here rather than papered over.
type LiveOrder struct {
	// OrderID is the lineage-resolved current order number.
	OrderID  string
	IntentID string
	Market   string
	Symbol   string
	Side     string
	Quantity string
	Price    string
	Currency string
}

// LiveOrdersForSymbol lists the working orders of one symbol on one account.
//
// It exists for the exit observation loop's `CancelPendingFirst` (task 7.4): a
// baseline breach liquidates the whole position, and submitting that sell while
// a buy is still working on the same symbol is how the projection ends up
// refusing an ENTRY_WHILE_CLOSING transition into RECONCILE (issues.md, task
// 6.1). The entry has to be cancelled first, and this is how the loop finds it.
//
// "Working" is the same definition TrackedFillOrders uses — a confirmed attempt
// whose order has not reached a terminal fill snapshot — narrowed to one venue
// and symbol. The lineage resolution is the same too: the official API answers a
// modify with a new order number, and cancelling the superseded one cancels
// nothing.
func (j *Journal) LiveOrdersForSymbol(ctx context.Context, accountRef, market, symbol string) ([]LiveOrder, error) {
	rows, err := j.db.QueryContext(ctx, `
		SELECT a.broker_order_id, i.id, i.market, i.symbol, i.side,
		       i.quantity, coalesce(i.price, ''), i.currency
		  FROM mutation_attempts a
		  JOIN intents i ON i.id = a.intent_id
		 WHERE a.state = ? AND a.broker_order_id <> ''
		   AND i.account_ref = ? AND i.market = ? AND i.symbol = ?
		   AND a.broker_order_id NOT IN (SELECT order_id FROM fill_snapshots WHERE terminal = 1)
		 ORDER BY a.recorded_at, a.rowid`,
		string(StateConfirmed), strings.TrimSpace(accountRef),
		normaliseMarket(market), normaliseSymbol(symbol))
	if err != nil {
		return nil, fmt.Errorf("journal: listing the working orders of %s: %w", symbol, err)
	}
	defer rows.Close()

	var out []LiveOrder
	for rows.Next() {
		var o LiveOrder
		if err := rows.Scan(&o.OrderID, &o.IntentID, &o.Market, &o.Symbol, &o.Side,
			&o.Quantity, &o.Price, &o.Currency); err != nil {
			return nil, fmt.Errorf("journal: listing the working orders of %s: %w", symbol, err)
		}
		out = append(out, o)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("journal: listing the working orders of %s: %w", symbol, err)
	}

	for i := range out {
		current, err := j.ResolveCurrentOrderID(ctx, out[i].OrderID)
		if err != nil {
			return nil, err
		}
		out[i].OrderID = current
	}
	return out, nil
}

// nearlyZero reports whether a delta is zero within a scale-relative tolerance.
// Quantities arrive as decimal strings and are compared as float64, so an exact
// == would make a 0.1+0.2 fractional-share fill look like a discrepancy.
func nearlyZero(delta, scale float64) bool {
	return math.Abs(delta) <= 1e-9*math.Max(1, math.Abs(scale))
}

func decimalString(v float64) string { return strconv.FormatFloat(v, 'f', -1, 64) }

func orZero(s string) string {
	if strings.TrimSpace(s) == "" {
		return "0"
	}
	return strings.TrimSpace(s)
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// RFC3339 renders an instant the way this table stores timestamps, so a caller
// building a FillObservation cannot get the format subtly wrong. Empty in, empty
// out: "the broker gave no timestamp" is a fact worth preserving.
func RFC3339(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format(time.RFC3339)
}

// schemaV2 adds the fill ledger.
//
// Additive per the migration rules in schema.go: two new tables and their
// indexes, no change to any released column, no rewrite of any historical row.
// Rollback is "run the previous binary", which refuses the newer user_version and
// stops rather than misreading it (ErrSchemaTooNew) — the safe direction for a
// live account, and the reason the version guard exists.
const schemaV2 = `
-- The last cumulative fill observation per broker order (lineage node). This is
-- durable state, not a cache: without it, the first poll after a restart would
-- report the whole cumulative quantity as new.
CREATE TABLE fill_snapshots (
	order_id          TEXT PRIMARY KEY,
	symbol            TEXT NOT NULL,
	market            TEXT NOT NULL DEFAULT '',
	state             TEXT NOT NULL DEFAULT '',   -- internal/brokerstate.State
	terminal          INTEGER NOT NULL DEFAULT 0,
	fail_closed       INTEGER NOT NULL DEFAULT 0,
	quantity          TEXT NOT NULL DEFAULT '0',  -- decimal string
	filled_quantity   TEXT NOT NULL DEFAULT '0',  -- cumulative decimal string
	average_price     TEXT NOT NULL DEFAULT '',   -- decimal string, replaced not accumulated
	broker_visible_at TEXT NOT NULL DEFAULT '',   -- execution.filledAt, RFC3339 UTC
	observed_at       TEXT NOT NULL DEFAULT '',
	committed_at      TEXT NOT NULL,
	reason_code       TEXT NOT NULL DEFAULT '',
	detail            TEXT NOT NULL DEFAULT ''
) STRICT;

CREATE INDEX idx_fill_snapshots_symbol   ON fill_snapshots(symbol);
CREATE INDEX idx_fill_snapshots_terminal ON fill_snapshots(terminal);

-- Append-only log of positive deltas. Rows are never updated or deleted: this is
-- the fill history the Phase 2 ledger imports.
CREATE TABLE fill_events (
	id                  INTEGER PRIMARY KEY AUTOINCREMENT,
	order_id            TEXT NOT NULL,
	symbol              TEXT NOT NULL,
	market              TEXT NOT NULL DEFAULT '',
	delta_quantity      TEXT NOT NULL,            -- decimal string, always > 0
	cumulative_quantity TEXT NOT NULL,
	average_price       TEXT NOT NULL DEFAULT '',
	broker_visible_at   TEXT NOT NULL DEFAULT '',
	committed_at        TEXT NOT NULL
) STRICT;

CREATE INDEX idx_fill_events_order  ON fill_events(order_id, id);
CREATE INDEX idx_fill_events_symbol ON fill_events(symbol, id);
`

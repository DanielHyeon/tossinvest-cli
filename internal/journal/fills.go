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

// ErrFillScopeAmbiguous means an order id names more than one canonical
// snapshot. Callers must use LookupFillScoped instead of guessing which account
// session the broker's opaque identifier refers to.
var ErrFillScopeAmbiguous = errors.New("journal: fill snapshot order id is ambiguous across scopes")

// ErrTrackedFillIdentityConflict means a confirmed local order cannot be bound
// to one account/market/symbol identity. It is returned only after the journal
// has durably entered an account-wide IDENTIFIER_CONFLICT state.
var ErrTrackedFillIdentityConflict = errors.New("journal: tracked fill order identity conflict")

// FillObservation is one cumulative snapshot offered to the ledger.
//
// Quantities and the average price are decimal strings for the same reason the
// intent's are: this is an audit record, and a broker's decimal quantity has no
// exact binary float form.
type FillObservation struct {
	OrderID string
	Symbol  string
	Market  string
	// TradingDay and Side are broker evidence used with account/market to scope
	// order ids that may be reused by another session.
	TradingDay string
	Side       string
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
	AccountRef      string
	Symbol          string
	Market          string
	TradingDay      string
	Side            string
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

// FillSnapshotScope is the canonical identity of one cumulative broker-order
// sequence. Broker order ids are opaque and may be reused, so OrderID alone is
// deliberately not sufficient for a scoped lookup.
type FillSnapshotScope struct {
	OrderID    string
	AccountRef string
	Market     string
	TradingDay string
	Symbol     string
	Side       string
}

func canonicalFillSnapshotScope(scope FillSnapshotScope) FillSnapshotScope {
	return FillSnapshotScope{
		OrderID:    scope.OrderID,
		AccountRef: strings.TrimSpace(scope.AccountRef),
		Market:     normaliseMarket(scope.Market),
		TradingDay: strings.TrimSpace(scope.TradingDay),
		Symbol:     normaliseSymbol(scope.Symbol),
		Side:       strings.ToUpper(strings.TrimSpace(scope.Side)),
	}
}

func fillSnapshotScopeOf(obs FillObservation) FillSnapshotScope {
	return canonicalFillSnapshotScope(FillSnapshotScope{
		OrderID: obs.OrderID, AccountRef: obs.AccountRef, Market: obs.Market,
		TradingDay: obs.TradingDay, Symbol: obs.Symbol, Side: obs.Side,
	})
}

func (scope FillSnapshotScope) complete() bool {
	return scope.OrderID != "" && scope.AccountRef != "" && scope.Market != "" &&
		scope.TradingDay != "" && scope.Symbol != "" && scope.Side != ""
}

// legacyUnscoped is the only incomplete identity accepted for a new journal
// observation. Schema-v15 snapshots carried symbol and market, but had no
// account, trading-day, or side columns. A partially populated v16 identity is
// neither that legacy shape nor a canonical key and cannot be made idempotent.
func (scope FillSnapshotScope) legacyUnscoped() bool {
	return scope.OrderID != "" && scope.AccountRef == "" && scope.TradingDay == "" && scope.Side == ""
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
	OrderID    string
	IntentID   string
	AccountRef string
	Symbol     string
	Market     string
	TradingDay string
	Side       string
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
	scope := fillSnapshotScopeOf(obs)
	if !scope.complete() && !scope.legacyUnscoped() {
		return FillResult{}, fmt.Errorf(
			"%w: fill snapshot %q has a partial canonical scope; account, market, trading day, symbol, and side must be complete together",
			ErrInvalidRequest, obs.OrderID)
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

	prev, err := lookupFillSnapshotScoped(ctx, tx, scope)
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
		if err := markFillRefused(ctx, tx, scope, obs, refusal, now, hadPrev); err != nil {
			return res, err
		}
		// Nothing is released on a refusal — that is the point. A snapshot the
		// derivation could not explain (CLOSED, nothing filled, no cancellation:
		// what an expiry would look like, [미측정 — 2b 2.1]) leaves the hold
		// standing, and the operator is told which holds it is standing on
		// (design D5: 만료 추정으로 해제하지 않는다).
		alerts, err := alertsForOrder(ctx, tx, scope, ReasonFillStateUnknown, refusal.detail)
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
	if err := upsertFillSnapshot(ctx, tx, scope, obs, now); err != nil {
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
	applied := AppliedFill{
		OrderID:                orderID,
		AccountRef:             obs.AccountRef,
		Symbol:                 obs.Symbol,
		Market:                 obs.Market,
		TradingDay:             obs.TradingDay,
		Side:                   obs.Side,
		State:                  obs.State,
		Terminal:               obs.Terminal,
		Delta:                  res.Delta,
		CumulativeQuantity:     orZero(obs.FilledQuantity),
		AveragePrice:           obs.AveragePrice,
		FilledAmount:           obs.FilledAmount,
		OrderedQuantity:        obs.Quantity,
		PrevCumulativeQuantity: prev.FilledQuantity,
		PrevAveragePrice:       prev.AveragePrice,
		Corrected:              res.Corrected,
		BrokerVisibleAt:        obs.BrokerVisibleAt,
		CommittedAt:            now,
	}
	// Resolve ownership before any local reservation or domain state can move.
	// A broker-only order remains an observation, while a colliding confirmed id
	// durably raises IDENTIFIER_CONFLICT inside this same transaction.
	ownershipHandle := &ApplyTx{tx: tx, now: now}
	ownedOrigin, locallyOwned, err := resolveFillOrigin(ctx, ownershipHandle, applied)
	ownershipHandle.invalidate()
	if err != nil {
		return res, err
	}

	// 5. A *derived* terminal state frees the decision's holds, in this same
	//    transaction (design D5). The verdict is the caller's derivation
	//    (internal/brokerstate), never a guess made here: this branch is
	//    unreachable for a fail-closed observation, which returned above.
	if obs.Terminal && locallyOwned {
		released, err := releaseReservationsForOrder(ctx, tx, orderID, ownedOrigin.IntentID,
			ReleaseReasonBrokerTerminal,
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
	if locallyOwned {
		if err := j.runApplyHooks(ctx, tx, applied); err != nil {
			return res, err
		}
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
func markFillRefused(ctx context.Context, tx *sql.Tx, scope FillSnapshotScope, obs FillObservation,
	refusal *fillRefusal, now string, hadPrev bool) error {
	orderID := scope.OrderID
	if hadPrev {
		if err := updateFillSnapshotRefusal(ctx, tx, scope, refusal, obs.ObservedAt, now); err != nil {
			return fmt.Errorf("journal: recording the refused snapshot of %s: %w", orderID, err)
		}
		return nil
	}
	refused := obs
	refused.Terminal = false
	refused.FilledQuantity = "0"
	refused.AveragePrice = ""
	refused.FilledAmount = ""
	if err := insertRefusedFillSnapshot(ctx, tx, scope, refused, refusal, now); err != nil {
		return fmt.Errorf("journal: recording the refused snapshot of %s: %w", orderID, err)
	}
	return nil
}

func upsertFillSnapshot(ctx context.Context, tx *sql.Tx, scope FillSnapshotScope,
	obs FillObservation, now string) error {
	if scope.complete() {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO scoped_fill_snapshots
			  (order_id, account_ref, market, trading_day, symbol, side,
			   state, terminal, fail_closed, quantity, filled_quantity,
			   average_price, filled_amount, broker_visible_at, observed_at, committed_at,
			   reason_code, detail)
			VALUES (?,?,?,?,?,?,?, ?,0,?,?,?,?,?,?,?,'','')
			ON CONFLICT(account_ref, market, trading_day, symbol, side, order_id) DO UPDATE SET
			  state = excluded.state, terminal = excluded.terminal, fail_closed = 0,
			  quantity = excluded.quantity, filled_quantity = excluded.filled_quantity,
			  average_price = excluded.average_price, filled_amount = excluded.filled_amount,
			  broker_visible_at = excluded.broker_visible_at,
			  observed_at = excluded.observed_at, committed_at = excluded.committed_at,
			  reason_code = '', detail = ''`,
			scope.OrderID, scope.AccountRef, scope.Market, scope.TradingDay, scope.Symbol, scope.Side,
			obs.State, boolToInt(obs.Terminal), orZero(obs.Quantity), orZero(obs.FilledQuantity),
			obs.AveragePrice, obs.FilledAmount, obs.BrokerVisibleAt, obs.ObservedAt, now); err != nil {
			return err
		}
		return mirrorScopedFillSnapshot(ctx, tx, scope, obs, now, false, "", "")
	}
	_, err := tx.ExecContext(ctx, `
		INSERT INTO fill_snapshots
		  (order_id, account_ref, symbol, market, trading_day, side,
		   state, terminal, fail_closed, quantity, filled_quantity,
		   average_price, filled_amount, broker_visible_at, observed_at, committed_at,
		   reason_code, detail)
		VALUES (?,?,?,?,?,?,?,?,0,?,?,?,?,?,?,?,'','')
		ON CONFLICT(order_id) DO UPDATE SET
		  account_ref = excluded.account_ref, symbol = excluded.symbol, market = excluded.market,
		  trading_day = excluded.trading_day, side = excluded.side, state = excluded.state,
		  terminal = excluded.terminal, fail_closed = 0, quantity = excluded.quantity,
		  filled_quantity = excluded.filled_quantity, average_price = excluded.average_price,
		  filled_amount = excluded.filled_amount,
		  broker_visible_at = excluded.broker_visible_at, observed_at = excluded.observed_at,
		  committed_at = excluded.committed_at, reason_code = '', detail = ''`,
		scope.OrderID, scope.AccountRef, scope.Symbol, scope.Market, scope.TradingDay, scope.Side,
		obs.State, boolToInt(obs.Terminal), orZero(obs.Quantity), orZero(obs.FilledQuantity),
		obs.AveragePrice, obs.FilledAmount, obs.BrokerVisibleAt, obs.ObservedAt, now)
	return err
}

// mirrorScopedFillSnapshot preserves the released order-id keyed compatibility
// row when it is absent or already names this exact scope. It never overwrites a
// different scope; the complete durable history lives in scoped_fill_snapshots.
func mirrorScopedFillSnapshot(ctx context.Context, tx *sql.Tx, scope FillSnapshotScope,
	obs FillObservation, now string, failClosed bool, reason, detail string) error {
	_, err := tx.ExecContext(ctx, `
		INSERT INTO fill_snapshots
		  (order_id, account_ref, symbol, market, trading_day, side,
		   state, terminal, fail_closed, quantity, filled_quantity,
		   average_price, filled_amount, broker_visible_at, observed_at, committed_at,
		   reason_code, detail)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)
		ON CONFLICT(order_id) DO UPDATE SET
		  state = excluded.state, terminal = excluded.terminal,
		  fail_closed = excluded.fail_closed, quantity = excluded.quantity,
		  filled_quantity = excluded.filled_quantity, average_price = excluded.average_price,
		  filled_amount = excluded.filled_amount,
		  broker_visible_at = excluded.broker_visible_at, observed_at = excluded.observed_at,
		  committed_at = excluded.committed_at,
		  reason_code = excluded.reason_code, detail = excluded.detail
		WHERE TRIM(fill_snapshots.account_ref) = excluded.account_ref
		  AND LOWER(TRIM(fill_snapshots.market)) = excluded.market
		  AND TRIM(fill_snapshots.trading_day) = excluded.trading_day
		  AND UPPER(TRIM(fill_snapshots.symbol)) = excluded.symbol
		  AND UPPER(TRIM(fill_snapshots.side)) = excluded.side`,
		scope.OrderID, scope.AccountRef, scope.Symbol, scope.Market, scope.TradingDay, scope.Side,
		obs.State, boolToInt(obs.Terminal), boolToInt(failClosed), orZero(obs.Quantity),
		orZero(obs.FilledQuantity), obs.AveragePrice, obs.FilledAmount, obs.BrokerVisibleAt,
		obs.ObservedAt, now, reason, detail)
	return err
}

func insertRefusedFillSnapshot(ctx context.Context, tx *sql.Tx, scope FillSnapshotScope,
	obs FillObservation, refusal *fillRefusal, now string) error {
	if scope.complete() {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO scoped_fill_snapshots
			  (order_id, account_ref, market, trading_day, symbol, side, state, terminal,
			   fail_closed, quantity, filled_quantity, average_price, filled_amount,
			   broker_visible_at, observed_at, committed_at, reason_code, detail)
			VALUES (?,?,?,?,?,?,?,0,1,?,'0','','','',?,?,?,?)`,
			scope.OrderID, scope.AccountRef, scope.Market, scope.TradingDay, scope.Symbol, scope.Side,
			obs.State, orZero(obs.Quantity), obs.ObservedAt, now, refusal.reason, refusal.detail); err != nil {
			return err
		}
		return mirrorScopedFillSnapshot(ctx, tx, scope, obs, now, true, refusal.reason, refusal.detail)
	}
	_, err := tx.ExecContext(ctx, `
		INSERT INTO fill_snapshots
		  (order_id, account_ref, symbol, market, trading_day, side,
		   state, terminal, fail_closed, quantity, filled_quantity,
		   average_price, broker_visible_at, observed_at, committed_at, reason_code, detail)
		VALUES (?,?,?,?,?,?,?,0,1,?,'0','','',?,?,?,?)`,
		scope.OrderID, scope.AccountRef, scope.Symbol, scope.Market, scope.TradingDay, scope.Side,
		obs.State, orZero(obs.Quantity), obs.ObservedAt, now, refusal.reason, refusal.detail)
	return err
}

func updateFillSnapshotRefusal(ctx context.Context, tx *sql.Tx, scope FillSnapshotScope,
	refusal *fillRefusal, observedAt, now string) error {
	if scope.complete() {
		if _, err := tx.ExecContext(ctx, `
			UPDATE scoped_fill_snapshots
			   SET fail_closed = 1, reason_code = ?, detail = ?, observed_at = ?, committed_at = ?
			 WHERE account_ref = ? AND market = ? AND trading_day = ? AND symbol = ?
			   AND side = ? AND order_id = ?`,
			refusal.reason, refusal.detail, observedAt, now, scope.AccountRef, scope.Market,
			scope.TradingDay, scope.Symbol, scope.Side, scope.OrderID); err != nil {
			return err
		}
		_, err := tx.ExecContext(ctx, `
			UPDATE fill_snapshots
			   SET fail_closed = 1, reason_code = ?, detail = ?, observed_at = ?, committed_at = ?
			 WHERE order_id = ? AND TRIM(account_ref) = ? AND LOWER(TRIM(market)) = ?
			   AND TRIM(trading_day) = ? AND UPPER(TRIM(symbol)) = ? AND UPPER(TRIM(side)) = ?`,
			refusal.reason, refusal.detail, observedAt, now, scope.OrderID, scope.AccountRef,
			scope.Market, scope.TradingDay, scope.Symbol, scope.Side)
		return err
	}
	_, err := tx.ExecContext(ctx, `
		UPDATE fill_snapshots
		   SET fail_closed = 1, reason_code = ?, detail = ?, observed_at = ?, committed_at = ?
		 WHERE order_id = ?`, refusal.reason, refusal.detail, observedAt, now, scope.OrderID)
	return err
}

// sameSnapshot reports whether an observation says exactly what the stored row
// already says about the broker's state. Timestamps of *our* reading are excluded
// on purpose: when we looked does not change what is true.
func sameSnapshot(obs FillObservation, prev FillSnapshotRecord) bool {
	return orZero(obs.FilledQuantity) == orZero(prev.FilledQuantity) &&
		strings.TrimSpace(obs.AccountRef) == strings.TrimSpace(prev.AccountRef) &&
		strings.TrimSpace(obs.TradingDay) == strings.TrimSpace(prev.TradingDay) &&
		strings.EqualFold(strings.TrimSpace(obs.Side), strings.TrimSpace(prev.Side)) &&
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

const fillSnapshotColumns = `order_id, account_ref, symbol, market, trading_day, side,
	state, terminal, fail_closed, quantity, filled_quantity, average_price, filled_amount,
	broker_visible_at, observed_at, committed_at, reason_code, detail`

const fillSelect = `SELECT ` + fillSnapshotColumns + ` FROM fill_snapshots`

const scopedFillSelect = `SELECT ` + fillSnapshotColumns + ` FROM scoped_fill_snapshots`

const allFillSnapshotsCTE = `WITH all_fill_snapshots AS (
	SELECT ` + fillSnapshotColumns + ` FROM scoped_fill_snapshots
	UNION ALL
	SELECT ` + fillSnapshotColumns + ` FROM fill_snapshots legacy
	 WHERE NOT EXISTS (
		SELECT 1 FROM scoped_fill_snapshots scoped
		 WHERE scoped.order_id = legacy.order_id
		   AND scoped.account_ref = TRIM(legacy.account_ref)
		   AND scoped.market = LOWER(TRIM(legacy.market))
		   AND scoped.trading_day = TRIM(legacy.trading_day)
		   AND scoped.symbol = UPPER(TRIM(legacy.symbol))
		   AND scoped.side = UPPER(TRIM(legacy.side))
	)
)`

func scanFillSnapshot(row rowScanner) (FillSnapshotRecord, error) {
	var (
		rec                  FillSnapshotRecord
		terminal, failClosed int
	)
	err := row.Scan(&rec.OrderID, &rec.AccountRef, &rec.Symbol, &rec.Market, &rec.TradingDay,
		&rec.Side, &rec.State, &terminal, &failClosed,
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

type queryRower interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

func lookupFillSnapshotScoped(ctx context.Context, q queryRower,
	scope FillSnapshotScope) (FillSnapshotRecord, error) {
	scope = canonicalFillSnapshotScope(scope)
	if !scope.complete() {
		return scanFillSnapshot(q.QueryRowContext(ctx, fillSelect+`
			WHERE order_id = ? AND TRIM(account_ref) = '' AND TRIM(trading_day) = ''`, scope.OrderID))
	}
	rec, err := scanFillSnapshot(q.QueryRowContext(ctx, scopedFillSelect+`
		WHERE account_ref = ? AND market = ? AND trading_day = ? AND symbol = ? AND side = ?
		  AND order_id = ?`, scope.AccountRef, scope.Market, scope.TradingDay, scope.Symbol,
		scope.Side, scope.OrderID))
	if err == nil || !errors.Is(err, ErrFillNotFound) {
		return rec, err
	}
	// A v16 build may already have written one fully scoped row into the released
	// order-id keyed table. Match every dimension; a v15 blank row is evidence,
	// never a wildcard for this API.
	return scanFillSnapshot(q.QueryRowContext(ctx, fillSelect+`
		WHERE order_id = ? AND TRIM(account_ref) = ? AND LOWER(TRIM(market)) = ?
		  AND TRIM(trading_day) = ? AND UPPER(TRIM(symbol)) = ? AND UPPER(TRIM(side)) = ?`,
		scope.OrderID, scope.AccountRef, scope.Market, scope.TradingDay, scope.Symbol, scope.Side))
}

// LookupFillScoped returns exactly one canonical cumulative sequence. It never
// attributes a v15 blank-scope row to a later scoped order.
func (j *Journal) LookupFillScoped(ctx context.Context, scope FillSnapshotScope) (FillSnapshotRecord, error) {
	return lookupFillSnapshotScoped(ctx, j.db, scope)
}

// LookupFill returns the stored snapshot of one order when the id is globally
// unambiguous. Reused ids fail closed; callers that have canonical identity must
// use LookupFillScoped.
//
// The identifier is used as given: the rows are keyed by what the broker sent,
// so normalising the lookup would miss them.
func (j *Journal) LookupFill(ctx context.Context, orderID string) (FillSnapshotRecord, error) {
	rows, err := j.db.QueryContext(ctx, allFillSnapshotsCTE+`
		SELECT `+fillSnapshotColumns+` FROM all_fill_snapshots WHERE order_id = ? LIMIT 2`, orderID)
	if err != nil {
		return FillSnapshotRecord{}, fmt.Errorf("journal: reading a fill snapshot: %w", err)
	}
	defer rows.Close()
	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return FillSnapshotRecord{}, fmt.Errorf("journal: reading a fill snapshot: %w", err)
		}
		return FillSnapshotRecord{}, ErrFillNotFound
	}
	rec, err := scanFillSnapshot(rows)
	if err != nil {
		return FillSnapshotRecord{}, err
	}
	if rows.Next() {
		return FillSnapshotRecord{}, fmt.Errorf("%w: order %s", ErrFillScopeAmbiguous, orderID)
	}
	if err := rows.Err(); err != nil {
		return FillSnapshotRecord{}, fmt.Errorf("journal: reading a fill snapshot: %w", err)
	}
	return rec, nil
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
		allFillSnapshotsCTE+` SELECT symbol, filled_quantity FROM all_fill_snapshots WHERE fail_closed = 0`)
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
	rows, err := j.db.QueryContext(ctx, allFillSnapshotsCTE+`
		SELECT f.symbol, i.side, f.filled_quantity
		  FROM all_fill_snapshots f
		  JOIN mutation_attempts a ON a.broker_order_id = f.order_id AND a.state = ?
		  JOIN intents i ON i.id = a.intent_id
		 WHERE f.fail_closed = 0
		   AND ((TRIM(f.account_ref) = TRIM(i.account_ref)
		         AND LOWER(TRIM(f.market)) = LOWER(TRIM(i.market))
		         AND TRIM(f.trading_day) = TRIM(i.trading_day)
		         AND UPPER(TRIM(f.symbol)) = UPPER(TRIM(i.symbol))
		         AND UPPER(TRIM(f.side)) = UPPER(TRIM(i.side)))
		        OR (TRIM(f.account_ref) = '' AND TRIM(f.trading_day) = ''
		            AND UPPER(TRIM(f.market)) = UPPER(TRIM(i.market))
		            AND UPPER(TRIM(f.symbol)) = UPPER(TRIM(i.symbol))
		            AND (TRIM(f.side) = '' OR UPPER(TRIM(f.side)) = UPPER(TRIM(i.side)))
		            AND NOT EXISTS (
		              SELECT 1
		                FROM mutation_attempts reused_attempt
		                JOIN intents reused_intent ON reused_intent.id = reused_attempt.intent_id
		               WHERE reused_attempt.state = ?
		                 AND reused_attempt.broker_order_id = f.order_id
		                 AND reused_intent.account_ref = i.account_ref
		                 AND UPPER(TRIM(reused_intent.market)) = UPPER(TRIM(i.market))
		                 AND TRIM(reused_intent.trading_day) <> TRIM(i.trading_day))))`,
		string(StateConfirmed), string(StateConfirmed))
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

// TrackedFillOrders returns every locally owned order the detector must keep
// reading: non-terminal snapshots backed by a confirmed attempt or replacement
// lineage, plus the orders confirmed attempts named that have not been observed
// yet. A broker snapshot alone is an external observation, not ownership proof.
//
// The second half matters more than it looks: an order that is acknowledged and
// then fills entirely before the first poll would otherwise never appear in the
// open list *or* in the snapshot table, and would be tracked by nothing.
func (j *Journal) TrackedFillOrders(ctx context.Context, accountRefs ...string) ([]TrackedFillOrder, error) {
	accountRef := ""
	if len(accountRefs) > 0 {
		accountRef = strings.TrimSpace(accountRefs[0])
	}
	if accountRef != "" {
		if err := j.guardTrackedFillIdentity(ctx, accountRef); err != nil {
			return nil, err
		}
	}
	rows, err := j.db.QueryContext(ctx, allFillSnapshotsCTE+`,
		all_confirmed_orders AS (
			SELECT a.broker_order_id AS order_id, i.id AS intent_id, i.account_ref,
			       i.symbol, i.market, i.trading_day, i.side
			  FROM mutation_attempts a
			  JOIN intents i ON i.id = a.intent_id
			 WHERE a.state = ? AND a.broker_order_id <> ''
		), confirmed_orders AS (
			SELECT c.order_id, c.intent_id, c.account_ref, c.symbol, c.market, c.trading_day, c.side
			  FROM all_confirmed_orders c
			 WHERE (? = '' OR c.account_ref = ?)
		), valid_lineage AS (
			SELECT s.parent_order_id, s.child_order_id, s.intent_id, s.attempt_id,
			       s.account_ref, s.symbol, s.market, s.trading_day, s.side
			  FROM scoped_lineage_edges s
			  JOIN mutation_attempts a ON a.id = s.attempt_id AND a.intent_id = s.intent_id
			  JOIN intents i ON i.id = s.intent_id
			 WHERE s.relation = ? AND a.kind = ? AND a.state = ?
			   AND a.target_order_id = s.parent_order_id
			   AND a.broker_order_id = s.child_order_id
			   AND TRIM(a.account_ref) = TRIM(i.account_ref)
			   AND s.account_ref = TRIM(i.account_ref)
			   AND s.market = LOWER(TRIM(i.market))
			   AND s.trading_day = TRIM(i.trading_day)
			   AND s.symbol = UPPER(TRIM(i.symbol))
			   AND s.side = UPPER(TRIM(i.side))
			UNION
			SELECT l.parent_order_id, l.child_order_id, l.intent_id, l.attempt_id,
			       TRIM(i.account_ref), UPPER(TRIM(i.symbol)), LOWER(TRIM(i.market)),
			       TRIM(i.trading_day), UPPER(TRIM(i.side))
			  FROM lineage_edges l
			  JOIN mutation_attempts a ON a.id = l.attempt_id AND a.intent_id = l.intent_id
			  JOIN intents i ON i.id = l.intent_id
			 WHERE l.relation = ? AND a.kind = ? AND a.state = ?
			   AND a.target_order_id = l.parent_order_id
			   AND a.broker_order_id = l.child_order_id
			   AND TRIM(a.account_ref) = TRIM(i.account_ref)
		)
		SELECT f.order_id, c.intent_id, c.account_ref, c.symbol, c.market, c.trading_day, c.side
		  FROM all_fill_snapshots f
		  JOIN confirmed_orders c ON c.order_id = f.order_id
		 WHERE f.terminal = 0
		   AND ((TRIM(f.account_ref) = TRIM(c.account_ref)
		         AND TRIM(f.trading_day) = TRIM(c.trading_day))
		        OR (TRIM(f.account_ref) = '' AND TRIM(f.trading_day) = ''
		            AND NOT EXISTS (
		              SELECT 1 FROM all_confirmed_orders reused
		               WHERE reused.order_id = c.order_id
		                 AND reused.account_ref = c.account_ref
		                 AND UPPER(TRIM(reused.market)) = UPPER(TRIM(c.market))
		                 AND TRIM(reused.trading_day) <> TRIM(c.trading_day))))
		   AND UPPER(TRIM(f.symbol)) = UPPER(TRIM(c.symbol))
		   AND UPPER(TRIM(f.market)) = UPPER(TRIM(c.market))
		   AND (TRIM(f.side) = '' OR UPPER(TRIM(f.side)) = UPPER(TRIM(c.side)))
		UNION
		SELECT f.order_id, l.intent_id, l.account_ref, l.symbol, l.market, l.trading_day, l.side
		  FROM all_fill_snapshots f
		  JOIN valid_lineage l
		    ON l.parent_order_id = f.order_id OR l.child_order_id = f.order_id
		  JOIN confirmed_orders c
		    ON (c.order_id = l.parent_order_id OR c.order_id = l.child_order_id)
		   AND c.account_ref = l.account_ref
		   AND UPPER(TRIM(c.market)) = UPPER(TRIM(l.market))
		   AND TRIM(c.trading_day) = TRIM(l.trading_day)
		   AND UPPER(TRIM(c.symbol)) = UPPER(TRIM(l.symbol))
		   AND UPPER(TRIM(c.side)) = UPPER(TRIM(l.side))
		 WHERE f.terminal = 0
		   AND ((TRIM(f.account_ref) = TRIM(l.account_ref)
		         AND TRIM(f.trading_day) = TRIM(l.trading_day))
		        OR (TRIM(f.account_ref) = '' AND TRIM(f.trading_day) = ''
		            AND NOT EXISTS (
		              SELECT 1 FROM all_confirmed_orders reused
		               WHERE reused.account_ref = l.account_ref
		                 AND UPPER(TRIM(reused.market)) = UPPER(TRIM(l.market))
		                 AND TRIM(reused.trading_day) <> TRIM(l.trading_day)
		                 AND (reused.order_id = l.parent_order_id
		                      OR reused.order_id = l.child_order_id))))
		   AND UPPER(TRIM(f.symbol)) = UPPER(TRIM(l.symbol))
		   AND UPPER(TRIM(f.market)) = UPPER(TRIM(l.market))
		   AND (TRIM(f.side) = '' OR UPPER(TRIM(f.side)) = UPPER(TRIM(l.side)))
		UNION
		SELECT c.order_id, c.intent_id, c.account_ref, c.symbol, c.market, c.trading_day, c.side
		  FROM confirmed_orders c
		 WHERE NOT EXISTS (
			SELECT 1 FROM all_fill_snapshots f
			 WHERE f.order_id = c.order_id
			   AND ((TRIM(f.account_ref) = TRIM(c.account_ref)
			         AND TRIM(f.trading_day) = TRIM(c.trading_day))
			        OR (TRIM(f.account_ref) = '' AND TRIM(f.trading_day) = ''
			            AND NOT EXISTS (
			              SELECT 1 FROM all_confirmed_orders reused
			               WHERE reused.order_id = c.order_id
			                 AND reused.account_ref = c.account_ref
			                 AND UPPER(TRIM(reused.market)) = UPPER(TRIM(c.market))
			                 AND TRIM(reused.trading_day) <> TRIM(c.trading_day))))
		 )
		ORDER BY 1`, string(StateConfirmed), accountRef, accountRef,
		RelationReplaces, string(KindAmend), string(StateConfirmed),
		RelationReplaces, string(KindAmend), string(StateConfirmed))
	if err != nil {
		return nil, fmt.Errorf("journal: listing tracked fill orders: %w", err)
	}
	defer rows.Close()

	var out []TrackedFillOrder
	for rows.Next() {
		var t TrackedFillOrder
		if err := rows.Scan(&t.OrderID, &t.IntentID, &t.AccountRef, &t.Symbol, &t.Market,
			&t.TradingDay, &t.Side); err != nil {
			return nil, fmt.Errorf("journal: listing tracked fill orders: %w", err)
		}
		out = append(out, t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("journal: listing tracked fill orders: %w", err)
	}

	// Lineage is a second read rather than a join. The tracked set is small (one
	// entry per live order), and carrying the canonical ownership scope into the
	// resolver is what prevents a reused broker id from following another
	// account's or trading day's replacement chain.
	for i := range out {
		current, err := j.ResolveCurrentOrderIDScoped(ctx, out[i].OrderID, OrderLineageScope{
			AccountRef: out[i].AccountRef,
			Market:     out[i].Market,
			TradingDay: out[i].TradingDay,
			Symbol:     out[i].Symbol,
			Side:       out[i].Side,
		})
		if err != nil {
			return nil, err
		}
		if current != out[i].OrderID {
			out[i].SuccessorOrderID = current
		}
	}
	return out, nil
}

func (j *Journal) guardTrackedFillIdentity(ctx context.Context, accountRef string) error {
	var orderID, symbol, market, detail string
	err := j.db.QueryRowContext(ctx, allFillSnapshotsCTE+`,
		all_confirmed_orders AS (
			SELECT a.broker_order_id AS order_id, i.id AS intent_id, i.account_ref,
			       i.symbol, i.market, i.trading_day, i.side
			  FROM mutation_attempts a
			  JOIN intents i ON i.id = a.intent_id
			 WHERE a.state = ? AND a.broker_order_id <> ''
		), valid_lineage AS (
			SELECT s.parent_order_id, s.child_order_id, s.intent_id,
			       s.account_ref, s.symbol, s.market, s.trading_day, s.side
			  FROM scoped_lineage_edges s
			  JOIN mutation_attempts a ON a.id = s.attempt_id AND a.intent_id = s.intent_id
			  JOIN intents i ON i.id = s.intent_id
			 WHERE s.relation = ? AND a.kind = ? AND a.state = ?
			   AND a.target_order_id = s.parent_order_id
			   AND a.broker_order_id = s.child_order_id
			   AND TRIM(a.account_ref) = TRIM(i.account_ref)
			   AND s.account_ref = TRIM(i.account_ref)
			   AND s.market = LOWER(TRIM(i.market))
			   AND s.trading_day = TRIM(i.trading_day)
			   AND s.symbol = UPPER(TRIM(i.symbol))
			   AND s.side = UPPER(TRIM(i.side))
			UNION
			SELECT l.parent_order_id, l.child_order_id, l.intent_id,
			       TRIM(i.account_ref), UPPER(TRIM(i.symbol)), LOWER(TRIM(i.market)),
			       TRIM(i.trading_day), UPPER(TRIM(i.side))
			  FROM lineage_edges l
			  JOIN mutation_attempts a ON a.id = l.attempt_id AND a.intent_id = l.intent_id
			  JOIN intents i ON i.id = l.intent_id
			 WHERE l.relation = ? AND a.kind = ? AND a.state = ?
			   AND a.target_order_id = l.parent_order_id
			   AND a.broker_order_id = l.child_order_id
			   AND TRIM(a.account_ref) = TRIM(i.account_ref)
		)
		SELECT c.order_id, c.symbol, c.market, 'order id is confirmed for more than one intent in the same canonical scope'
		  FROM all_confirmed_orders c
		 WHERE c.account_ref = ?
			   AND EXISTS (SELECT 1 FROM all_confirmed_orders other
			                WHERE other.order_id = c.order_id
			                  AND other.account_ref = c.account_ref
			                  AND UPPER(TRIM(other.market)) = UPPER(TRIM(c.market))
			                  AND TRIM(other.trading_day) = TRIM(c.trading_day)
			                  AND UPPER(TRIM(other.symbol)) = UPPER(TRIM(c.symbol))
			                  AND UPPER(TRIM(other.side)) = UPPER(TRIM(c.side))
			                  AND other.intent_id <> c.intent_id)
		UNION
		SELECT f.order_id, c.symbol, c.market, 'broker snapshot symbol or market contradicts the confirmed intent'
		  FROM all_confirmed_orders c
		  JOIN all_fill_snapshots f ON f.order_id = c.order_id
		 WHERE c.account_ref = ?
		   AND ((TRIM(f.account_ref) = TRIM(c.account_ref)
		         AND TRIM(f.trading_day) = TRIM(c.trading_day))
		        OR (TRIM(f.account_ref) = '' AND TRIM(f.trading_day) = '' AND f.terminal = 0
		            AND NOT EXISTS (
		              SELECT 1 FROM all_confirmed_orders reused
		               WHERE reused.order_id = c.order_id
		                 AND reused.account_ref = c.account_ref
		                 AND UPPER(TRIM(reused.market)) = UPPER(TRIM(c.market))
		                 AND TRIM(reused.trading_day) <> TRIM(c.trading_day))))
			   AND (UPPER(TRIM(f.symbol)) <> UPPER(TRIM(c.symbol))
			        OR UPPER(TRIM(f.market)) <> UPPER(TRIM(c.market))
			        OR (TRIM(f.side) <> '' AND UPPER(TRIM(f.side)) <> UPPER(TRIM(c.side))))
			   AND NOT EXISTS (
			     SELECT 1 FROM all_confirmed_orders matching
			      WHERE matching.order_id = f.order_id
			        AND TRIM(matching.account_ref) = TRIM(f.account_ref)
			        AND UPPER(TRIM(matching.market)) = UPPER(TRIM(f.market))
			        AND TRIM(matching.trading_day) = TRIM(f.trading_day)
			        AND UPPER(TRIM(matching.symbol)) = UPPER(TRIM(f.symbol))
			        AND (TRIM(f.side) = '' OR UPPER(TRIM(matching.side)) = UPPER(TRIM(f.side))))
			UNION
		SELECT f.order_id, l.symbol, l.market, 'lineage endpoint identity is ambiguous or contradicts its intent'
		  FROM all_fill_snapshots f
		  JOIN valid_lineage l ON l.parent_order_id = f.order_id OR l.child_order_id = f.order_id
		 WHERE l.account_ref = ?
		   AND EXISTS (SELECT 1 FROM all_confirmed_orders selected
		                WHERE selected.account_ref = l.account_ref
		                  AND UPPER(TRIM(selected.market)) = UPPER(TRIM(l.market))
		                  AND TRIM(selected.trading_day) = TRIM(l.trading_day)
		                  AND (selected.order_id = l.parent_order_id OR selected.order_id = l.child_order_id))
		   AND ((TRIM(f.account_ref) = TRIM(l.account_ref)
		         AND TRIM(f.trading_day) = TRIM(l.trading_day))
		        OR (TRIM(f.account_ref) = '' AND TRIM(f.trading_day) = '' AND f.terminal = 0
		            AND NOT EXISTS (
		              SELECT 1 FROM all_confirmed_orders reused
		               WHERE reused.account_ref = l.account_ref
		                 AND UPPER(TRIM(reused.market)) = UPPER(TRIM(l.market))
		                 AND TRIM(reused.trading_day) <> TRIM(l.trading_day)
		                 AND (reused.order_id = l.parent_order_id OR reused.order_id = l.child_order_id))))
			   AND (UPPER(TRIM(f.symbol)) <> UPPER(TRIM(l.symbol))
			        OR UPPER(TRIM(f.market)) <> UPPER(TRIM(l.market))
			        OR (TRIM(f.side) <> '' AND UPPER(TRIM(f.side)) <> UPPER(TRIM(l.side))))
			   AND NOT EXISTS (
			     SELECT 1 FROM valid_lineage matching
			      WHERE (matching.parent_order_id = f.order_id OR matching.child_order_id = f.order_id)
			        AND TRIM(matching.account_ref) = TRIM(f.account_ref)
			        AND UPPER(TRIM(matching.market)) = UPPER(TRIM(f.market))
			        AND TRIM(matching.trading_day) = TRIM(f.trading_day)
			        AND UPPER(TRIM(matching.symbol)) = UPPER(TRIM(f.symbol))
			        AND (TRIM(f.side) = '' OR UPPER(TRIM(matching.side)) = UPPER(TRIM(f.side))))
		LIMIT 1`, string(StateConfirmed),
		RelationReplaces, string(KindAmend), string(StateConfirmed),
		RelationReplaces, string(KindAmend), string(StateConfirmed),
		accountRef, accountRef, accountRef).
		Scan(&orderID, &symbol, &market, &detail)
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("journal: checking tracked fill ownership for account %s: %w", accountRef, err)
	}
	evidence := fmt.Sprintf("%s: order %s, expected %s/%s", detail, orderID,
		normaliseMarket(market), normaliseSymbol(symbol))
	if _, _, enterErr := j.EnterReconcile(ctx, EnterReconcileRequest{
		AccountRef: accountRef,
		Cause:      ReconcileCauseIdentifierConflict,
		Evidence:   evidence,
	}); enterErr != nil {
		return fmt.Errorf("%w: %s; durable fail-closed record failed: %v",
			ErrTrackedFillIdentityConflict, evidence, enterErr)
	}
	return fmt.Errorf("%w: %s", ErrTrackedFillIdentityConflict, evidence)
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
	OrderID    string
	IntentID   string
	AccountRef string
	Market     string
	TradingDay string
	Symbol     string
	Side       string
	Quantity   string
	Price      string
	Currency   string
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
	rows, err := j.db.QueryContext(ctx, allFillSnapshotsCTE+`,
		confirmed_orders AS (
			SELECT a.broker_order_id AS order_id, a.recorded_at, a.rowid AS attempt_rowid,
			       i.id AS intent_id, i.account_ref, i.market, i.trading_day,
			       i.symbol, i.side, i.quantity, coalesce(i.price, '') AS price, i.currency
			  FROM mutation_attempts a
			  JOIN intents i ON i.id = a.intent_id
			 WHERE a.state = ? AND a.broker_order_id <> ''
		), selected_orders AS (
			SELECT c.* FROM confirmed_orders c
			 WHERE c.account_ref = ? AND c.market = ? AND c.symbol = ?
		)
		SELECT c.order_id, c.intent_id, c.account_ref, c.market, c.trading_day,
		       c.symbol, c.side, c.quantity, c.price, c.currency
		  FROM selected_orders c
		 WHERE NOT EXISTS (
		   SELECT 1 FROM all_fill_snapshots f
		    WHERE f.order_id = c.order_id AND f.terminal = 1
		      AND UPPER(TRIM(f.symbol)) = UPPER(TRIM(c.symbol))
		      AND UPPER(TRIM(f.market)) = UPPER(TRIM(c.market))
		      AND (TRIM(f.side) = '' OR UPPER(TRIM(f.side)) = UPPER(TRIM(c.side)))
		      AND ((TRIM(f.account_ref) = TRIM(c.account_ref)
		            AND TRIM(f.trading_day) = TRIM(c.trading_day))
		           OR (TRIM(f.account_ref) = '' AND TRIM(f.trading_day) = ''
		               AND NOT EXISTS (
		                 SELECT 1 FROM confirmed_orders reused
		                  WHERE reused.order_id = c.order_id
		                    AND reused.account_ref = c.account_ref
		                    AND UPPER(TRIM(reused.market)) = UPPER(TRIM(c.market))
		                    AND TRIM(reused.trading_day) <> TRIM(c.trading_day))))
		 )
		 ORDER BY c.recorded_at, c.attempt_rowid`,
		string(StateConfirmed), strings.TrimSpace(accountRef),
		normaliseMarket(market), normaliseSymbol(symbol))
	if err != nil {
		return nil, fmt.Errorf("journal: listing the working orders of %s: %w", symbol, err)
	}
	defer rows.Close()

	var out []LiveOrder
	for rows.Next() {
		var o LiveOrder
		if err := rows.Scan(&o.OrderID, &o.IntentID, &o.AccountRef, &o.Market, &o.TradingDay,
			&o.Symbol, &o.Side,
			&o.Quantity, &o.Price, &o.Currency); err != nil {
			return nil, fmt.Errorf("journal: listing the working orders of %s: %w", symbol, err)
		}
		out = append(out, o)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("journal: listing the working orders of %s: %w", symbol, err)
	}

	for i := range out {
		current, err := j.ResolveCurrentOrderIDScoped(ctx, out[i].OrderID, OrderLineageScope{
			AccountRef: out[i].AccountRef,
			Market:     out[i].Market,
			TradingDay: out[i].TradingDay,
			Symbol:     out[i].Symbol,
			Side:       out[i].Side,
		})
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

const schemaV16 = `
ALTER TABLE fill_snapshots ADD COLUMN account_ref TEXT NOT NULL DEFAULT '';
ALTER TABLE fill_snapshots ADD COLUMN trading_day TEXT NOT NULL DEFAULT '';
ALTER TABLE fill_snapshots ADD COLUMN side TEXT NOT NULL DEFAULT '';
CREATE INDEX idx_fill_snapshots_scope
    ON fill_snapshots(account_ref, market, trading_day, order_id);

CREATE TABLE scoped_lineage_edges (
	id                     INTEGER PRIMARY KEY AUTOINCREMENT,
	parent_order_id        TEXT NOT NULL,
	child_order_id         TEXT NOT NULL,
	relation               TEXT NOT NULL,
	parent_filled_quantity TEXT NOT NULL,
	requested_quantity     TEXT NOT NULL,
	account_ref            TEXT NOT NULL,
	market                 TEXT NOT NULL,
	trading_day            TEXT NOT NULL,
	symbol                 TEXT NOT NULL,
	side                   TEXT NOT NULL,
	intent_id              TEXT NOT NULL REFERENCES intents(id),
	attempt_id             TEXT NOT NULL REFERENCES mutation_attempts(id),
	created_at             TEXT NOT NULL,
	UNIQUE(parent_order_id, child_order_id, relation, account_ref, market,
	       trading_day, symbol, side, intent_id, attempt_id)
) STRICT;
CREATE INDEX idx_scoped_lineage_parent
	ON scoped_lineage_edges(account_ref, market, trading_day, symbol, side, parent_order_id);
CREATE INDEX idx_scoped_lineage_child
	ON scoped_lineage_edges(account_ref, market, trading_day, symbol, side, child_order_id);
`

// schemaV17 fixes the storage key without rewriting or replacing the released
// v2 table. That table remains the compatibility surface for v15 blank-scope
// evidence and for foreign keys created by schema v5. New fully scoped rows can
// coexist here even when a broker reuses the same opaque order id.
const schemaV17 = `
CREATE TABLE scoped_fill_snapshots (
	order_id          TEXT NOT NULL,
	account_ref       TEXT NOT NULL,
	market            TEXT NOT NULL,
	trading_day       TEXT NOT NULL,
	symbol            TEXT NOT NULL,
	side              TEXT NOT NULL,
	state             TEXT NOT NULL DEFAULT '',
	terminal          INTEGER NOT NULL DEFAULT 0,
	fail_closed       INTEGER NOT NULL DEFAULT 0,
	quantity          TEXT NOT NULL DEFAULT '0',
	filled_quantity   TEXT NOT NULL DEFAULT '0',
	average_price     TEXT NOT NULL DEFAULT '',
	filled_amount     TEXT NOT NULL DEFAULT '',
	broker_visible_at TEXT NOT NULL DEFAULT '',
	observed_at       TEXT NOT NULL DEFAULT '',
	committed_at      TEXT NOT NULL,
	reason_code       TEXT NOT NULL DEFAULT '',
	detail            TEXT NOT NULL DEFAULT '',
	PRIMARY KEY (account_ref, market, trading_day, symbol, side, order_id)
) STRICT;

CREATE INDEX idx_scoped_fill_snapshots_order
	ON scoped_fill_snapshots(order_id);
CREATE INDEX idx_scoped_fill_snapshots_symbol
	ON scoped_fill_snapshots(account_ref, market, trading_day, symbol, terminal);
`

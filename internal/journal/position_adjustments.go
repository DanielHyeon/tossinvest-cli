package journal

// position_adjustments.go is compare-and-append adjustment of the position
// projection (change add-core-domain task 6.2; position-ledger "조정 이벤트",
// design D4).
//
// # What an adjustment is for
//
// The projection is built from fills, and fills are not the whole truth: a
// person can trade the same account by hand, a broker can hold something the
// engine never bought, and a transition the state machine refuses leaves the
// projection deliberately frozen at a quantity the account disagrees with. The
// account is the authority (SHALL — 계좌 값이 권위다), and this is how its value
// reaches the row.
//
// It reaches it as an append-only *event*, never as an overwrite (SHALL NOT —
// Position 행 직접 덮어쓰기는 금지된다). The row in position_adjustments is the
// justification for the projection disagreeing with the fills, and a projection
// that could be silently rewritten would leave the disagreement unexplainable.
//
// # Compare-and-append
//
// An adjustment is computed outside the transaction that applies it: something
// reads the account, reads the projection, decides they differ, and only then
// asks for the difference to be recorded. Everything can move in between. So the
// caller's view is carried into the commit and re-checked there (SHALL — 커밋
// 트랜잭션 안에서 기대 이전 값·체결 watermark의 불변을 재검증해 어긋나면
// 폐기·재수집한다):
//
//	expected_prev_quantity  the quantity the adjustment was computed against
//	expected fill watermark the fill-event high-water it was computed against
//
// Both, not either. The quantity alone misses the case the watermark exists for:
// a buy and a sell of the same size landing between the read and the commit
// leave the quantity identical and the world different, and an adjustment
// applied on top of that is applied to a position that is not the one it was
// computed for.
//
// A mismatch is a *discard*, not a failure: ErrAdjustmentStale tells the caller
// to collect again, which is what reconciliation's snapshot loop does. Nothing
// is written, so a re-collection starts from a clean state.
//
// # Idempotency across a crash
//
// The adjustment id is derived from what the adjustment *is* — the scope, the
// expected previous quantity, the new quantity, the broker's as-of and the kind
// — so a retry of the same adjustment carries the same id. Re-applying is
// recognised by that id and reports the stored row rather than the stale error
// the expected-previous check would otherwise produce, because after a
// successful apply the quantity has legitimately moved. The row and the
// projection move in one transaction, so a crash leaves both or neither.
//
// # External positions
//
// A difference on a symbol with no local instance is folded in as a position
// with `entry_decision_id` NULL (SHALL). NULL is not a missing value here, it is
// the fact: no decision justifies the position, so there is no entry stop, so
// there is no t0 baseline, so the exit policy has nothing to protect it with and
// must not pretend otherwise (D4, D5).
//
// # Broker-behaviour claims
//
// None. `broker_as_of` is stamped by the caller from the account snapshot it
// read; this file records it and never interprets it.

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"github.com/JungHoonGhae/tossinvest-cli/internal/position"
	"github.com/JungHoonGhae/tossinvest-cli/internal/riskcalc"
)

// ErrAdjustmentStale means the adjustment was computed against a world that has
// since moved, and was discarded rather than applied to a different one. The
// caller re-collects and tries again.
var ErrAdjustmentStale = errors.New("journal: the adjustment was computed against a stale view")

// StaleAdjustmentError says which invariant moved and by how much, so a caller
// that re-collects can tell a busy symbol from a stuck one.
type StaleAdjustmentError struct {
	AccountRef string
	Symbol     string
	// Invariant is "quantity" or "fill watermark".
	Invariant string
	Expected  string
	Actual    string
}

func (e *StaleAdjustmentError) Error() string {
	return fmt.Sprintf(
		"journal: the adjustment for %s/%s expected %s %s but found %s; discarded for re-collection",
		e.AccountRef, e.Symbol, e.Invariant, e.Expected, e.Actual)
}

func (e *StaleAdjustmentError) Is(target error) bool { return target == ErrAdjustmentStale }

// AdjustmentRequest asks for one compare-and-append adjustment.
type AdjustmentRequest struct {
	// ID is optional; an empty one is derived from what the adjustment is, so a
	// retry after a crash carries the same id.
	ID string

	AccountRef string
	Market     string
	Symbol     string

	// Kind is EXTERNAL, MANUAL or UNKNOWN. internal/position.Classify decides it
	// from what was known at classification time; this layer records the verdict
	// and refuses one it does not recognise.
	Kind string

	// ExpectedPrevQuantity is the projected quantity the adjustment was computed
	// against, re-checked inside the commit transaction.
	ExpectedPrevQuantity string
	// ExpectedFillWatermark is the value FillWatermark returned when the
	// adjustment was computed, re-checked the same way.
	ExpectedFillWatermark int64

	// NewQuantity is the account's own quantity — the value the projection
	// converges to.
	NewQuantity string
	// NewAvgPrice is the account's cost basis, or "" when the broker reported
	// none. "" leaves the projection's existing basis alone rather than erasing
	// it: an unreported price is not a price of zero.
	NewAvgPrice string

	// BrokerAsOf is when the account snapshot this was computed from was
	// current. Required — an adjustment with no as-of cannot be ordered against
	// the fills it competes with.
	BrokerAsOf string
	// Evidence is what was observed, for the operator who has to act on it.
	// Required, the same rule EnterReconcile applies.
	Evidence string
}

// Adjustment is one stored position_adjustments row.
type Adjustment struct {
	ID                   string
	PositionID           string
	Kind                 string
	ExpectedPrevQuantity string
	PrevQuantity         string
	NewQuantity          string
	PrevAvgPrice         string
	NewAvgPrice          string
	BrokerAsOf           string
	Evidence             string
	CreatedAt            string
}

// AdjustmentResult is what applying one adjustment did.
type AdjustmentResult struct {
	Adjustment Adjustment
	// Position is the projection after the adjustment.
	Position Position
	// Applied is false when this exact adjustment was already on disk — a retry
	// after a crash, or a duplicate from a re-collection that recomputed the
	// same difference.
	Applied bool
	// OpenedInstance reports that the adjustment folded in a position the
	// projection did not have: an external or manual holding, with no entry
	// decision and therefore no exit policy.
	OpenedInstance bool
}

// FillWatermark returns the fill-event high-water for one symbol.
//
// It is the second half of the compare in compare-and-append. fill_events' id is
// an autoincrementing integer appended once per positive delta, so a value that
// has not moved is a proof that no fill of that symbol has been applied since —
// which is a stronger statement than "the quantity is the same", and it is the
// statement an adjustment actually depends on.
//
// Corrections are deliberately outside it. An EXECUTION_CORRECTION moves a cost
// basis and never a quantity, so it cannot invalidate an adjustment about a
// quantity; folding it in would discard adjustments for a change that does not
// affect them.
func (j *Journal) FillWatermark(ctx context.Context, symbol string) (int64, error) {
	var watermark int64
	if err := j.db.QueryRowContext(ctx,
		`SELECT coalesce(MAX(id), 0) FROM fill_events WHERE upper(trim(symbol)) = ?`,
		normaliseSymbol(symbol)).Scan(&watermark); err != nil {
		return 0, fmt.Errorf("journal: reading the fill watermark of %s: %w", symbol, err)
	}
	return watermark, nil
}

// ApplyPositionAdjustment records one adjustment and converges the projection to
// the account's value, in one transaction.
func (j *Journal) ApplyPositionAdjustment(ctx context.Context, req AdjustmentRequest) (AdjustmentResult, error) {
	req, err := req.normalised()
	if err != nil {
		return AdjustmentResult{}, err
	}
	now := j.nowString()

	tx, err := j.db.BeginTx(ctx, nil) // BEGIN IMMEDIATE
	if err != nil {
		return AdjustmentResult{}, fmt.Errorf("journal: starting the adjustment transaction: %w", err)
	}
	defer tx.Rollback()

	// 1. Already applied? Recognised before the compare, because after a
	//    successful apply the quantity has legitimately moved and the compare
	//    would report a crash recovery as a stale view.
	existing, err := scanAdjustment(tx.QueryRowContext(ctx, adjustmentSelect+" WHERE id = ?", req.ID))
	switch {
	case err == nil:
		stored, err := scanPosition(tx.QueryRowContext(ctx, positionSelect+" WHERE id = ?", existing.PositionID))
		if err != nil {
			return AdjustmentResult{}, err
		}
		return AdjustmentResult{Adjustment: existing, Position: stored}, nil
	case errors.Is(err, ErrAdjustmentNotFound):
	default:
		return AdjustmentResult{}, err
	}

	// 2. The compare. Both halves, inside the transaction that will write.
	current, err := scanPosition(tx.QueryRowContext(ctx, positionSelect+`
		 WHERE account_ref = ? AND market = ? AND symbol = ?
		 ORDER BY instance_seq DESC LIMIT 1`, req.AccountRef, req.Market, req.Symbol))
	held := "0"
	switch {
	case err == nil:
		held = current.Quantity
		if current.State == PositionClosed {
			// A closed instance holds nothing. An adjustment computed against it
			// is computed against zero, and a positive account quantity means the
			// account holds something the engine's ledger has already finished
			// with — a new, external instance rather than a resurrection.
			held = "0"
		}
	case errors.Is(err, ErrPositionNotFound):
	default:
		return AdjustmentResult{}, err
	}

	if same, cerr := sameDecimal(held, req.ExpectedPrevQuantity); cerr != nil {
		return AdjustmentResult{}, cerr
	} else if !same {
		return AdjustmentResult{}, &StaleAdjustmentError{
			AccountRef: req.AccountRef, Symbol: req.Symbol,
			Invariant: "quantity", Expected: req.ExpectedPrevQuantity, Actual: held,
		}
	}

	var watermark int64
	if err := tx.QueryRowContext(ctx,
		`SELECT coalesce(MAX(id), 0) FROM fill_events WHERE upper(trim(symbol)) = ?`,
		req.Symbol).Scan(&watermark); err != nil {
		return AdjustmentResult{}, fmt.Errorf(
			"journal: re-reading the fill watermark of %s: %w", req.Symbol, err)
	}
	if watermark != req.ExpectedFillWatermark {
		return AdjustmentResult{}, &StaleAdjustmentError{
			AccountRef: req.AccountRef, Symbol: req.Symbol, Invariant: "fill watermark",
			Expected: fmt.Sprintf("%d", req.ExpectedFillWatermark),
			Actual:   fmt.Sprintf("%d", watermark),
		}
	}

	// 3. Append the event, then converge the projection. Both or neither.
	result := AdjustmentResult{Applied: true}
	// A live instance is adjusted in place. An absent one, or one that already
	// reached CLOSED, means the account holds something no local instance
	// covers: the adjustment opens the next instance rather than reopening a
	// finished one (CLOSED 종결성).
	adjustInPlace := current.ID != "" && current.State != PositionClosed

	target := current
	if !adjustInPlace {
		next := int64(1)
		if current.ID != "" {
			next = current.InstanceSeq + 1
		}
		target = Position{
			ID:          PositionID(req.AccountRef, req.Market, req.Symbol, next),
			AccountRef:  req.AccountRef,
			Market:      req.Market,
			Symbol:      req.Symbol,
			InstanceSeq: next,
			// No decision justifies a position the engine did not open.
			EntryDecisionID: "",
			AvgPrice:        position.Unknown,
			Quantity:        "0",
			OpenedAt:        now,
		}
		result.OpenedInstance = true
	}

	adjustment := Adjustment{
		ID:                   req.ID,
		PositionID:           target.ID,
		Kind:                 req.Kind,
		ExpectedPrevQuantity: req.ExpectedPrevQuantity,
		PrevQuantity:         held,
		NewQuantity:          req.NewQuantity,
		PrevAvgPrice:         target.AvgPrice,
		NewAvgPrice:          firstNonEmpty(req.NewAvgPrice, target.AvgPrice),
		BrokerAsOf:           req.BrokerAsOf,
		Evidence:             req.Evidence,
		CreatedAt:            now,
	}

	if result.OpenedInstance {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO positions
			  (id, account_ref, market, symbol, instance_seq, entry_decision_id, state,
			   quantity, avg_price, opened_at, closed_at)
			VALUES (?,?,?,?,?,NULL,?,?,?,?,?)`,
			target.ID, target.AccountRef, target.Market, target.Symbol, target.InstanceSeq,
			convergedState(target.State, req.NewQuantity), req.NewQuantity,
			adjustment.NewAvgPrice, now, closedStamp(req.NewQuantity, now)); err != nil {
			return AdjustmentResult{}, fmt.Errorf(
				"journal: folding in the external position of %s/%s: %w", req.AccountRef, req.Symbol, err)
		}
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO position_adjustments
		  (id, position_id, kind, expected_prev_quantity, prev_quantity, new_quantity,
		   prev_avg_price, new_avg_price, broker_as_of, evidence, created_at)
		VALUES (?,?,?,?,?,?,?,?,?,?,?)`,
		adjustment.ID, adjustment.PositionID, adjustment.Kind, adjustment.ExpectedPrevQuantity,
		adjustment.PrevQuantity, adjustment.NewQuantity, adjustment.PrevAvgPrice,
		adjustment.NewAvgPrice, adjustment.BrokerAsOf, adjustment.Evidence,
		adjustment.CreatedAt); err != nil {
		return AdjustmentResult{}, fmt.Errorf(
			"journal: recording the adjustment of %s: %w", adjustment.PositionID, err)
	}

	if !result.OpenedInstance {
		if _, err := tx.ExecContext(ctx, `
			UPDATE positions SET state = ?, quantity = ?, avg_price = ?, closed_at = ?
			 WHERE id = ?`,
			convergedState(target.State, req.NewQuantity), req.NewQuantity, adjustment.NewAvgPrice,
			closedStamp(req.NewQuantity, now), target.ID); err != nil {
			return AdjustmentResult{}, fmt.Errorf(
				"journal: converging position %s to the account's value: %w", target.ID, err)
		}
	}

	converged, err := scanPosition(tx.QueryRowContext(ctx, positionSelect+" WHERE id = ?", target.ID))
	if err != nil {
		return AdjustmentResult{}, err
	}
	if err := tx.Commit(); err != nil {
		return AdjustmentResult{}, fmt.Errorf(
			"journal: committing the adjustment of %s: %w", target.ID, err)
	}
	result.Adjustment, result.Position = adjustment, converged
	return result, nil
}

// convergedState is the state a position is in once it holds the account's
// quantity.
//
// Zero is CLOSED and nothing else: an instance that holds nothing is over. A
// positive quantity leaves a live state alone — the projection may know an order
// is working, which an account snapshot cannot tell us — and lifts a closed or
// absent one to OPEN, because the account says it is held and no local order
// explains it.
func convergedState(current, quantity string) string {
	if isZeroDecimal(quantity) {
		return PositionClosed
	}
	switch current {
	case PositionOpening, PositionOpen, PositionScaling, PositionClosing:
		return current
	default:
		return PositionOpen
	}
}

func closedStamp(quantity, now string) any {
	if isZeroDecimal(quantity) {
		return now
	}
	return nil
}

func isZeroDecimal(s string) bool {
	cmp, err := riskcalc.CompareDecimal(orZero(s), "0")
	return err == nil && cmp == 0
}

func sameDecimal(a, b string) (bool, error) {
	cmp, err := riskcalc.CompareDecimal(orZero(a), orZero(b))
	if err != nil {
		return false, fmt.Errorf("%w: comparing quantities %q and %q: %v", ErrInvalidRequest, a, b, err)
	}
	return cmp == 0, nil
}

func (r AdjustmentRequest) normalised() (AdjustmentRequest, error) {
	r.AccountRef = strings.TrimSpace(r.AccountRef)
	r.Market = normaliseMarket(r.Market)
	r.Symbol = normaliseSymbol(r.Symbol)
	r.Kind = strings.TrimSpace(r.Kind)
	r.BrokerAsOf = strings.TrimSpace(r.BrokerAsOf)
	r.Evidence = strings.TrimSpace(r.Evidence)
	r.ID = strings.TrimSpace(r.ID)

	switch {
	case r.AccountRef == "":
		return r, fmt.Errorf("%w: an adjustment is scoped to an account; none was named", ErrInvalidRequest)
	case r.Market == "" || r.Symbol == "":
		return r, fmt.Errorf("%w: an adjustment is scoped to a market and a symbol", ErrInvalidRequest)
	case !position.ValidKind(position.Kind(r.Kind)):
		return r, fmt.Errorf(
			"%w: adjustment kind %q is not one this build records; an unclassified difference "+
				"has no audit meaning", ErrInvalidRequest, r.Kind)
	case r.BrokerAsOf == "":
		return r, fmt.Errorf(
			"%w: an adjustment records the broker snapshot it was computed from; none was named",
			ErrInvalidRequest)
	case r.Evidence == "":
		return r, fmt.Errorf(
			"%w: an adjustment requires the evidence that raised it", ErrInvalidRequest)
	case r.ExpectedFillWatermark < 0:
		return r, fmt.Errorf("%w: a fill watermark is not negative", ErrInvalidRequest)
	}

	var err error
	if r.ExpectedPrevQuantity, err = canonicalQuantity("expected previous quantity",
		orZero(r.ExpectedPrevQuantity)); err != nil {
		return r, err
	}
	if r.NewQuantity, err = canonicalQuantity("new quantity", orZero(r.NewQuantity)); err != nil {
		return r, err
	}
	if r.NewAvgPrice != "" {
		if r.NewAvgPrice, err = canonicalQuantity("new average price", r.NewAvgPrice); err != nil {
			return r, err
		}
	}
	if r.ID == "" {
		r.ID = adjustmentID(r)
	}
	return r, nil
}

func canonicalQuantity(field, value string) (string, error) {
	canonical, err := riskcalc.CanonicalDecimal(value)
	if err != nil {
		return "", fmt.Errorf("%w: %s %q is not a decimal", ErrInvalidRequest, field, value)
	}
	negative, err := riskcalc.IsNegativeDecimal(canonical)
	if err != nil {
		return "", fmt.Errorf("%w: %s %q: %v", ErrInvalidRequest, field, value, err)
	}
	if negative {
		return "", fmt.Errorf("%w: %s is %s; a long-only projection holds no negative quantity",
			ErrInvalidRequest, field, canonical)
	}
	return canonical, nil
}

// adjustmentID derives the id of an adjustment from what it is, so a retry of
// the same adjustment is recognised as one. The length prefix keeps ("ab","c")
// and ("a","bc") different keys, the same rule correctionID follows.
func adjustmentID(r AdjustmentRequest) string {
	h := sha256.New()
	for _, part := range []string{
		r.AccountRef, r.Market, r.Symbol, r.Kind,
		r.ExpectedPrevQuantity, r.NewQuantity, r.NewAvgPrice, r.BrokerAsOf,
	} {
		fmt.Fprintf(h, "%d:%s|", len(part), part)
	}
	return "adj-" + hex.EncodeToString(h.Sum(nil))[:24]
}

// ErrAdjustmentNotFound means no position_adjustments row matches.
var ErrAdjustmentNotFound = errors.New("journal: no such position adjustment")

const adjustmentSelect = `SELECT id, position_id, kind, expected_prev_quantity, prev_quantity,
	new_quantity, prev_avg_price, new_avg_price, broker_as_of, coalesce(evidence, ''), created_at
	FROM position_adjustments`

func scanAdjustment(row rowScanner) (Adjustment, error) {
	var a Adjustment
	err := row.Scan(&a.ID, &a.PositionID, &a.Kind, &a.ExpectedPrevQuantity, &a.PrevQuantity,
		&a.NewQuantity, &a.PrevAvgPrice, &a.NewAvgPrice, &a.BrokerAsOf, &a.Evidence, &a.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return Adjustment{}, ErrAdjustmentNotFound
	}
	if err != nil {
		return Adjustment{}, fmt.Errorf("journal: reading a position adjustment: %w", err)
	}
	return a, nil
}

// PositionAdjustments returns every adjustment recorded against one position,
// oldest first.
func (j *Journal) PositionAdjustments(ctx context.Context, positionID string) ([]Adjustment, error) {
	rows, err := j.db.QueryContext(ctx,
		adjustmentSelect+" WHERE position_id = ? ORDER BY created_at, rowid",
		strings.TrimSpace(positionID))
	if err != nil {
		return nil, fmt.Errorf("journal: reading the adjustments of %s: %w", positionID, err)
	}
	defer rows.Close()

	var out []Adjustment
	for rows.Next() {
		a, err := scanAdjustment(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("journal: reading the adjustments of %s: %w", positionID, err)
	}
	return out, nil
}

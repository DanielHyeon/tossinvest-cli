package journal

// position_projection.go is the storage half of the position projection
// (change add-core-domain task 6.1; position-ledger "Position 투영과 단일 권위").
//
// # The split, and why it is where it is
//
// internal/position owns the rule — the transition table, the cost basis, the
// refusals. This file owns everything that rule needs from the ledger and
// everything it produces for the ledger: the direction the fill's intent
// implies, the entry decision the instance was opened under, the replace edge
// that says an amendment carries the remainder, and the `positions` row the
// outcome is written to.
//
// It is an ApplyFunc, so it runs inside the fill's own transaction
// (apply_hook.go): the snapshot, the fill event and the projection commit
// together or not at all. Nothing here calls an exported *Journal method — the
// journal holds one connection and this code is inside it (apply_hook.go rule
// 4) — so every read goes through the handle.
//
// Wiring is one line at startup:
//
//	j.SetApplyHooks(journal.ApplyHooks{Project: journal.ProjectPosition})
//
// # What the projection refuses to do
//
// It writes no quantity that did not come from a fill (position-ledger: 직접
// 변이 API를 노출하지 않는다 SHALL NOT), and when the transition table refuses an
// event it writes nothing at all and enters the durable RECONCILE state instead.
// It does not clamp, does not net, does not guess: the account is the authority
// on a quantity the projection cannot derive, and task 6.2's adjustment path is
// how that authority reaches the row.
//
// A refusal does not fail the fill. The broker's snapshot is a fact and losing
// it would be worse than disagreeing with it — the fill is recorded, the
// projection is frozen, entries are blocked for the symbol, and the operator and
// the reconciliation loop both have what they need.
//
// # Direction
//
// From the intent behind the order, never from the fill (SHALL). Every local
// fill in this change's scope belongs to an order an intent produced; an order
// no local intent claims is an external order, and the projection leaves it
// alone rather than inventing a side for it. The direction source for a
// broker-resident triggered order is 2c's to define.
//
// # Broker-behaviour claims
//
// None. Every input is a column another part of the journal already derived.

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"github.com/JungHoonGhae/tossinvest-cli/internal/position"
)

// ProjectPosition is the Project apply hook: it turns one applied fill into the
// position it implies, inside the fill's transaction.
func ProjectPosition(ctx context.Context, tx *ApplyTx, fill AppliedFill) error {
	origin, found, err := resolveFillOrigin(ctx, tx, fill)
	if err != nil {
		return err
	}
	if !found {
		// An order no local intent claims. There is no side to project it in and
		// no instance it belongs to; the reconciliation comparison is what
		// notices the account holds something the engine did not buy, and task
		// 6.2 folds it in as an EXTERNAL adjustment.
		return nil
	}

	instance, err := currentInstance(ctx, tx, origin)
	if err != nil {
		return err
	}

	outcome, err := position.Apply(instance.Instance, position.Event{
		Role:              origin.Role,
		Delta:             fill.Delta,
		OrderQuantity:     fill.OrderedQuantity,
		OrderFilled:       fill.CumulativeQuantity,
		PrevOrderFilled:   fill.PrevCumulativeQuantity,
		OrderAvgPrice:     fill.AveragePrice,
		PrevOrderAvgPrice: fill.PrevAveragePrice,
		Terminal:          fill.Terminal,
		HasSuccessor:      origin.HasSuccessor,
	})
	if err != nil {
		// Unreadable input, not a disagreement: a decimal the ledger cannot read
		// is a bug in whatever wrote it, and letting the fill commit around it
		// would bury the bug under a projection that quietly stopped moving.
		return fmt.Errorf("journal: projecting the fill of %s: %w", fill.OrderID, err)
	}

	if outcome.Reconcile() {
		return enterReconcileInTx(ctx, tx, origin, outcome, fill)
	}
	return writeProjection(ctx, tx, origin, instance, outcome, fill)
}

// fillOrigin is what the ledger knows about the order behind a fill.
type fillOrigin struct {
	IntentID   string
	AccountRef string
	Market     string
	Symbol     string
	TradingDay string
	Side       string
	Role       position.Role
	// DecisionID is the decision that authorised the attempt, empty when the
	// attempt predates the decision contract or carried none.
	DecisionID string
	// HasSuccessor reports a `replaces` edge leaving this order.
	HasSuccessor bool
}

// resolveFillOrigin finds the intent behind a broker order and reads the
// direction, the account and the authorising decision off it.
//
// The join is the one NetPositions and accountRefForOrder already use: the
// attempt that named the broker order, and the intent that attempt belongs to.
// The most recent attempt wins, for the same reason it does there — an order id
// can be named by a retry as well as by the attempt that first got it.
//
// The bool is false when no attempt names the order. That is the definition of
// an external order and it is not an error.
func resolveFillOrigin(ctx context.Context, tx *ApplyTx, fill AppliedFill) (fillOrigin, bool, error) {
	rows, err := tx.Query(ctx, `
		SELECT i.id, i.account_ref, i.market, i.symbol, i.trading_day, i.side,
		       coalesce(a.decision_id, '')
		  FROM mutation_attempts a
		  JOIN intents i ON i.id = a.intent_id
		 WHERE a.broker_order_id = ? AND a.state = ?
		 ORDER BY a.recorded_at DESC, a.rowid DESC`, fill.OrderID, string(StateConfirmed))
	if err != nil {
		return fillOrigin{}, false, fmt.Errorf(
			"journal: resolving the intent behind order %s: %w", fill.OrderID, err)
	}
	defer rows.Close()

	type candidate struct {
		intentID string
		origin   fillOrigin
		side     string
	}
	var candidates []candidate
	for rows.Next() {
		var item candidate
		if err := rows.Scan(&item.intentID, &item.origin.AccountRef, &item.origin.Market,
			&item.origin.Symbol, &item.origin.TradingDay, &item.side, &item.origin.DecisionID); err != nil {
			return fillOrigin{}, false, fmt.Errorf(
				"journal: resolving the intent behind order %s: %w", fill.OrderID, err)
		}
		candidates = append(candidates, item)
	}
	if err := rows.Err(); err != nil {
		return fillOrigin{}, false, fmt.Errorf(
			"journal: resolving the intent behind order %s: %w", fill.OrderID, err)
	}
	rows.Close()
	if len(candidates) == 0 {
		return fillOrigin{}, false, nil
	}

	expectedAccount := strings.TrimSpace(firstNonEmpty(fill.AccountRef, candidates[0].origin.AccountRef))
	expectedMarket := normaliseMarket(firstNonEmpty(fill.Market, candidates[0].origin.Market))
	expectedSymbol := normaliseSymbol(firstNonEmpty(fill.Symbol, candidates[0].origin.Symbol))
	expectedDay := strings.TrimSpace(firstNonEmpty(fill.TradingDay, candidates[0].origin.TradingDay))
	expectedSide := strings.ToUpper(strings.TrimSpace(firstNonEmpty(fill.Side, candidates[0].side)))
	strictBrokerScope := strings.TrimSpace(fill.AccountRef) != ""
	conflict := strictBrokerScope && (strings.TrimSpace(fill.Market) == "" ||
		strings.TrimSpace(fill.Symbol) == "" || strings.TrimSpace(fill.TradingDay) == "" ||
		strings.TrimSpace(fill.Side) == "")
	var matches []candidate
	for _, item := range candidates {
		if strings.TrimSpace(item.origin.AccountRef) == expectedAccount &&
			normaliseMarket(item.origin.Market) == expectedMarket &&
			normaliseSymbol(item.origin.Symbol) == expectedSymbol &&
			strings.TrimSpace(item.origin.TradingDay) == expectedDay &&
			strings.ToUpper(strings.TrimSpace(item.side)) == expectedSide {
			matches = append(matches, item)
		}
	}
	if len(matches) == 0 || (!strictBrokerScope && len(matches) != len(candidates)) {
		conflict = true
	}
	if len(matches) > 0 {
		intentID := matches[0].intentID
		for _, item := range matches[1:] {
			conflict = conflict || item.intentID != intentID
		}
	}
	if conflict {
		evidence := fmt.Sprintf(
			"confirmed ownership for broker order %s conflicts with observed account/market/day/symbol/side %s/%s/%s/%s/%s",
			fill.OrderID, expectedAccount, expectedMarket, expectedDay, expectedSymbol, expectedSide)
		if err := enterReconcileScopeInTx(ctx, tx, expectedAccount, "",
			ReconcileCauseIdentifierConflict, evidence, fill.CommittedAt); err != nil {
			return fillOrigin{}, false, err
		}
		return fillOrigin{}, false, nil
	}

	origin := matches[0].origin
	side := matches[0].side
	origin.IntentID = matches[0].intentID
	origin.Side = side

	origin.Role, err = position.RoleForSide(side)
	if err != nil {
		return fillOrigin{}, false, fmt.Errorf(
			"journal: the direction of order %s: %w", fill.OrderID, err)
	}
	origin.Market = expectedMarket
	origin.Symbol = expectedSymbol
	origin.AccountRef = expectedAccount

	origin.HasSuccessor, err = hasReplaceSuccessor(ctx, tx, fill.OrderID, origin.IntentID)
	if err != nil {
		return fillOrigin{}, false, err
	}
	return origin, true, nil
}

// hasReplaceSuccessor reports whether an amendment created a child that carries
// this order's remainder. It is the lineage dimension of the transition table.
func hasReplaceSuccessor(ctx context.Context, tx *ApplyTx, orderID, intentID string) (bool, error) {
	rows, err := tx.Query(ctx,
		`SELECT 1 FROM scoped_lineage_edges
		  WHERE parent_order_id = ? AND intent_id = ? AND relation = ?
		 UNION
		 SELECT 1 FROM lineage_edges
		  WHERE parent_order_id = ? AND intent_id = ? AND relation = ?
		 LIMIT 1`,
		orderID, strings.TrimSpace(intentID), RelationReplaces,
		strings.TrimSpace(orderID), strings.TrimSpace(intentID), RelationReplaces)
	if err != nil {
		return false, fmt.Errorf("journal: reading the lineage of %s: %w", orderID, err)
	}
	defer rows.Close()
	found := rows.Next()
	if err := rows.Err(); err != nil {
		return false, fmt.Errorf("journal: reading the lineage of %s: %w", orderID, err)
	}
	return found, nil
}

// projectedInstance is the stored row the projection is about to advance.
type projectedInstance struct {
	position.Instance
	// ID and Seq are empty/zero when no instance of this symbol exists yet.
	ID   string
	Seq  int64
	Held bool
}

// currentInstance reads the latest instance of one symbol on one account.
//
// Latest by instance_seq, which is the ordering the UNIQUE key already
// establishes: a re-entry is a higher sequence than the instance it re-enters,
// and CLOSED instances are kept rather than replaced (position-ledger: 이전
// 인스턴스 기록은 보존된다).
func currentInstance(ctx context.Context, tx *ApplyTx, origin fillOrigin) (projectedInstance, error) {
	rows, err := tx.Query(ctx, `
		SELECT id, instance_seq, state, quantity, avg_price
		  FROM positions
		 WHERE account_ref = ? AND market = ? AND symbol = ?
		 ORDER BY instance_seq DESC
		 LIMIT 1`, origin.AccountRef, origin.Market, origin.Symbol)
	if err != nil {
		return projectedInstance{}, fmt.Errorf(
			"journal: reading the position of %s/%s: %w", origin.AccountRef, origin.Symbol, err)
	}
	defer rows.Close()

	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return projectedInstance{}, fmt.Errorf(
				"journal: reading the position of %s/%s: %w", origin.AccountRef, origin.Symbol, err)
		}
		// No row is FLAT. The projection never writes a FLAT row, so its absence
		// is the state rather than a missing one.
		return projectedInstance{Instance: position.Instance{State: position.Flat, Quantity: "0"}}, nil
	}
	var inst projectedInstance
	var state string
	if err := rows.Scan(&inst.ID, &inst.Seq, &state, &inst.Quantity, &inst.AvgPrice); err != nil {
		return projectedInstance{}, fmt.Errorf(
			"journal: reading the position of %s/%s: %w", origin.AccountRef, origin.Symbol, err)
	}
	inst.State = position.State(state)
	inst.Held = true
	return inst, nil
}

// writeProjection persists an allowed transition.
func writeProjection(ctx context.Context, tx *ApplyTx, origin fillOrigin,
	current projectedInstance, outcome position.Outcome, fill AppliedFill) error {
	if outcome.NewInstance || !current.Held {
		if outcome.Next == position.Flat {
			// E01–E03 and X01–X03: nothing was held and nothing moved, so there
			// is no instance to create. Writing a FLAT row here would put a
			// position in the ledger that never existed.
			return nil
		}
		return insertInstance(ctx, tx, origin, current.Seq+1, outcome, fill)
	}

	closedAt := any(nil)
	if outcome.Next == position.Closed {
		closedAt = fill.CommittedAt
	}
	if _, err := tx.Exec(ctx, `
		UPDATE positions
		   SET state = ?, quantity = ?, avg_price = ?, closed_at = ?
		 WHERE id = ?`,
		string(outcome.Next), outcome.Quantity, outcome.AvgPrice, closedAt, current.ID); err != nil {
		return fmt.Errorf("journal: projecting the fill of %s onto position %s: %w",
			fill.OrderID, current.ID, err)
	}
	return nil
}

// insertInstance opens a new position instance.
//
// opened_at is stamped here and never later, which is the code invariant the
// schema deliberately does not carry (issues.md, task 0.1): a row is created by
// the fill that first gave it a quantity, so an instance that exists has always
// been opened.
//
// entry_decision_id comes from the attempt that placed the entry order, and is
// NULL when there is none. NULL is the external/manual position: no decision
// justifies it, so it has no entry stop and is not an exit-policy target (D4).
func insertInstance(ctx context.Context, tx *ApplyTx, origin fillOrigin, seq int64,
	outcome position.Outcome, fill AppliedFill) error {
	id := PositionID(origin.AccountRef, origin.Market, origin.Symbol, seq)
	closedAt := any(nil)
	if outcome.Next == position.Closed {
		closedAt = fill.CommittedAt
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO positions
		  (id, account_ref, market, symbol, instance_seq, entry_decision_id, state,
		   quantity, avg_price, opened_at, closed_at)
		VALUES (?,?,?,?,?,?,?,?,?,?,?)`,
		id, origin.AccountRef, origin.Market, origin.Symbol, seq,
		nullableString(origin.DecisionID), string(outcome.Next),
		outcome.Quantity, outcome.AvgPrice, fill.CommittedAt, closedAt); err != nil {
		return fmt.Errorf("journal: opening position instance %d of %s/%s: %w",
			seq, origin.AccountRef, origin.Symbol, err)
	}
	return nil
}

// enterReconcileInTx raises the durable RECONCILE state for a refused
// transition, in the fill's own transaction.
//
// It is the same rule EnterReconcile applies — one active state per scope, the
// first observation kept — written against the handle because a hook cannot call
// an exported *Journal method (apply_hook.go rule 4). Symbol-scoped: the
// disagreement is about one symbol's quantity, and blocking the whole account
// for it would be a wider latch than the evidence supports.
func enterReconcileInTx(ctx context.Context, tx *ApplyTx, origin fillOrigin,
	outcome position.Outcome, fill AppliedFill) error {
	cause := reconcileCauseFor(outcome.Refusal)
	evidence := fmt.Sprintf(
		"position projection refused transition %s (%s): %s [order %s, delta %s, cumulative %s]",
		outcome.Row, outcome.Refusal, outcome.Reason, fill.OrderID, fill.Delta, fill.CumulativeQuantity)

	return enterReconcileScopeInTx(ctx, tx, origin.AccountRef, origin.Symbol, cause, evidence, fill.CommittedAt)
}

func enterReconcileScopeInTx(ctx context.Context, tx *ApplyTx, accountRef, symbol, cause, evidence,
	enteredAt string,
) error {
	accountRef = strings.TrimSpace(accountRef)
	symbol = normaliseSymbol(symbol)
	rows, err := tx.Query(ctx,
		`SELECT 1 FROM reconcile_states
		  WHERE account_ref = ? AND ((? = '' AND symbol IS NULL) OR symbol = ?)
		    AND released_at IS NULL LIMIT 1`,
		accountRef, symbol, symbol)
	if err != nil {
		return fmt.Errorf("journal: reading the RECONCILE state of %s/%s: %w",
			accountRef, symbol, err)
	}
	active := rows.Next()
	scanErr := rows.Err()
	rows.Close()
	if scanErr != nil {
		return fmt.Errorf("journal: reading the RECONCILE state of %s/%s: %w",
			accountRef, symbol, scanErr)
	}
	if active {
		// Already blocked for this symbol. Re-entering would either fail the
		// partial UNIQUE index or move `entered_at` forward on every poll, and a
		// state whose "since" advances forever looks new forever.
		return nil
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO reconcile_states
		  (id, account_ref, symbol, cause, evidence, entered_at, released_at, release_cause)
		VALUES (?,?,?,?,?,?,NULL,NULL)`,
		reconcileStateID(accountRef, symbol, cause, enteredAt),
		accountRef, nullableString(symbol), cause, evidence, enteredAt); err != nil {
		return fmt.Errorf("journal: entering RECONCILE for %s/%s: %w",
			accountRef, symbol, err)
	}
	return nil
}

// reconcileCauseFor maps a projection refusal onto the RECONCILE causes the
// ledger already enumerates (execution_contract.go).
//
// Two causes, because two different things went wrong. An oversell is the local
// quantity and the broker's fills disagreeing — QUANTITY_MISMATCH, literally.
// The other three are a fill that cannot be placed in any position the ledger
// has — ATTRIBUTION_FAILED, which is the cause for "a broker record could not be
// attributed to anything local".
func reconcileCauseFor(refusal position.Refusal) string {
	if refusal == position.RefusalOversell {
		return ReconcileCauseQuantityMismatch
	}
	return ReconcileCauseAttributionFailed
}

// PositionID derives the stable identifier of one position instance.
//
// Derived rather than minted, the same way correctionID and reconcileStateID
// are: the projection has no id scheme of its own, and a re-applied fill must
// address the row it addressed the first time. The length prefix keeps
// ("ab","c") and ("a","bc") different keys.
func PositionID(accountRef, market, symbol string, seq int64) string {
	h := sha256.New()
	for _, part := range []string{accountRef, market, symbol, fmt.Sprintf("%d", seq)} {
		fmt.Fprintf(h, "%d:%s|", len(part), part)
	}
	return "pos-" + hex.EncodeToString(h.Sum(nil))[:24]
}

// LookupPosition returns one projected instance by id.
func (j *Journal) LookupPosition(ctx context.Context, id string) (Position, error) {
	return scanPosition(j.db.QueryRowContext(ctx,
		positionSelect+" WHERE id = ?", strings.TrimSpace(id)))
}

// CurrentPosition returns the latest instance of one symbol on one account.
//
// Latest, not "open": a caller asking what the projection believes about a
// symbol needs the closed instance too, because "it closed" is an answer and
// "no rows" is a different one.
func (j *Journal) CurrentPosition(ctx context.Context, accountRef, market, symbol string) (Position, error) {
	return scanPosition(j.db.QueryRowContext(ctx, positionSelect+`
		 WHERE account_ref = ? AND market = ? AND symbol = ?
		 ORDER BY instance_seq DESC LIMIT 1`,
		strings.TrimSpace(accountRef), normaliseMarket(market), normaliseSymbol(symbol)))
}

// Positions returns every instance of one account, oldest first.
func (j *Journal) Positions(ctx context.Context, accountRef string) ([]Position, error) {
	rows, err := j.db.QueryContext(ctx, positionSelect+
		" WHERE account_ref = ? ORDER BY market, symbol, instance_seq",
		strings.TrimSpace(accountRef))
	if err != nil {
		return nil, fmt.Errorf("journal: listing the positions of %s: %w", accountRef, err)
	}
	defer rows.Close()

	var out []Position
	for rows.Next() {
		p, err := scanPosition(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("journal: listing the positions of %s: %w", accountRef, err)
	}
	return out, nil
}

// Position is one projected position instance.
type Position struct {
	ID          string
	AccountRef  string
	Market      string
	Symbol      string
	InstanceSeq int64
	// EntryDecisionID is empty for an externally or manually acquired position:
	// no decision justifies it. It is set once by the fill that opens the
	// instance and never written again.
	EntryDecisionID string
	// AdoptionID is empty until the position is adopted into exit management
	// (adoption.go). It is the second — and last — record that can justify a
	// baseline, and it too is set once.
	AdoptionID string
	State      string
	// Quantity is a decimal string; AvgPrice is a decimal string or "" when the
	// cost basis is not known.
	Quantity string
	AvgPrice string
	OpenedAt string
	ClosedAt string
}

// ExitEligible reports whether the exit policy may manage this position. It is
// the single predicate (internal/position) and never an inline column test.
func (p Position) ExitEligible() bool {
	return position.ExitEligible(p.EntryDecisionID, p.AdoptionID)
}

// Adopted reports that the position's eligibility comes from an adoption record
// rather than from an entry decision. The two open an exit state from different
// sources, so the branch is a stored fact rather than an inference.
func (p Position) Adopted() bool { return strings.TrimSpace(p.AdoptionID) != "" }

// ErrPositionNotFound means no projected instance matches.
var ErrPositionNotFound = errors.New("journal: no projected position")

const positionSelect = `SELECT id, account_ref, market, symbol, instance_seq,
	coalesce(entry_decision_id, ''), coalesce(adoption_id, ''), state, quantity, avg_price,
	coalesce(opened_at, ''), coalesce(closed_at, '') FROM positions`

func scanPosition(row rowScanner) (Position, error) {
	var p Position
	err := row.Scan(&p.ID, &p.AccountRef, &p.Market, &p.Symbol, &p.InstanceSeq,
		&p.EntryDecisionID, &p.AdoptionID, &p.State, &p.Quantity, &p.AvgPrice,
		&p.OpenedAt, &p.ClosedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return Position{}, ErrPositionNotFound
	}
	if err != nil {
		return Position{}, fmt.Errorf("journal: reading a position: %w", err)
	}
	return p, nil
}

// normaliseSymbol and normaliseMarket give the projection one spelling per
// venue. The comparison in position-ledger is at symbol level (D4: 비교는 심볼
// 수준 합산), so two spellings of one symbol would be two positions.
func normaliseSymbol(s string) string { return strings.ToUpper(strings.TrimSpace(s)) }

func normaliseMarket(s string) string { return strings.ToLower(strings.TrimSpace(s)) }

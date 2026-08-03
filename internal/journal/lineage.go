package journal

// lineage.go records how one broker order replaced another
// (harden-execution-base task 2.8).
//
// The official API answers a modify with a *new* order number. So an amendment is
// not a mutation of an order row, it is an edge: the parent closes carrying
// whatever it filled, and a child carries the requested remainder. Without that
// edge the engine sees an order vanish and a stranger appear — which reads as
// "flat, then a new position" when the truth is "same exposure, new id".
//
// New file, no edits to durability.go: the transition primitives there are
// released and pinned by their own tests. resolveWithLineage below runs its own
// transaction rather than growing transitionOpts a hook, and reuses the shared
// legality check so the two paths cannot disagree about what a legal move is.

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

// RelationReplaces is the only lineage relation this phase records.
const RelationReplaces = "replaces"

// ErrLineageCycle means the recorded replace chain loops.
var ErrLineageCycle = errors.New("journal: order lineage contains a cycle")

// LineageEdge is one replace relation.
//
// Quantities are decimal strings for the same reason the intent's are: a broker's
// decimal quantity has no exact binary float form, and this is an audit record.
type LineageEdge struct {
	ParentOrderID string
	ChildOrderID  string
	Relation      string
	// ParentFilledQuantity is what the parent had filled at replace time. It is
	// the number that makes the child's remainder add up.
	ParentFilledQuantity string
	// RequestedQuantity is what the child was asked to carry.
	RequestedQuantity string
	// IntentID and AttemptID are filled in on read; on write they come from the
	// attempt recording the edge.
	IntentID  string
	AttemptID string
	CreatedAt string
}

// OrderLineageScope is the canonical local ownership identity used while
// following broker replacement ids. Broker order ids are not globally unique,
// so a resolver that carries only the id can cross accounts or trading days.
//
// Every field is required. Market, symbol and side are compared
// case-insensitively after trimming; account and trading day are compared as
// trimmed exact values. The broker order id itself remains opaque.
type OrderLineageScope struct {
	AccountRef string
	Market     string
	TradingDay string
	Symbol     string
	Side       string
}

func (s OrderLineageScope) canonical() (OrderLineageScope, error) {
	s.AccountRef = strings.TrimSpace(s.AccountRef)
	s.Market = normaliseMarket(s.Market)
	s.TradingDay = strings.TrimSpace(s.TradingDay)
	s.Symbol = normaliseSymbol(s.Symbol)
	s.Side = strings.ToUpper(strings.TrimSpace(s.Side))
	switch {
	case s.AccountRef == "":
		return OrderLineageScope{}, fmt.Errorf("%w: a lineage scope needs an account ref", ErrInvalidRequest)
	case s.Market != "kr" && s.Market != "us":
		return OrderLineageScope{}, fmt.Errorf("%w: a lineage scope needs a supported market", ErrInvalidRequest)
	case s.TradingDay == "":
		return OrderLineageScope{}, fmt.Errorf("%w: a lineage scope needs a trading day", ErrInvalidRequest)
	case s.Symbol == "":
		return OrderLineageScope{}, fmt.Errorf("%w: a lineage scope needs a symbol", ErrInvalidRequest)
	case s.Side != "BUY" && s.Side != "SELL":
		return OrderLineageScope{}, fmt.Errorf("%w: a lineage scope needs a BUY or SELL side", ErrInvalidRequest)
	}
	return s, nil
}

func (e LineageEdge) validate() error {
	switch {
	case strings.TrimSpace(e.ParentOrderID) == "":
		return fmt.Errorf("%w: a lineage edge needs a parent order id", ErrInvalidRequest)
	case strings.TrimSpace(e.ChildOrderID) == "":
		return fmt.Errorf("%w: a lineage edge needs a child order id", ErrInvalidRequest)
	case strings.TrimSpace(e.ParentOrderID) == strings.TrimSpace(e.ChildOrderID):
		return fmt.Errorf("%w: an order cannot replace itself (%s)", ErrInvalidRequest, e.ParentOrderID)
	case strings.TrimSpace(e.ParentFilledQuantity) == "":
		return fmt.Errorf("%w: a lineage edge needs the parent's filled quantity", ErrInvalidRequest)
	case strings.TrimSpace(e.RequestedQuantity) == "":
		return fmt.Errorf("%w: a lineage edge needs the requested quantity", ErrInvalidRequest)
	}
	return nil
}

// ResolveConfirmedWithLineage confirms an attempt and records the replace edge it
// discovered, in one transaction.
//
// One transaction is the requirement, not an optimisation (order-execution: "두
// 기록은 동일 트랜잭션에서 커밋된다"). A confirmed amend without its edge would
// leave the engine holding an order id that no longer exists at the broker, with
// nothing to point at the order that took its place.
func (a *Attempt) ResolveConfirmedWithLineage(ctx context.Context, edge LineageEdge, reasonCode, detail string) error {
	if err := edge.validate(); err != nil {
		return err
	}
	if edge.Relation == "" {
		edge.Relation = RelationReplaces
	}
	if edge.Relation != RelationReplaces {
		return fmt.Errorf("%w: unknown lineage relation %q", ErrInvalidRequest, edge.Relation)
	}
	return a.resolveWithLineage(ctx, edge, firstNonEmpty(reasonCode, ReasonResolvedFound), detail)
}

// resolveWithLineage performs the settle and the edge insert in one
// BEGIN IMMEDIATE transaction.
func (a *Attempt) resolveWithLineage(ctx context.Context, edge LineageEdge, reasonCode, detail string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	from := []AttemptState{StateInDoubt, StateAcked}
	now := a.j.nowString()

	tx, err := a.j.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("journal: starting the lineage transaction for %s: %w", a.id, err)
	}
	defer tx.Rollback()

	var current AttemptState
	if err := tx.QueryRowContext(ctx,
		"SELECT state FROM mutation_attempts WHERE id = ?", a.id).Scan(&current); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("%w: %s", ErrAttemptNotFound, a.id)
		}
		return fmt.Errorf("journal: reading attempt %s: %w", a.id, err)
	}
	if err := checkTransitionAllowed(a.id, current, StateConfirmed, from); err != nil {
		return err
	}

	// The child's identifier is stored exactly as the broker sent it, the same
	// rule ResolveConfirmed follows (order-execution "브로커 식별자의 opaque
	// 취급": 저장 SHALL 원문, 비교 SHALL 바이트 동일). edge.validate above already
	// did the only job trimming has here — judging whether there is a name at
	// all. Storing a trimmed copy would put an identifier the broker never issued
	// in the attempt row, and the later byte comparison would silently miss it.
	res, err := tx.ExecContext(ctx,
		`UPDATE mutation_attempts
		    SET state = ?, settled_at = ?, broker_order_id = ?, reason_code = ?, detail = ?
		  WHERE id = ? AND state = ?`,
		string(StateConfirmed), now, edge.ChildOrderID, reasonCode, detail,
		a.id, string(current))
	if err != nil {
		return fmt.Errorf("journal: confirming attempt %s: %w", a.id, err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("journal: confirming attempt %s: %w", a.id, err)
	}
	if affected != 1 {
		return &StateError{AttemptID: a.id, Want: from, Got: current, To: StateConfirmed}
	}

	if _, err := tx.ExecContext(ctx,
		`INSERT INTO attempt_transitions (attempt_id, from_state, to_state, at, reason_code, detail)
		 VALUES (?,?,?,?,?,?)`,
		a.id, string(current), string(StateConfirmed), now, reasonCode, detail); err != nil {
		return fmt.Errorf("journal: recording the %s->CONFIRMED transition for %s: %w", current, a.id, err)
	}

	if _, err := tx.ExecContext(ctx,
		`INSERT INTO lineage_edges
		   (parent_order_id, child_order_id, relation, parent_filled_quantity,
		    requested_quantity, intent_id, attempt_id, created_at)
		 VALUES (?,?,?,?,?,?,?,?)
		 ON CONFLICT(parent_order_id, child_order_id, relation) DO NOTHING`,
		strings.TrimSpace(edge.ParentOrderID), strings.TrimSpace(edge.ChildOrderID), edge.Relation,
		strings.TrimSpace(edge.ParentFilledQuantity), strings.TrimSpace(edge.RequestedQuantity),
		a.intentID, a.id, now); err != nil {
		return fmt.Errorf("journal: recording lineage %s->%s: %w",
			edge.ParentOrderID, edge.ChildOrderID, err)
	}

	// The legacy table's global parent/child uniqueness cannot represent a
	// broker reusing the same pair in another account or trading session. Keep
	// writing it for compatibility, and write the canonical owner beside it in
	// the additive v16 table. This insert is in the same transaction as the
	// confirmation: a confirmed replacement without scoped ownership evidence
	// is not a state we may publish.
	scoped, err := tx.ExecContext(ctx, `
		INSERT INTO scoped_lineage_edges
		       (parent_order_id, child_order_id, relation, parent_filled_quantity,
		        requested_quantity, account_ref, market, trading_day, symbol, side,
		        intent_id, attempt_id, created_at)
		SELECT ?, ?, ?, ?, ?, TRIM(i.account_ref), LOWER(TRIM(i.market)),
		       TRIM(i.trading_day), UPPER(TRIM(i.symbol)), UPPER(TRIM(i.side)),
		       i.id, a.id, ?
		  FROM mutation_attempts a
		  JOIN intents i ON i.id = a.intent_id
		 WHERE a.id = ? AND a.intent_id = ?
		   AND TRIM(a.account_ref) = TRIM(i.account_ref)
		ON CONFLICT(parent_order_id, child_order_id, relation, account_ref, market,
		            trading_day, symbol, side, intent_id, attempt_id) DO NOTHING`,
		edge.ParentOrderID, edge.ChildOrderID, edge.Relation,
		strings.TrimSpace(edge.ParentFilledQuantity), strings.TrimSpace(edge.RequestedQuantity),
		now, a.id, a.intentID)
	if err != nil {
		return fmt.Errorf("journal: recording scoped lineage %s->%s: %w",
			edge.ParentOrderID, edge.ChildOrderID, err)
	}
	affected, err = scoped.RowsAffected()
	if err != nil {
		return fmt.Errorf("journal: recording scoped lineage %s->%s: %w",
			edge.ParentOrderID, edge.ChildOrderID, err)
	}
	if affected != 1 {
		return fmt.Errorf("%w: lineage %s->%s has no unique local intent ownership",
			ErrInvalidRequest, edge.ParentOrderID, edge.ChildOrderID)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("journal: committing the confirmation and lineage of %s: %w", a.id, err)
	}
	a.state = StateConfirmed
	return nil
}

const lineageSelect = `SELECT parent_order_id, child_order_id, relation, parent_filled_quantity,
	requested_quantity, intent_id, attempt_id, created_at FROM lineage_edges`

// LineageChildren returns the edges leaving an order.
func (j *Journal) LineageChildren(ctx context.Context, parentOrderID string) ([]LineageEdge, error) {
	rows, err := j.db.QueryContext(ctx, lineageSelect+" WHERE parent_order_id = ? ORDER BY id",
		strings.TrimSpace(parentOrderID))
	if err != nil {
		return nil, fmt.Errorf("journal: reading lineage of %s: %w", parentOrderID, err)
	}
	defer rows.Close()

	var out []LineageEdge
	for rows.Next() {
		var e LineageEdge
		if err := rows.Scan(&e.ParentOrderID, &e.ChildOrderID, &e.Relation, &e.ParentFilledQuantity,
			&e.RequestedQuantity, &e.IntentID, &e.AttemptID, &e.CreatedAt); err != nil {
			return nil, fmt.Errorf("journal: reading lineage of %s: %w", parentOrderID, err)
		}
		out = append(out, e)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("journal: reading lineage of %s: %w", parentOrderID, err)
	}
	return out, nil
}

// ResolveCurrentOrderID follows the replace chain from an order id to the order
// that carries its exposure now.
//
// An unknown id resolves to itself: the caller may be asking about an order that
// was never amended, and answering "not found" would make every caller special-case
// the common path. A cycle is refused rather than followed — it means the record is
// wrong, and looping forever on a live account's order chain is not an option.
func (j *Journal) ResolveCurrentOrderID(ctx context.Context, orderID string) (string, error) {
	current := strings.TrimSpace(orderID)
	if current == "" {
		return "", fmt.Errorf("%w: no order id", ErrInvalidRequest)
	}
	seen := map[string]bool{current: true}
	for {
		children, err := j.LineageChildren(ctx, current)
		if err != nil {
			return "", err
		}
		var next string
		for _, edge := range children {
			if edge.Relation == RelationReplaces {
				next = edge.ChildOrderID
				break
			}
		}
		if next == "" {
			return current, nil
		}
		if seen[next] {
			return "", fmt.Errorf("%w: %s leads back to %s", ErrLineageCycle, orderID, next)
		}
		seen[next] = true
		current = next
	}
}

// ResolveCurrentOrderIDScoped follows only replacement edges proven by one
// confirmed local AMEND in the supplied canonical ownership scope.
//
// An edge is evidence only when its lineage intent and attempt agree, the
// attempt names the edge's exact parent and child, and account/market/trading
// day/symbol/side all match. Edges from another account or session are ignored,
// even when the broker reused the same parent id and inserted that edge first.
// More than one valid child in the same scope is an identity conflict: choosing
// either would silently invent which broker order still carries exposure.
//
// As with ResolveCurrentOrderID, an id with no valid successor resolves to
// itself and a cycle is refused.
func (j *Journal) ResolveCurrentOrderIDScoped(ctx context.Context, orderID string, scope OrderLineageScope) (string, error) {
	if strings.TrimSpace(orderID) == "" {
		return "", fmt.Errorf("%w: no order id", ErrInvalidRequest)
	}
	canonical, err := scope.canonical()
	if err != nil {
		return "", err
	}

	current := orderID
	seen := map[string]bool{current: true}
	for {
		children, err := j.scopedLineageChildren(ctx, current, canonical)
		if err != nil {
			return "", err
		}
		switch len(children) {
		case 0:
			return current, nil
		case 1:
			// Continue below.
		default:
			return "", j.recordScopedLineageConflict(ctx, current, children, canonical)
		}

		next := children[0]
		if seen[next] {
			return "", fmt.Errorf("%w: %s leads back to %s", ErrLineageCycle, orderID, next)
		}
		seen[next] = true
		current = next
	}
}

func (j *Journal) scopedLineageChildren(ctx context.Context, parentOrderID string, scope OrderLineageScope) ([]string, error) {
	current, err := j.scopedLineageChildrenV16(ctx, parentOrderID, scope)
	if err != nil {
		return nil, err
	}
	legacy, err := j.legacyScopedLineageChildren(ctx, parentOrderID, scope)
	if err != nil {
		return nil, err
	}

	// A database may contain a pre-v16 legacy child and a different child
	// recorded after migration. Merge both evidence sources and deduplicate the
	// ordinary dual-write copy; otherwise preferring v16 outright would hide a
	// genuine same-scope branch.
	children := make([]string, 0, 2)
	seen := make(map[string]bool, len(current)+len(legacy))
	for _, child := range append(current, legacy...) {
		if seen[child] {
			continue
		}
		seen[child] = true
		children = append(children, child)
		if len(children) == 2 {
			break
		}
	}
	return children, nil
}

func (j *Journal) scopedLineageChildrenV16(ctx context.Context, parentOrderID string, scope OrderLineageScope) ([]string, error) {
	return j.queryScopedLineageChildren(ctx, "v16", `
		SELECT DISTINCT s.child_order_id
		  FROM scoped_lineage_edges s
		  JOIN mutation_attempts a
		    ON a.id = s.attempt_id AND a.intent_id = s.intent_id
		  JOIN intents i ON i.id = s.intent_id
		 WHERE s.parent_order_id = ? AND s.relation = ?
		   AND a.kind = ? AND a.state = ?
		   AND a.target_order_id = s.parent_order_id
		   AND a.broker_order_id = s.child_order_id
		   AND TRIM(a.account_ref) = TRIM(i.account_ref)
		   AND s.account_ref = ? AND s.market = ? AND s.trading_day = ?
		   AND s.symbol = ? AND s.side = ?
		   AND s.account_ref = TRIM(i.account_ref)
		   AND s.market = LOWER(TRIM(i.market))
		   AND s.trading_day = TRIM(i.trading_day)
		   AND s.symbol = UPPER(TRIM(i.symbol))
		   AND s.side = UPPER(TRIM(i.side))
		 ORDER BY s.child_order_id
		 LIMIT 2`, parentOrderID, scope)
}

func (j *Journal) legacyScopedLineageChildren(ctx context.Context, parentOrderID string, scope OrderLineageScope) ([]string, error) {
	return j.queryScopedLineageChildren(ctx, "legacy", `
		SELECT DISTINCT l.child_order_id
		  FROM lineage_edges l
		  JOIN mutation_attempts a
		    ON a.id = l.attempt_id AND a.intent_id = l.intent_id
		  JOIN intents i ON i.id = l.intent_id
		 WHERE l.parent_order_id = ? AND l.relation = ?
		   AND a.kind = ? AND a.state = ?
		   AND a.target_order_id = l.parent_order_id
		   AND a.broker_order_id = l.child_order_id
		   AND TRIM(a.account_ref) = TRIM(i.account_ref)
		   AND TRIM(i.account_ref) = ?
		   AND LOWER(TRIM(i.market)) = ?
		   AND TRIM(i.trading_day) = ?
		   AND UPPER(TRIM(i.symbol)) = ?
		   AND UPPER(TRIM(i.side)) = ?
		 ORDER BY l.child_order_id
		 LIMIT 2`, parentOrderID, scope)
}

func (j *Journal) queryScopedLineageChildren(ctx context.Context, source, query, parentOrderID string, scope OrderLineageScope) ([]string, error) {
	rows, err := j.db.QueryContext(ctx, query,
		parentOrderID, RelationReplaces, string(KindAmend), string(StateConfirmed),
		scope.AccountRef, scope.Market, scope.TradingDay, scope.Symbol, scope.Side)
	if err != nil {
		return nil, fmt.Errorf("journal: reading %s scoped lineage of %s: %w", source, parentOrderID, err)
	}
	defer rows.Close()

	children := make([]string, 0, 2)
	for rows.Next() {
		var child string
		if err := rows.Scan(&child); err != nil {
			return nil, fmt.Errorf("journal: reading %s scoped lineage of %s: %w", source, parentOrderID, err)
		}
		children = append(children, child)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("journal: reading %s scoped lineage of %s: %w", source, parentOrderID, err)
	}
	return children, nil
}

func (j *Journal) recordScopedLineageConflict(ctx context.Context, parentOrderID string, children []string, scope OrderLineageScope) error {
	conflict := fmt.Errorf(
		"%w: order %s has multiple confirmed replacement successors in %s/%s/%s/%s/%s",
		ErrTrackedFillIdentityConflict, parentOrderID, scope.AccountRef, scope.Market,
		scope.TradingDay, scope.Symbol, scope.Side)
	evidence := fmt.Sprintf(
		"broker order %s has multiple confirmed local replacement successors %q in canonical scope %s/%s/%s/%s/%s",
		parentOrderID, children, scope.AccountRef, scope.Market, scope.TradingDay, scope.Symbol, scope.Side)
	state, _, err := j.EnterReconcile(ctx, EnterReconcileRequest{
		AccountRef: scope.AccountRef,
		Cause:      ReconcileCauseIdentifierConflict,
		Evidence:   evidence,
	})
	if err != nil {
		return fmt.Errorf("%w; recording the durable identifier conflict: %v", conflict, err)
	}
	if state.Cause != ReconcileCauseIdentifierConflict {
		return fmt.Errorf("%w; account already has active RECONCILE cause %s", conflict, state.Cause)
	}
	return conflict
}

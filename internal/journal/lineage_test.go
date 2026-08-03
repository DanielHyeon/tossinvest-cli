package journal

import (
	"context"
	"errors"
	"testing"
)

func TestResolveCurrentOrderIDScopedIgnoresEarlierReplacementOutsideCanonicalScope(t *testing.T) {
	base := OrderLineageScope{
		AccountRef: "acct-1",
		Market:     "us",
		TradingDay: "2026-08-03",
		Symbol:     "AAPL",
		Side:       "BUY",
	}

	tests := []struct {
		name   string
		mutate func(*OrderLineageScope)
	}{
		{name: "account", mutate: func(s *OrderLineageScope) { s.AccountRef = "acct-2" }},
		{name: "market", mutate: func(s *OrderLineageScope) { s.Market = "kr" }},
		{name: "trading day", mutate: func(s *OrderLineageScope) { s.TradingDay = "2026-08-04" }},
		{name: "symbol", mutate: func(s *OrderLineageScope) { s.Symbol = "MSFT" }},
		{name: "side", mutate: func(s *OrderLineageScope) { s.Side = "SELL" }},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			j := openTestJournal(t)
			ctx := context.Background()
			outside := base
			tc.mutate(&outside)

			// The out-of-scope edge is deliberately inserted first. The legacy
			// bare-id resolver follows insertion order and would select this child.
			recordConfirmedReplacement(t, j, outside, "shared-parent", "outside-child", "outside")
			recordConfirmedReplacement(t, j, base, "shared-parent", "owned-child", "owned")

			got, err := j.ResolveCurrentOrderIDScoped(ctx, "shared-parent", base)
			if err != nil {
				t.Fatalf("ResolveCurrentOrderIDScoped: %v", err)
			}
			if got != "owned-child" {
				t.Fatalf("current order = %q, want the exact scoped child owned-child", got)
			}
		})
	}
}

func TestResolveCurrentOrderIDScopedRequiresConfirmedMatchingAmendOwnership(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()
	scope := OrderLineageScope{
		AccountRef: "acct-1",
		Market:     "us",
		TradingDay: "2026-08-03",
		Symbol:     "AAPL",
		Side:       "BUY",
	}

	recordUnconfirmedReplacement(t, j, scope, "shared-parent", "unconfirmed-child", "unconfirmed")
	recordConfirmedReplacement(t, j, scope, "shared-parent", "owned-child", "owned")

	got, err := j.ResolveCurrentOrderIDScoped(ctx, "shared-parent", scope)
	if err != nil {
		t.Fatalf("ResolveCurrentOrderIDScoped: %v", err)
	}
	if got != "owned-child" {
		t.Fatalf("current order = %q, want the confirmed local AMEND child owned-child", got)
	}
}

func TestResolveCurrentOrderIDScopedPreservesSamePairReusedByAnotherAccountAndDay(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()
	first := OrderLineageScope{
		AccountRef: "acct-2",
		Market:     "us",
		TradingDay: "2026-08-02",
		Symbol:     "AAPL",
		Side:       "BUY",
	}
	selected := first
	selected.AccountRef = "acct-1"
	selected.TradingDay = "2026-08-03"

	// The legacy table's global parent/child UNIQUE absorbs the second insert.
	// The additive scoped evidence must retain both local AMENDs nevertheless.
	recordConfirmedReplacement(t, j, first, "reused-parent", "reused-child", "first-scope")
	recordConfirmedReplacement(t, j, selected, "reused-parent", "reused-child", "selected-scope")

	var scopedRows int
	if err := j.db.QueryRowContext(ctx, `SELECT count(*) FROM scoped_lineage_edges
		WHERE parent_order_id='reused-parent' AND child_order_id='reused-child'`).Scan(&scopedRows); err != nil {
		t.Fatalf("count scoped lineage rows: %v", err)
	}
	if scopedRows != 2 {
		t.Fatalf("scoped lineage rows = %d, want both account/day owners", scopedRows)
	}

	got, err := j.ResolveCurrentOrderIDScoped(ctx, "reused-parent", selected)
	if err != nil {
		t.Fatalf("ResolveCurrentOrderIDScoped: %v", err)
	}
	if got != "reused-child" {
		t.Fatalf("current order = %q, want the selected account/day successor reused-child", got)
	}
}

func TestResolveCurrentOrderIDScopedPreservesOpaqueBrokerOrderIDs(t *testing.T) {
	j := openTestJournal(t)
	scope := OrderLineageScope{
		AccountRef: "acct-1",
		Market:     "us",
		TradingDay: "2026-08-03",
		Symbol:     "AAPL",
		Side:       "BUY",
	}
	const parent = " parent-with-space "
	const child = " child-with-space "
	recordConfirmedReplacement(t, j, scope, parent, child, "opaque")

	got, err := j.ResolveCurrentOrderIDScoped(context.Background(), parent, scope)
	if err != nil {
		t.Fatalf("ResolveCurrentOrderIDScoped: %v", err)
	}
	if got != child {
		t.Fatalf("current order = %q, want opaque child bytes %q", got, child)
	}
}

func TestResolveCurrentOrderIDScopedFallsBackToValidatedLegacyLineage(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()
	scope := OrderLineageScope{
		AccountRef: "acct-1",
		Market:     "us",
		TradingDay: "2026-08-03",
		Symbol:     "AAPL",
		Side:       "BUY",
	}
	recordConfirmedReplacement(t, j, scope, "legacy-parent", "legacy-child", "legacy")
	if _, err := j.db.ExecContext(ctx, `DELETE FROM scoped_lineage_edges
		WHERE parent_order_id='legacy-parent' AND child_order_id='legacy-child'`); err != nil {
		t.Fatalf("remove additive row to model pre-v16 lineage: %v", err)
	}

	got, err := j.ResolveCurrentOrderIDScoped(ctx, "legacy-parent", scope)
	if err != nil {
		t.Fatalf("ResolveCurrentOrderIDScoped legacy fallback: %v", err)
	}
	if got != "legacy-child" {
		t.Fatalf("current order = %q, want validated legacy successor", got)
	}
}

func TestResolveCurrentOrderIDScopedRejectsLegacyAndV16ChildrenInOneScope(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()
	scope := OrderLineageScope{
		AccountRef: "acct-1",
		Market:     "us",
		TradingDay: "2026-08-03",
		Symbol:     "AAPL",
		Side:       "BUY",
	}

	// Model an edge written before v16, then a different edge written after the
	// migration in the same canonical scope. The ordinary post-v16 dual-write of
	// v16-child is deduplicated; legacy-child must still make the chain ambiguous.
	recordConfirmedReplacement(t, j, scope, "mixed-parent", "legacy-child", "mixed-legacy")
	if _, err := j.db.ExecContext(ctx, `DELETE FROM scoped_lineage_edges
		WHERE parent_order_id='mixed-parent' AND child_order_id='legacy-child'`); err != nil {
		t.Fatalf("remove additive row to model pre-v16 lineage: %v", err)
	}
	recordConfirmedReplacement(t, j, scope, "mixed-parent", "v16-child", "mixed-v16")

	_, err := j.ResolveCurrentOrderIDScoped(ctx, "mixed-parent", scope)
	if !errors.Is(err, ErrTrackedFillIdentityConflict) {
		t.Fatalf("error = %v, want conflict across validated legacy and v16 children", err)
	}
}

func TestResolveCurrentOrderIDScopedRejectsMultipleChildrenInOneScope(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()
	scope := OrderLineageScope{
		AccountRef: "acct-1",
		Market:     "us",
		TradingDay: "2026-08-03",
		Symbol:     "AAPL",
		Side:       "BUY",
	}

	recordConfirmedReplacement(t, j, scope, "shared-parent", "first-child", "first")
	recordConfirmedReplacement(t, j, scope, "shared-parent", "second-child", "second")

	_, err := j.ResolveCurrentOrderIDScoped(ctx, "shared-parent", scope)
	if !errors.Is(err, ErrTrackedFillIdentityConflict) {
		t.Fatalf("error = %v, want ErrTrackedFillIdentityConflict", err)
	}
	states, stateErr := j.ActiveReconcileStates(ctx)
	if stateErr != nil {
		t.Fatalf("ActiveReconcileStates: %v", stateErr)
	}
	if len(states) != 1 || states[0].AccountRef != scope.AccountRef ||
		states[0].Cause != ReconcileCauseIdentifierConflict || !states[0].AccountWide() {
		t.Fatalf("durable conflict states = %+v, want one account-wide IDENTIFIER_CONFLICT", states)
	}
}

func TestResolveCurrentOrderIDScopedValidatesCompleteScope(t *testing.T) {
	j := openTestJournal(t)
	_, err := j.ResolveCurrentOrderIDScoped(context.Background(), "parent", OrderLineageScope{
		AccountRef: "acct-1",
		Market:     "us",
		TradingDay: "2026-08-03",
		Symbol:     "AAPL",
	})
	if !errors.Is(err, ErrInvalidRequest) {
		t.Fatalf("error = %v, want ErrInvalidRequest for an incomplete scope", err)
	}
}

func recordConfirmedReplacement(t *testing.T, j *Journal, scope OrderLineageScope, parent, child, suffix string) {
	t.Helper()
	attempt := prepareReplacement(t, j, scope, parent, suffix)
	ctx := context.Background()
	if err := attempt.MarkDispatchStarted(ctx); err != nil {
		t.Fatalf("mark replacement dispatch started: %v", err)
	}
	if err := attempt.MarkAcked(ctx, child); err != nil {
		t.Fatalf("mark replacement acked: %v", err)
	}
	if err := attempt.ResolveConfirmedWithLineage(ctx, LineageEdge{
		ParentOrderID:        parent,
		ChildOrderID:         child,
		ParentFilledQuantity: "0",
		RequestedQuantity:    "1",
	}, ReasonResolvedFound, "test replacement confirmed"); err != nil {
		t.Fatalf("confirm replacement lineage: %v", err)
	}
}

func recordUnconfirmedReplacement(t *testing.T, j *Journal, scope OrderLineageScope, parent, child, suffix string) {
	t.Helper()
	attempt := prepareReplacement(t, j, scope, parent, suffix)
	ctx := context.Background()
	if err := attempt.MarkDispatchStarted(ctx); err != nil {
		t.Fatalf("mark replacement dispatch started: %v", err)
	}
	if err := attempt.MarkAcked(ctx, child); err != nil {
		t.Fatalf("mark replacement acked: %v", err)
	}
	if _, err := j.db.ExecContext(ctx, `
		INSERT INTO lineage_edges
		       (parent_order_id, child_order_id, relation, parent_filled_quantity,
		        requested_quantity, intent_id, attempt_id, created_at)
		VALUES (?, ?, ?, '0', '1', ?, ?, '2026-08-03T00:00:00Z')`,
		parent, child, RelationReplaces, attempt.IntentID(), attempt.ID()); err != nil {
		t.Fatalf("record unconfirmed replacement lineage: %v", err)
	}
}

func prepareReplacement(t *testing.T, j *Journal, scope OrderLineageScope, parent, suffix string) *Attempt {
	t.Helper()
	attempt, err := j.Prepare(context.Background(), PrepareRequest{
		Intent: Intent{
			ID:          "lineage-intent-" + suffix,
			Market:      scope.Market,
			TradingDay:  scope.TradingDay,
			AccountRef:  scope.AccountRef,
			Symbol:      scope.Symbol,
			Side:        scope.Side,
			OrderType:   "LIMIT",
			Quantity:    "1",
			Price:       "100",
			Currency:    "USD",
			Source:      "lineage-test",
			Fingerprint: "lineage-fp-" + suffix,
		},
		Kind:          KindAmend,
		AttemptID:     "lineage-attempt-" + suffix,
		TargetOrderID: parent,
	})
	if err != nil {
		t.Fatalf("prepare replacement: %v", err)
	}
	return attempt
}

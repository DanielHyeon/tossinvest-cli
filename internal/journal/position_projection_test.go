package journal

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/position"
)

// position_projection_test.go is the storage half of task 6.1: the transition
// table's decisions reaching a `positions` row, inside the fill's own
// transaction.
//
// internal/position's tests prove the rule; these prove the wiring — the
// direction comes from the intent, the entry decision comes from the attempt,
// the lineage edge is consulted, a refusal freezes the row and raises RECONCILE
// instead of guessing, and everything lands in the same commit as the fill.

// order is one placed order the projection can attribute a fill to.
type order struct {
	intentID   string
	attemptID  string
	orderID    string
	decisionID string
	side       string
	symbol     string
	market     string
	account    string
	tradingDay string
	quantity   string
}

func (o order) withDefaults() order {
	if o.side == "" {
		o.side = "BUY"
	}
	if o.symbol == "" {
		o.symbol = "005930"
	}
	if o.market == "" {
		o.market = "kr"
	}
	if o.account == "" {
		o.account = "acct-1"
	}
	if o.quantity == "" {
		o.quantity = "10"
	}
	if o.tradingDay == "" {
		o.tradingDay = "2026-03-30"
	}
	return o
}

// place records the intent and the acked attempt that names a broker order, so
// the projection has an intent to take the direction from.
func place(t *testing.T, j *Journal, o order) order {
	t.Helper()
	o = o.withDefaults()
	ctx := context.Background()

	req := PrepareRequest{
		Intent: Intent{
			ID: o.intentID, Market: o.market, TradingDay: o.tradingDay,
			AccountRef: o.account, Symbol: o.symbol, Side: o.side,
			OrderType: "LIMIT", TimeInForce: "DAY", Quantity: o.quantity,
			Price: "70000", Currency: "KRW", Source: "engine/test",
			Fingerprint: "fp-" + o.intentID,
		},
		Kind:      KindPlace,
		AttemptID: o.attemptID,
	}
	if o.decisionID != "" {
		insertDecision(t, j, o.decisionID, "nonce-"+o.decisionID)
		req.DecisionID = o.decisionID
		req.SafetyClass = SafetyClassExposureRaising
	}
	attempt, err := j.Prepare(ctx, req)
	if err != nil {
		t.Fatalf("Prepare(%s): %v", o.attemptID, err)
	}
	if err := attempt.MarkDispatchStarted(ctx); err != nil {
		t.Fatalf("MarkDispatchStarted(%s): %v", o.attemptID, err)
	}
	if err := attempt.MarkAcked(ctx, o.orderID); err != nil {
		t.Fatalf("MarkAcked(%s): %v", o.attemptID, err)
	}
	if err := attempt.Settle(ctx, StateConfirmed, ReasonBrokerAcknowledged, "acked"); err != nil {
		t.Fatalf("Settle(%s): %v", o.attemptID, err)
	}
	return o
}

// fillOf is one cumulative snapshot of a placed order.
func fillOf(o order, filled, avg string) FillObservation {
	return FillObservation{
		OrderID: o.orderID, Symbol: o.symbol, Market: o.market, AccountRef: o.account,
		TradingDay: o.tradingDay, Side: o.side,
		State: "OPEN_PARTIALLY_FILLED", Quantity: o.quantity,
		FilledQuantity: filled, AveragePrice: avg,
		ObservedAt: "2026-03-30T00:30:00Z",
	}
}

func terminalFill(o order, filled, avg string) FillObservation {
	obs := fillOf(o, filled, avg)
	obs.State, obs.Terminal = "CLOSED_FILLED", true
	return obs
}

// projectingJournal is a journal with the projection hook bound, which is the
// only way the projection ever runs.
func projectingJournal(t *testing.T) *Journal {
	t.Helper()
	j := openTestJournal(t)
	if err := j.SetApplyHooks(ApplyHooks{Project: ProjectPosition}); err != nil {
		t.Fatalf("SetApplyHooks: %v", err)
	}
	return j
}

func currentPosition(t *testing.T, j *Journal, o order) Position {
	t.Helper()
	p, err := j.CurrentPosition(context.Background(), o.account, o.market, o.symbol)
	if err != nil {
		t.Fatalf("CurrentPosition(%s): %v", o.symbol, err)
	}
	return p
}

// TestTheFirstFillOpensTheInstance is the projection's entry point: a fill, and
// only a fill, brings a position into existence.
func TestTheFirstFillOpensTheInstance(t *testing.T) {
	j := projectingJournal(t)
	ctx := context.Background()
	o := place(t, j, order{intentID: "i-1", attemptID: "a-1", orderID: "o-1", decisionID: "d-1"})

	if _, err := j.CurrentPosition(ctx, o.account, o.market, o.symbol); !errors.Is(err, ErrPositionNotFound) {
		t.Fatalf("a symbol with no fills must have no row, got %v", err)
	}

	if _, err := j.RecordFill(ctx, fillOf(o, "3", "70000")); err != nil {
		t.Fatalf("RecordFill: %v", err)
	}

	p := currentPosition(t, j, o)
	if p.State != PositionOpening {
		t.Errorf("state = %s, want OPENING (3 of 10)", p.State)
	}
	if p.Quantity != "3" || p.AvgPrice != "70000" {
		t.Errorf("projection = (%s, %s), want (3, 70000)", p.Quantity, p.AvgPrice)
	}
	if p.InstanceSeq != 1 {
		t.Errorf("instance_seq = %d, want 1", p.InstanceSeq)
	}
	if p.EntryDecisionID != "d-1" {
		t.Errorf("entry_decision_id = %q, want the decision that authorised the entry", p.EntryDecisionID)
	}
	if p.OpenedAt == "" {
		t.Error("opened_at is empty; an instance that exists has always been opened")
	}
	if p.ClosedAt != "" {
		t.Errorf("closed_at = %q on an open instance", p.ClosedAt)
	}
	if p.ID != PositionID(o.account, o.market, o.symbol, 1) {
		t.Errorf("id = %q, want the derived id so a re-applied fill addresses the same row", p.ID)
	}
	if !p.ExitEligible() {
		t.Error("an engine position with an entry decision is an exit-policy target")
	}
}

func TestCollidingConfirmedOrderIdentityBlocksWithoutProjectingAPosition(t *testing.T) {
	j := projectingJournal(t)
	ctx := context.Background()
	owned := place(t, j, order{
		intentID: "owned-intent", attemptID: "owned-attempt", orderID: "account-scoped-id",
		account: "acct-1", market: "us", symbol: "AAPL",
	})
	_ = place(t, j, order{
		intentID: "other-intent", attemptID: "other-attempt", orderID: "account-scoped-id",
		account: "acct-1", market: "us", symbol: "AAPL",
	})

	if _, err := j.RecordFill(ctx, fillOf(owned, "1", "200")); err != nil {
		t.Fatalf("RecordFill: %v", err)
	}
	if _, err := j.CurrentPosition(ctx, "acct-1", "us", "AAPL"); !errors.Is(err, ErrPositionNotFound) {
		t.Fatalf("colliding broker id projected a local position: %v", err)
	}
	active, err := j.ActiveReconcileStates(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(active) != 1 || active[0].AccountRef != "acct-1" ||
		active[0].Cause != ReconcileCauseIdentifierConflict || active[0].Symbol != "" {
		t.Fatalf("active reconcile states = %+v, want account-wide identifier conflict", active)
	}
}

func TestReusedOrderIDIsOwnedByTheObservedTradingDay(t *testing.T) {
	j := projectingJournal(t)
	ctx := context.Background()
	_ = place(t, j, order{
		intentID: "prior-day-intent", attemptID: "prior-day-attempt", orderID: "reused-id",
		decisionID: "prior-day-decision", account: "acct-1", market: "us", symbol: "AAPL",
		tradingDay: "2026-03-29",
	})
	current := place(t, j, order{
		intentID: "current-day-intent", attemptID: "current-day-attempt", orderID: "reused-id",
		decisionID: "current-day-decision", account: "acct-1", market: "us", symbol: "AAPL",
		tradingDay: "2026-03-30",
	})

	if _, err := j.RecordFill(ctx, fillOf(current, "1", "200")); err != nil {
		t.Fatalf("RecordFill: %v", err)
	}
	position := currentPosition(t, j, current)
	if position.EntryDecisionID != "current-day-decision" {
		t.Fatalf("entry decision = %q, want current trading day ownership", position.EntryDecisionID)
	}
	active, err := j.ActiveReconcileStates(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(active) != 0 {
		t.Fatalf("trading-day scoped reuse raised a false conflict: %+v", active)
	}
}

// TestTheProjectionRunsTheWholeRoundTrip walks one instance from the first
// partial to CLOSED and checks the state at every step against the transition
// table's row for it.
func TestTheProjectionRunsTheWholeRoundTrip(t *testing.T) {
	j := projectingJournal(t)
	ctx := context.Background()
	buy := place(t, j, order{intentID: "i-1", attemptID: "a-1", orderID: "o-1", decisionID: "d-1"})

	if _, err := j.RecordFill(ctx, fillOf(buy, "3", "70000")); err != nil {
		t.Fatal(err)
	}
	if got := currentPosition(t, j, buy).State; got != PositionOpening {
		t.Fatalf("after the first partial: %s, want OPENING", got)
	}

	// 3 @ 70000 then 7 @ 71000 → the order's average is 70700.
	if _, err := j.RecordFill(ctx, terminalFill(buy, "10", "70700")); err != nil {
		t.Fatal(err)
	}
	opened := currentPosition(t, j, buy)
	if opened.State != PositionOpen || opened.Quantity != "10" || opened.AvgPrice != "70700" {
		t.Fatalf("after completion = (%s, %s, %s), want (OPEN, 10, 70700)",
			opened.State, opened.Quantity, opened.AvgPrice)
	}

	sell := place(t, j, order{intentID: "i-2", attemptID: "a-2", orderID: "o-2",
		side: "SELL", quantity: "10"})
	if _, err := j.RecordFill(ctx, fillOf(sell, "4", "72000")); err != nil {
		t.Fatal(err)
	}
	closing := currentPosition(t, j, buy)
	if closing.State != PositionClosing || closing.Quantity != "6" {
		t.Fatalf("after the partial exit = (%s, %s), want (CLOSING, 6)", closing.State, closing.Quantity)
	}
	if closing.AvgPrice != "70700" {
		t.Errorf("average = %s; a sell realises P&L and does not move the unit cost", closing.AvgPrice)
	}

	if _, err := j.RecordFill(ctx, terminalFill(sell, "10", "72500")); err != nil {
		t.Fatal(err)
	}
	closed := currentPosition(t, j, buy)
	if closed.State != PositionClosed || closed.Quantity != "0" {
		t.Fatalf("after the full exit = (%s, %s), want (CLOSED, 0)", closed.State, closed.Quantity)
	}
	if closed.ClosedAt == "" {
		t.Error("closed_at is empty on a CLOSED instance")
	}
}

// TestReEntryOpensTheNextInstanceAndKeepsTheOld is 청산 후 재진입: a new
// instance, a new decision, and the previous instance still on disk.
func TestReEntryOpensTheNextInstanceAndKeepsTheOld(t *testing.T) {
	j := projectingJournal(t)
	ctx := context.Background()

	buy := place(t, j, order{intentID: "i-1", attemptID: "a-1", orderID: "o-1", decisionID: "d-1"})
	if _, err := j.RecordFill(ctx, terminalFill(buy, "10", "70000")); err != nil {
		t.Fatal(err)
	}
	sell := place(t, j, order{intentID: "i-2", attemptID: "a-2", orderID: "o-2", side: "SELL"})
	if _, err := j.RecordFill(ctx, terminalFill(sell, "10", "72000")); err != nil {
		t.Fatal(err)
	}

	again := place(t, j, order{intentID: "i-3", attemptID: "a-3", orderID: "o-3",
		decisionID: "d-2", quantity: "5"})
	if _, err := j.RecordFill(ctx, terminalFill(again, "5", "80000")); err != nil {
		t.Fatal(err)
	}

	all, err := j.Positions(ctx, buy.account)
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 2 {
		t.Fatalf("instances = %d, want 2 (the closed one is preserved)", len(all))
	}
	if all[0].InstanceSeq != 1 || all[0].State != PositionClosed {
		t.Errorf("instance 1 = (%d, %s), want the closed original", all[0].InstanceSeq, all[0].State)
	}
	second := all[1]
	if second.InstanceSeq != 2 || second.State != PositionOpen {
		t.Fatalf("instance 2 = (%d, %s), want (2, OPEN)", second.InstanceSeq, second.State)
	}
	if second.EntryDecisionID != "d-2" {
		t.Errorf("entry_decision_id = %q, want the new entry's decision", second.EntryDecisionID)
	}
	if second.AvgPrice != "80000" {
		t.Errorf("average = %s; the closed instance's basis must not price the new one", second.AvgPrice)
	}
}

// TestAnAmendmentKeepsOneInstance is lineage 승계 through the journal: the
// parent goes terminal at 3 of 10 with a replace edge, and the child's fills
// complete the same instance rather than opening a scale-in.
func TestAnAmendmentKeepsOneInstance(t *testing.T) {
	j := projectingJournal(t)
	ctx := context.Background()

	parent := place(t, j, order{intentID: "i-1", attemptID: "a-1", orderID: "o-1", decisionID: "d-1"})
	if _, err := j.RecordFill(ctx, fillOf(parent, "3", "70000")); err != nil {
		t.Fatal(err)
	}

	// The amendment: a second attempt on the same intent, confirmed with the
	// replace edge the official API's new order number implies.
	amend, err := j.Prepare(ctx, PrepareRequest{
		Intent:        testIntentFor(parent),
		Kind:          KindAmend,
		AttemptID:     "a-1-amend",
		TargetOrderID: "o-1",
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := amend.MarkDispatchStarted(ctx); err != nil {
		t.Fatal(err)
	}
	if err := amend.MarkAcked(ctx, "o-1b"); err != nil {
		t.Fatal(err)
	}
	if err := amend.ResolveConfirmedWithLineage(ctx, LineageEdge{
		ParentOrderID: "o-1", ChildOrderID: "o-1b", Relation: RelationReplaces,
		ParentFilledQuantity: "3", RequestedQuantity: "7",
	}, "", "amended"); err != nil {
		t.Fatal(err)
	}

	// The parent goes terminal having filled 3 of 10. With the edge, the entry
	// is not over.
	parentDone := terminalFill(parent, "3", "70000")
	if _, err := j.RecordFill(ctx, parentDone); err != nil {
		t.Fatal(err)
	}
	if got := currentPosition(t, j, parent).State; got != PositionOpening {
		t.Fatalf("after the amended parent went terminal: %s, want OPENING (the child carries the rest)", got)
	}

	child := order{orderID: "o-1b", symbol: parent.symbol, market: parent.market,
		account: parent.account, tradingDay: parent.tradingDay, quantity: "7", side: "BUY"}
	if _, err := j.RecordFill(ctx, terminalFill(child, "7", "71000")); err != nil {
		t.Fatal(err)
	}

	all, err := j.Positions(ctx, parent.account)
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 1 {
		t.Fatalf("instances = %d, want 1: an amendment is the same position, not a new one", len(all))
	}
	if all[0].State != PositionOpen || all[0].Quantity != "10" {
		t.Fatalf("after the child completed = (%s, %s), want (OPEN, 10)", all[0].State, all[0].Quantity)
	}
	if all[0].AvgPrice != "70700" {
		t.Errorf("average = %s, want 70700 across the whole replace chain", all[0].AvgPrice)
	}
}

func testIntentFor(o order) Intent {
	return Intent{
		ID: o.intentID, Market: o.market, TradingDay: o.tradingDay,
		AccountRef: o.account, Symbol: o.symbol, Side: o.side,
		OrderType: "LIMIT", TimeInForce: "DAY", Quantity: o.quantity,
		Price: "70000", Currency: "KRW", Source: "engine/test",
		Fingerprint: "fp-" + o.intentID,
	}
}

// TestACorrectionMovesTheBasisAndNotTheQuantity is the delta-0 contract from
// task 0.3 arriving at the projection: an EXECUTION_CORRECTION reaches the hook
// with Delta "0", and it is a normal input rather than a no-op.
func TestACorrectionMovesTheBasisAndNotTheQuantity(t *testing.T) {
	j := projectingJournal(t)
	ctx := context.Background()
	o := place(t, j, order{intentID: "i-1", attemptID: "a-1", orderID: "o-1", decisionID: "d-1"})

	if _, err := j.RecordFill(ctx, fillOf(o, "10", "70000")); err != nil {
		t.Fatal(err)
	}
	before := currentPosition(t, j, o)

	res, err := j.RecordFill(ctx, fillOf(o, "10", "70500"))
	if err != nil {
		t.Fatal(err)
	}
	if !res.Corrected || res.Delta != "0" {
		t.Fatalf("restatement = %+v, want a correction with no delta", res)
	}

	after := currentPosition(t, j, o)
	if after.Quantity != before.Quantity {
		t.Errorf("quantity moved from %s to %s; a correction restates a price", before.Quantity, after.Quantity)
	}
	if after.AvgPrice != "70500" {
		t.Errorf("average = %s, want 70500 — the restated cost basis", after.AvgPrice)
	}
}

// TestARefusedSnapshotLeavesTheProjectionAlone is the other side of the delta-0
// contract: hooks do not fire for a fail-closed snapshot, so a snapshot the
// ledger declined to believe never reaches the position.
func TestARefusedSnapshotLeavesTheProjectionAlone(t *testing.T) {
	j := projectingJournal(t)
	ctx := context.Background()
	o := place(t, j, order{intentID: "i-1", attemptID: "a-1", orderID: "o-1", decisionID: "d-1"})

	if _, err := j.RecordFill(ctx, fillOf(o, "6", "70000")); err != nil {
		t.Fatal(err)
	}
	res, err := j.RecordFill(ctx, fillOf(o, "4", "70000"))
	if err != nil {
		t.Fatal(err)
	}
	if !res.FailClosed {
		t.Fatalf("a shrinking cumulative quantity must fail closed, got %+v", res)
	}
	if got := currentPosition(t, j, o).Quantity; got != "6" {
		t.Errorf("quantity = %s, want the 6 from the last trusted snapshot", got)
	}
}

// TestAnOversellFreezesTheProjectionAndBlocksTheSymbol is 산식 보정 금지 end to
// end: the fill is still recorded, the position is not corrected, and the
// durable RECONCILE state is raised in the same transaction.
func TestAnOversellFreezesTheProjectionAndBlocksTheSymbol(t *testing.T) {
	j := projectingJournal(t)
	ctx := context.Background()

	buy := place(t, j, order{intentID: "i-1", attemptID: "a-1", orderID: "o-1", decisionID: "d-1"})
	if _, err := j.RecordFill(ctx, terminalFill(buy, "10", "70000")); err != nil {
		t.Fatal(err)
	}
	sell := place(t, j, order{intentID: "i-2", attemptID: "a-2", orderID: "o-2",
		side: "SELL", quantity: "14"})

	res, err := j.RecordFill(ctx, terminalFill(sell, "14", "72000"))
	if err != nil {
		t.Fatalf("the fill itself must still be recorded: %v", err)
	}
	if res.Delta != "14" || !res.Changed {
		t.Fatalf("fill result = %+v, want the broker's own snapshot recorded", res)
	}

	frozen := currentPosition(t, j, buy)
	if frozen.Quantity != "10" || frozen.State != PositionOpen {
		t.Errorf("projection = (%s, %s), want the frozen (OPEN, 10)", frozen.State, frozen.Quantity)
	}

	states, err := j.ActiveReconcileStates(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(states) != 1 {
		t.Fatalf("active RECONCILE states = %d, want 1", len(states))
	}
	if states[0].Symbol != buy.symbol || states[0].Cause != ReconcileCauseQuantityMismatch {
		t.Errorf("RECONCILE = (%s, %s), want the symbol scope and QUANTITY_MISMATCH",
			states[0].Symbol, states[0].Cause)
	}
	if !strings.Contains(states[0].Evidence, "X") || !strings.Contains(states[0].Evidence, "OVERSELL") {
		t.Errorf("evidence = %q, want the refused row and its refusal", states[0].Evidence)
	}

	// Re-observing does not stack a second state: the first observation is the
	// one the operator should be looking at.
	if _, err := j.RecordFill(ctx, terminalFill(sell, "15", "72000")); err != nil {
		t.Fatal(err)
	}
	again, err := j.ActiveReconcileStates(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(again) != 1 {
		t.Errorf("active RECONCILE states = %d after a second refusal, want 1", len(again))
	}
}

// TestASellWithNothingHeldIsAttributionFailure separates the two refusals that
// are not arithmetic: nothing is held, so the sell cannot be attributed at all.
func TestASellWithNothingHeldIsAttributionFailure(t *testing.T) {
	j := projectingJournal(t)
	ctx := context.Background()
	sell := place(t, j, order{intentID: "i-1", attemptID: "a-1", orderID: "o-1", side: "SELL"})

	if _, err := j.RecordFill(ctx, terminalFill(sell, "10", "72000")); err != nil {
		t.Fatal(err)
	}
	if _, err := j.CurrentPosition(ctx, sell.account, sell.market, sell.symbol); !errors.Is(err, ErrPositionNotFound) {
		t.Fatalf("a refused sell must not create a position, got %v", err)
	}
	states, err := j.ActiveReconcileStates(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(states) != 1 || states[0].Cause != ReconcileCauseAttributionFailed {
		t.Fatalf("RECONCILE states = %+v, want one ATTRIBUTION_FAILED", states)
	}
}

// TestAnOrderWithNoLocalIntentIsNotProjected pins the direction rule's edge: the
// projection takes the side from the intent, so an order no intent claims is
// left alone rather than projected in a made-up direction.
func TestAnOrderWithNoLocalIntentIsNotProjected(t *testing.T) {
	j := projectingJournal(t)
	ctx := context.Background()

	if _, err := j.RecordFill(ctx, FillObservation{
		OrderID: "outside-1", Symbol: "005930", Market: "kr", State: "CLOSED_FILLED",
		Terminal: true, Quantity: "5", FilledQuantity: "5", AveragePrice: "70000",
		ObservedAt: "2026-03-30T00:30:00Z",
	}); err != nil {
		t.Fatalf("an external order must still be recorded: %v", err)
	}
	if _, err := j.CurrentPosition(ctx, "acct-1", "kr", "005930"); !errors.Is(err, ErrPositionNotFound) {
		t.Fatalf("an order with no intent has no direction and must not be projected, got %v", err)
	}
	states, err := j.ActiveReconcileStates(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(states) != 0 {
		t.Errorf("RECONCILE states = %+v; an unattributable order is the reconciliation's "+
			"comparison to find, not the projection's", states)
	}
}

// TestTheProjectionAndTheFillAreOneCommit is the atomicity requirement seen from
// the projection: a projection that cannot be written takes the fill with it, so
// there is never a fill the position did not see.
func TestTheProjectionAndTheFillAreOneCommit(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()
	boom := errors.New("the projection could not be written")
	if err := j.SetApplyHooks(ApplyHooks{
		Project: func(ctx context.Context, tx *ApplyTx, fill AppliedFill) error {
			if err := ProjectPosition(ctx, tx, fill); err != nil {
				return err
			}
			return boom
		},
	}); err != nil {
		t.Fatal(err)
	}
	o := place(t, j, order{intentID: "i-1", attemptID: "a-1", orderID: "o-1", decisionID: "d-1"})

	if _, err := j.RecordFill(ctx, fillOf(o, "3", "70000")); !errors.Is(err, boom) {
		t.Fatalf("RecordFill = %v, want the projection's error", err)
	}
	if _, err := j.LookupFill(ctx, o.orderID); !errors.Is(err, ErrFillNotFound) {
		t.Errorf("the fill survived a failed projection: %v", err)
	}
	if _, err := j.CurrentPosition(ctx, o.account, o.market, o.symbol); !errors.Is(err, ErrPositionNotFound) {
		t.Errorf("the projection's own write survived the rollback: %v", err)
	}
}

// TestTheProjectionCarriesThePreviousSnapshotForItsCostBasis pins the AppliedFill
// extension task 6.1 needed: without the previous cumulative and average, an
// order that filled in two pieces at two prices would be priced at the second
// piece's running average and the basis would be wrong.
func TestTheProjectionCarriesThePreviousSnapshotForItsCostBasis(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()

	var seen []AppliedFill
	if err := j.SetApplyHooks(ApplyHooks{
		Project: func(ctx context.Context, tx *ApplyTx, fill AppliedFill) error {
			seen = append(seen, fill)
			return ProjectPosition(ctx, tx, fill)
		},
	}); err != nil {
		t.Fatal(err)
	}
	o := place(t, j, order{intentID: "i-1", attemptID: "a-1", orderID: "o-1", decisionID: "d-1"})

	if _, err := j.RecordFill(ctx, fillOf(o, "5", "70000")); err != nil {
		t.Fatal(err)
	}
	// 5 @ 70000 then 5 more @ 80000 → the order's average is 75000, and the
	// marginal 5 cost 80000 each. A projection that priced the delta at the
	// order's average would land on 75000 either way, so the second observation
	// is what separates the exact rule from the approximation.
	if _, err := j.RecordFill(ctx, terminalFill(o, "10", "75000")); err != nil {
		t.Fatal(err)
	}

	if len(seen) != 2 {
		t.Fatalf("hook calls = %d, want 2", len(seen))
	}
	if seen[0].PrevCumulativeQuantity != "" || seen[0].PrevAveragePrice != "" {
		t.Errorf("first observation carried a previous snapshot: %+v", seen[0])
	}
	if seen[1].PrevCumulativeQuantity != "5" || seen[1].PrevAveragePrice != "70000" {
		t.Errorf("second observation's previous = (%q, %q), want (5, 70000)",
			seen[1].PrevCumulativeQuantity, seen[1].PrevAveragePrice)
	}
	if seen[1].OrderedQuantity != "10" {
		t.Errorf("ordered quantity = %q, want the 원주문 수량 the completion is judged against",
			seen[1].OrderedQuantity)
	}
	if got := currentPosition(t, j, o).AvgPrice; got != "75000" {
		t.Errorf("average = %s, want 75000", got)
	}
}

// TestAnUnpricedFillLeavesTheBasisUnknown is the fail-closed direction for a
// missing average price reaching storage: "" and not a made-up zero.
func TestAnUnpricedFillLeavesTheBasisUnknown(t *testing.T) {
	j := projectingJournal(t)
	ctx := context.Background()
	o := place(t, j, order{intentID: "i-1", attemptID: "a-1", orderID: "o-1", decisionID: "d-1"})

	obs := terminalFill(o, "10", "")
	if _, err := j.RecordFill(ctx, obs); err != nil {
		t.Fatal(err)
	}
	p := currentPosition(t, j, o)
	if p.Quantity != "10" {
		t.Errorf("quantity = %s, want 10 — the quantity is known even when the price is not", p.Quantity)
	}
	if p.AvgPrice != position.Unknown {
		t.Errorf("average = %q, want the unknown marker", p.AvgPrice)
	}
}

// TestTheProjectedStatesAreTheSchemasStates ties the two halves together. The
// domain package names the states and the schema's CHECK constraint pins them;
// if the two ever disagreed, every projection write would fail at the constraint
// on a live account instead of here.
func TestTheProjectedStatesAreTheSchemasStates(t *testing.T) {
	t.Parallel()

	for domain, stored := range map[position.State]string{
		position.Flat:    PositionFlat,
		position.Opening: PositionOpening,
		position.Open:    PositionOpen,
		position.Scaling: PositionScaling,
		position.Closing: PositionClosing,
		position.Closed:  PositionClosed,
	} {
		if string(domain) != stored {
			t.Errorf("internal/position says %q where the schema says %q", domain, stored)
		}
	}
}

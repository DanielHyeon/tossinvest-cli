package journal

import (
	"context"
	"errors"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/position"
)

// position_adjustments_test.go is task 6.2: compare-and-append adjustment of the
// projection, the account's value winning, and the two invariants that make the
// compare mean something.

// adjustmentFor builds a well-formed request against one order's venue.
func adjustmentFor(o order, expectedPrev, newQuantity string, watermark int64) AdjustmentRequest {
	return AdjustmentRequest{
		AccountRef: o.account, Market: o.market, Symbol: o.symbol,
		Kind:                  string(position.KindUnknown),
		ExpectedPrevQuantity:  expectedPrev,
		ExpectedFillWatermark: watermark,
		NewQuantity:           newQuantity,
		BrokerAsOf:            "2026-03-30T00:31:00Z",
		Evidence:              "holdings snapshot reported " + newQuantity,
	}
}

// openTenShares projects a straightforward 10-share position to adjust.
func openTenShares(t *testing.T, j *Journal) (*Journal, order) {
	t.Helper()
	o := place(t, j, order{intentID: "i-1", attemptID: "a-1", orderID: "o-1", decisionID: "d-1"})
	if _, err := j.RecordFill(context.Background(), terminalFill(o, "10", "70000")); err != nil {
		t.Fatalf("RecordFill: %v", err)
	}
	return j, o
}

// TestTheAccountsQuantityWinsAndTheProjectionConverges is the spec's own
// scenario: 계좌 보유수량이 로컬보다 적으면 조정 이벤트가 기대 이전 값 검증을
// 통과해 기록되고 투영이 계좌 값으로 수렴한다.
func TestTheAccountsQuantityWinsAndTheProjectionConverges(t *testing.T) {
	j, o := openTenShares(t, projectingJournal(t))
	ctx := context.Background()

	watermark, err := j.FillWatermark(ctx, o.symbol)
	if err != nil {
		t.Fatal(err)
	}
	res, err := j.ApplyPositionAdjustment(ctx, adjustmentFor(o, "10", "7", watermark))
	if err != nil {
		t.Fatalf("ApplyPositionAdjustment: %v", err)
	}
	if !res.Applied || res.OpenedInstance {
		t.Fatalf("result = %+v, want an in-place adjustment", res)
	}
	if res.Position.Quantity != "7" {
		t.Errorf("projection = %s, want the account's 7", res.Position.Quantity)
	}
	if res.Position.State != PositionOpen {
		t.Errorf("state = %s, want the live state kept", res.Position.State)
	}
	if res.Position.AvgPrice != "70000" {
		t.Errorf("average = %s; an adjustment that reports no cost basis must not erase one",
			res.Position.AvgPrice)
	}

	// The event, not just the effect: the row is what justifies the projection
	// disagreeing with the fills.
	events, err := j.PositionAdjustments(ctx, res.Position.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 {
		t.Fatalf("adjustments = %d, want 1", len(events))
	}
	got := events[0]
	if got.ExpectedPrevQuantity != "10" || got.PrevQuantity != "10" || got.NewQuantity != "7" {
		t.Errorf("adjustment quantities = (%s, %s, %s), want (10, 10, 7)",
			got.ExpectedPrevQuantity, got.PrevQuantity, got.NewQuantity)
	}
	if got.BrokerAsOf != "2026-03-30T00:31:00Z" || got.Evidence == "" {
		t.Errorf("adjustment evidence = (%q, %q), want the broker as-of and the observation",
			got.BrokerAsOf, got.Evidence)
	}
}

// TestAnAdjustmentAgainstAMovedQuantityIsDiscarded is the compare half. The
// adjustment was computed against 10; by the time it commits the projection
// holds 6, and applying the difference on top of that would produce a quantity
// nobody observed.
func TestAnAdjustmentAgainstAMovedQuantityIsDiscarded(t *testing.T) {
	j, o := openTenShares(t, projectingJournal(t))
	ctx := context.Background()

	watermark, err := j.FillWatermark(ctx, o.symbol)
	if err != nil {
		t.Fatal(err)
	}
	// A sell lands between the collection and the commit.
	sell := place(t, j, order{intentID: "i-2", attemptID: "a-2", orderID: "o-2", side: "SELL"})
	if _, err := j.RecordFill(ctx, fillOf(sell, "4", "72000")); err != nil {
		t.Fatal(err)
	}

	_, err = j.ApplyPositionAdjustment(ctx, adjustmentFor(o, "10", "7", watermark))
	if !errors.Is(err, ErrAdjustmentStale) {
		t.Fatalf("err = %v, want ErrAdjustmentStale so the caller re-collects", err)
	}
	var stale *StaleAdjustmentError
	if !errors.As(err, &stale) || stale.Invariant != "quantity" {
		t.Fatalf("err = %v, want the quantity invariant named", err)
	}
	if stale.Expected != "10" || stale.Actual != "6" {
		t.Errorf("stale = (%s, %s), want (10, 6)", stale.Expected, stale.Actual)
	}

	// Nothing was written, so a re-collection starts clean.
	if got := currentPosition(t, j, o).Quantity; got != "6" {
		t.Errorf("quantity = %s, want the 6 the fill left", got)
	}
	events, err := j.PositionAdjustments(ctx, currentPosition(t, j, o).ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 0 {
		t.Errorf("a discarded adjustment left %d rows behind", len(events))
	}

	// Re-collecting against the new truth succeeds.
	fresh, err := j.FillWatermark(ctx, o.symbol)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := j.ApplyPositionAdjustment(ctx, adjustmentFor(o, "6", "5", fresh)); err != nil {
		t.Fatalf("the re-collected adjustment must apply: %v", err)
	}
	if got := currentPosition(t, j, o).Quantity; got != "5" {
		t.Errorf("quantity = %s, want 5", got)
	}
}

// TestAMovedFillWatermarkIsDiscardedEvenWhenTheQuantityMatches is why the
// watermark exists at all. A buy and a sell of the same size land between the
// collection and the commit: the quantity is identical and the world is not, and
// the expected-previous check alone would happily apply the adjustment to it.
func TestAMovedFillWatermarkIsDiscardedEvenWhenTheQuantityMatches(t *testing.T) {
	j, o := openTenShares(t, projectingJournal(t))
	ctx := context.Background()

	watermark, err := j.FillWatermark(ctx, o.symbol)
	if err != nil {
		t.Fatal(err)
	}

	sell := place(t, j, order{intentID: "i-2", attemptID: "a-2", orderID: "o-2",
		side: "SELL", quantity: "3"})
	if _, err := j.RecordFill(ctx, terminalFill(sell, "3", "72000")); err != nil {
		t.Fatal(err)
	}
	buy := place(t, j, order{intentID: "i-3", attemptID: "a-3", orderID: "o-3",
		decisionID: "d-2", quantity: "3"})
	if _, err := j.RecordFill(ctx, terminalFill(buy, "3", "73000")); err != nil {
		t.Fatal(err)
	}
	if got := currentPosition(t, j, o).Quantity; got != "10" {
		t.Fatalf("quantity = %s, want the round trip to net back to 10", got)
	}

	_, err = j.ApplyPositionAdjustment(ctx, adjustmentFor(o, "10", "7", watermark))
	if !errors.Is(err, ErrAdjustmentStale) {
		t.Fatalf("err = %v, want the watermark to catch what the quantity could not", err)
	}
	var stale *StaleAdjustmentError
	if !errors.As(err, &stale) || stale.Invariant != "fill watermark" {
		t.Fatalf("err = %v, want the fill watermark invariant named", err)
	}
}

// TestReapplyingTheSameAdjustmentIsANoOp is the crash case: the transaction
// committed and the caller never heard, so it retries. The id is derived from
// what the adjustment is, so the retry is recognised rather than applied twice
// or rejected as stale.
func TestReapplyingTheSameAdjustmentIsANoOp(t *testing.T) {
	j, o := openTenShares(t, projectingJournal(t))
	ctx := context.Background()

	watermark, err := j.FillWatermark(ctx, o.symbol)
	if err != nil {
		t.Fatal(err)
	}
	req := adjustmentFor(o, "10", "7", watermark)
	first, err := j.ApplyPositionAdjustment(ctx, req)
	if err != nil {
		t.Fatal(err)
	}

	second, err := j.ApplyPositionAdjustment(ctx, req)
	if err != nil {
		t.Fatalf("a retry of an applied adjustment must not fail: %v", err)
	}
	if second.Applied {
		t.Error("the retry reported a fresh apply; the adjustment was already on disk")
	}
	if second.Adjustment.ID != first.Adjustment.ID {
		t.Errorf("ids = (%s, %s), want the same derived id", first.Adjustment.ID, second.Adjustment.ID)
	}
	if second.Position.Quantity != "7" {
		t.Errorf("quantity = %s, want the converged 7 rather than a second application",
			second.Position.Quantity)
	}
	events, err := j.PositionAdjustments(ctx, first.Position.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 {
		t.Fatalf("adjustments = %d, want 1: the append-only log records the event once", len(events))
	}
}

// TestAnAdjustmentSurvivesARestart is the durability half of the same claim: the
// row and the converged projection are one commit, so a reopened journal sees
// both.
func TestAnAdjustmentSurvivesARestart(t *testing.T) {
	path := t.TempDir() + "/journal.db"
	j := openTestJournalAt(t, path)
	if err := j.SetApplyHooks(ApplyHooks{Project: ProjectPosition}); err != nil {
		t.Fatal(err)
	}
	_, o := openTenShares(t, j)
	ctx := context.Background()

	watermark, err := j.FillWatermark(ctx, o.symbol)
	if err != nil {
		t.Fatal(err)
	}
	applied, err := j.ApplyPositionAdjustment(ctx, adjustmentFor(o, "10", "7", watermark))
	if err != nil {
		t.Fatal(err)
	}
	if err := j.Close(); err != nil {
		t.Fatal(err)
	}

	reopened := openTestJournalAt(t, path)
	restored, err := reopened.LookupPosition(ctx, applied.Position.ID)
	if err != nil {
		t.Fatal(err)
	}
	if restored.Quantity != "7" {
		t.Errorf("quantity after the restart = %s, want 7", restored.Quantity)
	}
	events, err := reopened.PositionAdjustments(ctx, applied.Position.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 {
		t.Fatalf("adjustments after the restart = %d, want 1", len(events))
	}
}

// TestAnExternalHoldingIsFoldedInWithNoEntryDecision is D4's external position:
// the ledger stops lying about the account, and the row carries NULL where a
// decision would be — so it is not an exit-policy target.
func TestAnExternalHoldingIsFoldedInWithNoEntryDecision(t *testing.T) {
	j := projectingJournal(t)
	ctx := context.Background()

	req := AdjustmentRequest{
		AccountRef: "acct-1", Market: "kr", Symbol: "000660",
		Kind:                  string(position.KindExternal),
		ExpectedPrevQuantity:  "0",
		ExpectedFillWatermark: 0,
		NewQuantity:           "25",
		BrokerAsOf:            "2026-03-30T00:31:00Z",
		Evidence:              "holdings shows 25 of 000660 with no local instance",
	}
	res, err := j.ApplyPositionAdjustment(ctx, req)
	if err != nil {
		t.Fatalf("ApplyPositionAdjustment: %v", err)
	}
	if !res.OpenedInstance {
		t.Fatal("an account holding with no local instance opens one")
	}
	p := res.Position
	if p.Quantity != "25" || p.State != PositionOpen {
		t.Errorf("folded-in position = (%s, %s), want (OPEN, 25)", p.State, p.Quantity)
	}
	if p.EntryDecisionID != "" {
		t.Errorf("entry_decision_id = %q, want NULL: no decision justifies it", p.EntryDecisionID)
	}
	if p.ExitEligible() {
		t.Error("an external position has no entry stop and must not be an exit-policy target")
	}
	if p.AvgPrice != position.Unknown {
		t.Errorf("average = %q, want the unknown marker when the broker reported no cost basis",
			p.AvgPrice)
	}
	if p.InstanceSeq != 1 {
		t.Errorf("instance_seq = %d, want 1", p.InstanceSeq)
	}
}

// TestAnExternalHoldingOnAClosedSymbolOpensTheNextInstance is CLOSED 종결성
// meeting convergence: the account holds something on a symbol the engine has
// finished with, and that is a new instance rather than a reopened one.
func TestAnExternalHoldingOnAClosedSymbolOpensTheNextInstance(t *testing.T) {
	j, o := openTenShares(t, projectingJournal(t))
	ctx := context.Background()

	sell := place(t, j, order{intentID: "i-2", attemptID: "a-2", orderID: "o-2", side: "SELL"})
	if _, err := j.RecordFill(ctx, terminalFill(sell, "10", "72000")); err != nil {
		t.Fatal(err)
	}
	if got := currentPosition(t, j, o).State; got != PositionClosed {
		t.Fatalf("state = %s, want CLOSED before the adjustment", got)
	}

	watermark, err := j.FillWatermark(ctx, o.symbol)
	if err != nil {
		t.Fatal(err)
	}
	req := adjustmentFor(o, "0", "4", watermark)
	req.Kind = string(position.KindExternal)
	res, err := j.ApplyPositionAdjustment(ctx, req)
	if err != nil {
		t.Fatalf("ApplyPositionAdjustment: %v", err)
	}
	if !res.OpenedInstance || res.Position.InstanceSeq != 2 {
		t.Fatalf("result = (%v, %d), want a new instance 2", res.OpenedInstance, res.Position.InstanceSeq)
	}
	if res.Position.EntryDecisionID != "" {
		t.Errorf("entry_decision_id = %q, want NULL", res.Position.EntryDecisionID)
	}

	all, err := j.Positions(ctx, o.account)
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 2 || all[0].State != PositionClosed {
		t.Errorf("instances = %+v, want the closed original preserved beside the new one", all)
	}
}

// TestAnAdjustmentToZeroClosesTheInstance is the convergence rule's other end.
func TestAnAdjustmentToZeroClosesTheInstance(t *testing.T) {
	j, o := openTenShares(t, projectingJournal(t))
	ctx := context.Background()

	watermark, err := j.FillWatermark(ctx, o.symbol)
	if err != nil {
		t.Fatal(err)
	}
	res, err := j.ApplyPositionAdjustment(ctx, adjustmentFor(o, "10", "0", watermark))
	if err != nil {
		t.Fatal(err)
	}
	if res.Position.State != PositionClosed || res.Position.Quantity != "0" {
		t.Fatalf("converged = (%s, %s), want (CLOSED, 0)", res.Position.State, res.Position.Quantity)
	}
	if res.Position.ClosedAt == "" {
		t.Error("closed_at is empty on an instance an adjustment closed")
	}
}

// TestAnAdjustmentConvergesAFrozenProjection is the loop tasks 6.1 and 6.2 make
// together: the transition table refuses an event and freezes the row, and the
// account's own value is what unfreezes it.
func TestAnAdjustmentConvergesAFrozenProjection(t *testing.T) {
	j, o := openTenShares(t, projectingJournal(t))
	ctx := context.Background()

	// An oversell freezes the projection at 10 and raises RECONCILE.
	sell := place(t, j, order{intentID: "i-2", attemptID: "a-2", orderID: "o-2",
		side: "SELL", quantity: "14"})
	if _, err := j.RecordFill(ctx, terminalFill(sell, "14", "72000")); err != nil {
		t.Fatal(err)
	}
	if got := currentPosition(t, j, o).Quantity; got != "10" {
		t.Fatalf("quantity = %s, want the frozen 10", got)
	}

	watermark, err := j.FillWatermark(ctx, o.symbol)
	if err != nil {
		t.Fatal(err)
	}
	req := adjustmentFor(o, "10", "0", watermark)
	req.Evidence = "the account holds none of it; the refused sell of 14 was real"
	res, err := j.ApplyPositionAdjustment(ctx, req)
	if err != nil {
		t.Fatalf("ApplyPositionAdjustment: %v", err)
	}
	if res.Position.Quantity != "0" || res.Position.State != PositionClosed {
		t.Errorf("converged = (%s, %s), want the account's (CLOSED, 0)",
			res.Position.State, res.Position.Quantity)
	}
	// The RECONCILE state stays until something releases it. Releasing on the
	// adjustment alone would clear a block from the same evidence that raised
	// it; the release rule is task 6.3's (ADJUSTMENT_APPLIED).
	states, err := j.ActiveReconcileStates(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(states) != 1 {
		t.Errorf("active RECONCILE states = %d, want the block still standing", len(states))
	}
}

// TestAdjustmentsAreAppendOnly walks two adjustments over one position and
// checks the history is both of them: the sequence is the justification, and a
// row that could be overwritten would erase it.
func TestAdjustmentsAreAppendOnly(t *testing.T) {
	j, o := openTenShares(t, projectingJournal(t))
	ctx := context.Background()

	watermark, err := j.FillWatermark(ctx, o.symbol)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := j.ApplyPositionAdjustment(ctx, adjustmentFor(o, "10", "8", watermark)); err != nil {
		t.Fatal(err)
	}
	second := adjustmentFor(o, "8", "6", watermark)
	second.BrokerAsOf = "2026-03-30T00:32:00Z"
	res, err := j.ApplyPositionAdjustment(ctx, second)
	if err != nil {
		t.Fatal(err)
	}

	events, err := j.PositionAdjustments(ctx, res.Position.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 2 {
		t.Fatalf("adjustments = %d, want both", len(events))
	}
	if events[0].NewQuantity != "8" || events[1].NewQuantity != "6" {
		t.Errorf("history = (%s, %s), want (8, 6) in order",
			events[0].NewQuantity, events[1].NewQuantity)
	}
	if events[1].PrevQuantity != "8" {
		t.Errorf("the second adjustment's prev = %s, want the 8 the first left", events[1].PrevQuantity)
	}
}

// TestANewCostBasisFromTheAccountReplacesTheProjection covers the price half:
// when the broker does report a cost basis, the account's value wins there too.
func TestANewCostBasisFromTheAccountReplacesTheProjection(t *testing.T) {
	j, o := openTenShares(t, projectingJournal(t))
	ctx := context.Background()

	watermark, err := j.FillWatermark(ctx, o.symbol)
	if err != nil {
		t.Fatal(err)
	}
	req := adjustmentFor(o, "10", "10", watermark)
	req.NewAvgPrice = "69500"
	req.Kind = string(position.KindManual)
	res, err := j.ApplyPositionAdjustment(ctx, req)
	if err != nil {
		t.Fatal(err)
	}
	if res.Position.AvgPrice != "69500" {
		t.Errorf("average = %s, want the account's 69500", res.Position.AvgPrice)
	}
	events, err := j.PositionAdjustments(ctx, res.Position.ID)
	if err != nil {
		t.Fatal(err)
	}
	if events[0].PrevAvgPrice != "70000" || events[0].NewAvgPrice != "69500" {
		t.Errorf("recorded prices = (%s, %s), want (70000, 69500)",
			events[0].PrevAvgPrice, events[0].NewAvgPrice)
	}
	if events[0].Kind != string(position.KindManual) {
		t.Errorf("kind = %s, want MANUAL", events[0].Kind)
	}
}

// TestAdjustmentRequestsAreValidatedBeforeAnythingIsWritten checks the refusals
// that keep the audit trail meaningful: an unclassified difference, one with no
// evidence, one with no broker as-of, and a negative quantity.
func TestAdjustmentRequestsAreValidatedBeforeAnythingIsWritten(t *testing.T) {
	j, o := openTenShares(t, projectingJournal(t))
	ctx := context.Background()
	watermark, err := j.FillWatermark(ctx, o.symbol)
	if err != nil {
		t.Fatal(err)
	}

	mutate := map[string]func(*AdjustmentRequest){
		"an unclassified difference":       func(r *AdjustmentRequest) { r.Kind = "PROBABLY_EXTERNAL" },
		"no classification at all":         func(r *AdjustmentRequest) { r.Kind = "" },
		"no evidence":                      func(r *AdjustmentRequest) { r.Evidence = "" },
		"no broker as-of":                  func(r *AdjustmentRequest) { r.BrokerAsOf = "" },
		"no account":                       func(r *AdjustmentRequest) { r.AccountRef = "" },
		"no symbol":                        func(r *AdjustmentRequest) { r.Symbol = "" },
		"a negative quantity":              func(r *AdjustmentRequest) { r.NewQuantity = "-1" },
		"a quantity that is not a decimal": func(r *AdjustmentRequest) { r.NewQuantity = "lots" },
		"a negative watermark":             func(r *AdjustmentRequest) { r.ExpectedFillWatermark = -1 },
	}
	for name, break_ := range mutate {
		req := adjustmentFor(o, "10", "7", watermark)
		break_(&req)
		if _, err := j.ApplyPositionAdjustment(ctx, req); !errors.Is(err, ErrInvalidRequest) {
			t.Errorf("%s: err = %v, want ErrInvalidRequest", name, err)
		}
	}
	if got := currentPosition(t, j, o).Quantity; got != "10" {
		t.Errorf("quantity = %s; a refused request must write nothing", got)
	}
}

// TestClassifyNamesTheProvenance pins the three kinds apart. They are three
// different facts and the audit trail has to keep them so.
func TestClassifyNamesTheProvenance(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		in   position.ProvenanceInputs
		want position.Kind
	}{
		{"a person said so", position.ProvenanceInputs{LocalInstance: true, OperatorDeclared: true},
			position.KindManual},
		{"a person said so on a symbol we never traded",
			position.ProvenanceInputs{OperatorDeclared: true}, position.KindManual},
		{"a symbol the engine never traded", position.ProvenanceInputs{}, position.KindExternal},
		{"a difference on a symbol the engine holds",
			position.ProvenanceInputs{LocalInstance: true}, position.KindUnknown},
	}
	for _, tc := range cases {
		if got := position.Classify(tc.in); got != tc.want {
			t.Errorf("%s: Classify = %s, want %s", tc.name, got, tc.want)
		}
	}
}

// TestTheAdjustmentKindsAreTheSchemasKinds is the cross-check the projection
// test does for states: the CHECK constraint and the domain enum must be the
// same three values.
func TestTheAdjustmentKindsAreTheSchemasKinds(t *testing.T) {
	t.Parallel()

	for domain, stored := range map[position.Kind]string{
		position.KindExternal: AdjustmentExternal,
		position.KindManual:   AdjustmentManual,
		position.KindUnknown:  AdjustmentUnknown,
	} {
		if string(domain) != stored {
			t.Errorf("internal/position says %q where the schema says %q", domain, stored)
		}
	}
}

// TestTheFillWatermarkIsPerSymbol pins the scope. An account-wide watermark
// would discard an adjustment for a symbol nothing happened on, every time
// anything filled anywhere.
func TestTheFillWatermarkIsPerSymbol(t *testing.T) {
	j, o := openTenShares(t, projectingJournal(t))
	ctx := context.Background()

	quiet, err := j.FillWatermark(ctx, "000660")
	if err != nil {
		t.Fatal(err)
	}
	if quiet != 0 {
		t.Errorf("watermark of an untraded symbol = %d, want 0", quiet)
	}
	busy, err := j.FillWatermark(ctx, o.symbol)
	if err != nil {
		t.Fatal(err)
	}
	if busy == 0 {
		t.Fatal("watermark of a filled symbol is 0")
	}

	// A fill elsewhere does not invalidate an adjustment here.
	other := place(t, j, order{intentID: "i-9", attemptID: "a-9", orderID: "o-9",
		decisionID: "d-9", symbol: "000660", quantity: "2"})
	if _, err := j.RecordFill(ctx, terminalFill(other, "2", "120000")); err != nil {
		t.Fatal(err)
	}
	if _, err := j.ApplyPositionAdjustment(ctx, adjustmentFor(o, "10", "9", busy)); err != nil {
		t.Fatalf("an unrelated symbol's fill must not discard this adjustment: %v", err)
	}
}

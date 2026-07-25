package journal

import (
	"context"
	"path/filepath"
	"testing"
)

// Fill ledger tests (harden-execution-base task 3.2).
//
// The requirement being pinned is the fill-detection spec's "누적 스냅샷 기반 멱등
// 반영": only a positive delta is a new fill, a shrinking or out-of-order
// cumulative quantity fails closed, and the same snapshot arriving twice — from
// the poll loop, from an SSE re-fetch, or on the first cycle after a restart —
// changes nothing.

func observation(orderID, filled string) FillObservation {
	return FillObservation{
		OrderID:        orderID,
		Symbol:         "AAPL",
		Market:         "us",
		State:          "OPEN_PARTIALLY_FILLED",
		Quantity:       "10",
		FilledQuantity: filled,
		AveragePrice:   "213.4",
		ObservedAt:     "2026-03-30T00:30:00Z",
	}
}

// TestOnlyPositiveDeltasBecomeFills is the core rule: the broker reports a
// cumulative number, so a fill is a difference, and the difference is computed
// once, inside the transaction that records it.
func TestOnlyPositiveDeltasBecomeFills(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()

	first, err := j.RecordFill(ctx, observation("o-1", "3"))
	if err != nil {
		t.Fatalf("first snapshot: %v", err)
	}
	if first.Delta != "3" || !first.Changed || first.FailClosed {
		t.Fatalf("first snapshot = %+v, want a delta of 3", first)
	}

	second, err := j.RecordFill(ctx, observation("o-1", "7"))
	if err != nil {
		t.Fatalf("second snapshot: %v", err)
	}
	if second.Delta != "4" {
		t.Fatalf("second delta = %q, want 4 (7 cumulative minus the 3 already known)", second.Delta)
	}

	events, err := j.FillEvents(ctx, "o-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 2 {
		t.Fatalf("fill events = %d, want 2", len(events))
	}
	if events[0].DeltaQuantity != "3" || events[1].DeltaQuantity != "4" {
		t.Fatalf("event deltas = %q, %q; want 3 then 4",
			events[0].DeltaQuantity, events[1].DeltaQuantity)
	}
	if events[1].CumulativeQuantity != "7" {
		t.Fatalf("event cumulative = %q, want 7", events[1].CumulativeQuantity)
	}
}

// TestIdenticalSnapshotChangesNothing is the spec's "동일 스냅샷 중복 수신"
// scenario. Not "is harmless" — changes nothing: no event, no row write.
func TestIdenticalSnapshotChangesNothing(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()

	if _, err := j.RecordFill(ctx, observation("o-1", "3")); err != nil {
		t.Fatal(err)
	}
	before, err := j.LookupFill(ctx, "o-1")
	if err != nil {
		t.Fatal(err)
	}

	// The same snapshot arriving again, this time from an SSE-triggered re-fetch.
	repeat := observation("o-1", "3")
	repeat.ObservedAt = "2026-03-30T00:30:05Z"
	again, err := j.RecordFill(ctx, repeat)
	if err != nil {
		t.Fatalf("repeat snapshot: %v", err)
	}
	if again.Changed || again.Delta != "0" || again.FailClosed {
		t.Fatalf("repeat snapshot = %+v, want a no-op", again)
	}

	after, err := j.LookupFill(ctx, "o-1")
	if err != nil {
		t.Fatal(err)
	}
	if after != before {
		t.Fatalf("the stored row moved on a duplicate:\n before %+v\n after  %+v", before, after)
	}
	events, err := j.FillEvents(ctx, "o-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 {
		t.Fatalf("fill events = %d, want the single original fill", len(events))
	}
}

// TestDuplicateSurvivesRestart covers the "재시작 후" half of the same
// requirement: the last observation is on disk, so a fresh process does not
// mistake the whole cumulative quantity for a new fill.
func TestDuplicateSurvivesRestart(t *testing.T) {
	path := filepath.Join(t.TempDir(), "journal.db")
	ctx := context.Background()

	first := openTestJournalAt(t, path)
	if _, err := first.RecordFill(ctx, observation("o-1", "5")); err != nil {
		t.Fatal(err)
	}
	if err := first.Close(); err != nil {
		t.Fatal(err)
	}

	second := openTestJournalAt(t, path)
	res, err := second.RecordFill(ctx, observation("o-1", "5"))
	if err != nil {
		t.Fatalf("post-restart snapshot: %v", err)
	}
	if res.Changed || res.Delta != "0" {
		t.Fatalf("post-restart snapshot = %+v, want a no-op", res)
	}
	events, err := second.FillEvents(ctx, "o-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 {
		t.Fatalf("fill events after restart = %d, want 1", len(events))
	}
}

// TestShrinkingCumulativeQuantityFailsClosed is the spec's "filledQuantity 감소
// 관측" scenario. There is no safe correction: either the broker is describing a
// different order or our record is wrong, and both mean stop.
func TestShrinkingCumulativeQuantityFailsClosed(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()

	if _, err := j.RecordFill(ctx, observation("o-1", "7")); err != nil {
		t.Fatal(err)
	}
	res, err := j.RecordFill(ctx, observation("o-1", "3"))
	if err != nil {
		t.Fatalf("shrinking snapshot: %v", err)
	}
	if !res.FailClosed {
		t.Fatalf("a shrinking cumulative quantity must fail closed, got %+v", res)
	}
	if res.Reason != ReasonFillDecreased {
		t.Fatalf("reason = %q, want %q", res.Reason, ReasonFillDecreased)
	}

	stored, err := j.LookupFill(ctx, "o-1")
	if err != nil {
		t.Fatal(err)
	}
	if stored.FilledQuantity != "7" {
		t.Fatalf("stored quantity = %q, want the last trusted 7 — a refusal must not advance it",
			stored.FilledQuantity)
	}
	if !stored.FailClosed || stored.ReasonCode != ReasonFillDecreased {
		t.Fatalf("the refusal must be durable and explained, got %+v", stored)
	}
	events, err := j.FillEvents(ctx, "o-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 {
		t.Fatalf("fill events = %d, want only the original fill", len(events))
	}
}

// TestOutOfOrderSnapshotFailsClosed: a snapshot the broker timestamped earlier
// than the one already recorded cannot also carry a different quantity. Accepting
// it would let a late-arriving stale read rewrite a newer truth.
func TestOutOfOrderSnapshotFailsClosed(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()

	newer := observation("o-1", "5")
	newer.BrokerVisibleAt = "2026-03-30T00:30:00Z"
	if _, err := j.RecordFill(ctx, newer); err != nil {
		t.Fatal(err)
	}

	older := observation("o-1", "8")
	older.BrokerVisibleAt = "2026-03-30T00:29:00Z"
	res, err := j.RecordFill(ctx, older)
	if err != nil {
		t.Fatalf("out-of-order snapshot: %v", err)
	}
	if !res.FailClosed || res.Reason != ReasonFillOutOfOrder {
		t.Fatalf("an older snapshot with a different quantity must fail closed, got %+v", res)
	}
	stored, err := j.LookupFill(ctx, "o-1")
	if err != nil {
		t.Fatal(err)
	}
	if stored.FilledQuantity != "5" {
		t.Fatalf("stored quantity = %q, want 5", stored.FilledQuantity)
	}
}

// TestOlderTimestampWithTheSameQuantityIsJustAReRead keeps the out-of-order rule
// from firing on the common case of two reads of one unchanged state.
func TestOlderTimestampWithTheSameQuantityIsJustAReRead(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()

	newer := observation("o-1", "5")
	newer.BrokerVisibleAt = "2026-03-30T00:30:00Z"
	if _, err := j.RecordFill(ctx, newer); err != nil {
		t.Fatal(err)
	}
	older := observation("o-1", "5")
	older.BrokerVisibleAt = "2026-03-30T00:29:00Z"
	res, err := j.RecordFill(ctx, older)
	if err != nil {
		t.Fatal(err)
	}
	if res.FailClosed {
		t.Fatalf("an unchanged quantity is a re-read, not a contradiction: %+v", res)
	}
}

// TestCallerFailClosedIsRecordedWithoutAdvancing lets the derivation's own
// UNKNOWN_BROKER_STATE verdict stop the snapshot, so an order whose state the
// priority table refuses cannot quietly keep accumulating fills.
func TestCallerFailClosedIsRecordedWithoutAdvancing(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()

	if _, err := j.RecordFill(ctx, observation("o-1", "4")); err != nil {
		t.Fatal(err)
	}
	refused := observation("o-1", "9")
	refused.FailClosed = true
	refused.Reason = "unknown_status"
	refused.Detail = `broker status "SETTLING" is not one this build understands`

	res, err := j.RecordFill(ctx, refused)
	if err != nil {
		t.Fatal(err)
	}
	if !res.FailClosed || res.Reason != "unknown_status" {
		t.Fatalf("the caller's verdict must be honoured, got %+v", res)
	}
	stored, err := j.LookupFill(ctx, "o-1")
	if err != nil {
		t.Fatal(err)
	}
	if stored.FilledQuantity != "4" {
		t.Fatalf("stored quantity = %q, want 4", stored.FilledQuantity)
	}
	if stored.Detail != refused.Detail {
		t.Fatalf("detail = %q, want the derivation's explanation", stored.Detail)
	}
}

// TestFirstObservationThatFailsClosedIsStillVisible: an order whose very first
// snapshot is untrustworthy must not be invisible until it behaves.
func TestFirstObservationThatFailsClosedIsStillVisible(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()

	obs := observation("o-new", "2")
	obs.FailClosed = true
	obs.Reason = "unknown_status"
	if _, err := j.RecordFill(ctx, obs); err != nil {
		t.Fatal(err)
	}
	stored, err := j.LookupFill(ctx, "o-new")
	if err != nil {
		t.Fatalf("a refused first observation must still be recorded: %v", err)
	}
	if !stored.FailClosed || stored.FilledQuantity != "0" {
		t.Fatalf("stored = %+v, want a fail-closed row with no trusted quantity", stored)
	}
}

// TestAveragePriceIsReplacedNotAccumulated: the average is a property of the
// whole filled quantity, so a changed average with an unchanged quantity is a
// correction, not a fill.
func TestAveragePriceIsReplacedNotAccumulated(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()

	if _, err := j.RecordFill(ctx, observation("o-1", "3")); err != nil {
		t.Fatal(err)
	}
	corrected := observation("o-1", "3")
	corrected.AveragePrice = "214.05"
	res, err := j.RecordFill(ctx, corrected)
	if err != nil {
		t.Fatal(err)
	}
	if res.Delta != "0" {
		t.Fatalf("delta = %q, want 0 — an average-price change is not a fill", res.Delta)
	}
	stored, err := j.LookupFill(ctx, "o-1")
	if err != nil {
		t.Fatal(err)
	}
	if stored.AveragePrice != "214.05" {
		t.Fatalf("average price = %q, want the replacement", stored.AveragePrice)
	}
	events, err := j.FillEvents(ctx, "o-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 {
		t.Fatalf("fill events = %d, want 1", len(events))
	}
}

// TestFractionalFillsUseTheDecimalTolerance: US fractional shares arrive as
// decimal strings and are compared as float64, so an exact == would report a
// discrepancy where there is none.
func TestFractionalFillsUseTheDecimalTolerance(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()

	if _, err := j.RecordFill(ctx, observation("o-1", "0.1")); err != nil {
		t.Fatal(err)
	}
	res, err := j.RecordFill(ctx, observation("o-1", "0.30000000000000004"))
	if err != nil {
		t.Fatal(err)
	}
	if res.FailClosed {
		t.Fatalf("a float-representation wobble must not fail closed: %+v", res)
	}
	if res.DeltaQuantity < 0.19 || res.DeltaQuantity > 0.21 {
		t.Fatalf("delta = %v, want ~0.2", res.DeltaQuantity)
	}
}

// TestTerminalSnapshotsLeaveTheTrackedSet stops the detector reading an order
// forever once the broker can no longer change it.
func TestTrackedFillOrders(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()

	live := observation("o-open", "2")
	if _, err := j.RecordFill(ctx, live); err != nil {
		t.Fatal(err)
	}
	done := observation("o-done", "10")
	done.Terminal = true
	done.State = "FILLED"
	if _, err := j.RecordFill(ctx, done); err != nil {
		t.Fatal(err)
	}

	tracked, err := j.TrackedFillOrders(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(tracked) != 1 || tracked[0].OrderID != "o-open" {
		t.Fatalf("tracked = %+v, want only the live order", tracked)
	}
	if tracked[0].Symbol != "AAPL" || tracked[0].Market != "us" {
		t.Fatalf("tracked order lost its identity: %+v", tracked[0])
	}
}

// TestTrackedFillOrdersIncludesConfirmedAttemptsNotYetObserved is the gap that
// matters most: an order acknowledged and filled before the first poll appears in
// neither the open list nor the snapshot table, and would be tracked by nothing.
func TestTrackedFillOrdersIncludesConfirmedAttemptsNotYetObserved(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()

	attempt, err := j.Prepare(ctx, PrepareRequest{
		Intent: Intent{
			ID: "intent-1", Market: "us", TradingDay: "2026-03-30", AccountRef: "acct-1",
			Symbol: "AAPL", Side: "BUY", OrderType: "LIMIT", Quantity: "10",
			Price: "200", Currency: "USD", Source: "engine", Fingerprint: "fp-1",
		},
		Kind: KindPlace, AttemptID: "attempt-1",
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := attempt.MarkDispatchStarted(ctx); err != nil {
		t.Fatal(err)
	}
	if err := attempt.MarkAcked(ctx, "broker-9"); err != nil {
		t.Fatal(err)
	}
	if err := attempt.Settle(ctx, StateConfirmed, ReasonBrokerAcknowledged, "acked"); err != nil {
		t.Fatal(err)
	}

	tracked, err := j.TrackedFillOrders(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(tracked) != 1 || tracked[0].OrderID != "broker-9" {
		t.Fatalf("tracked = %+v, want the confirmed order that has never been polled", tracked)
	}
	if tracked[0].Symbol != "AAPL" {
		t.Fatalf("tracked order = %+v, want the intent's symbol", tracked[0])
	}
}

// TestTrackedFillOrdersCarryLineage keeps an amended order's successor attached,
// so the state table can tell a replace from a cancellation.
func TestTrackedFillOrdersCarryLineage(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()

	if _, err := j.RecordFill(ctx, observation("parent-1", "2")); err != nil {
		t.Fatal(err)
	}
	insertIntent(t, j, "intent-1")
	insertAttempt(t, j, "attempt-1", "intent-1")
	if _, err := j.db.ExecContext(ctx,
		`INSERT INTO lineage_edges
		   (parent_order_id, child_order_id, relation, parent_filled_quantity,
		    requested_quantity, intent_id, attempt_id, created_at)
		 VALUES ('parent-1','child-1','replaces','2','8','intent-1','attempt-1',
		         '2026-03-30T00:30:00Z')`); err != nil {
		t.Fatal(err)
	}

	tracked, err := j.TrackedFillOrders(ctx)
	if err != nil {
		t.Fatal(err)
	}
	var found bool
	for _, tr := range tracked {
		if tr.OrderID == "parent-1" {
			found = true
			if tr.SuccessorOrderID != "child-1" {
				t.Fatalf("successor = %q, want child-1", tr.SuccessorOrderID)
			}
		}
	}
	if !found {
		t.Fatalf("tracked = %+v, want parent-1", tracked)
	}
}

// TestFilledQuantitiesAggregatesPerSymbol feeds the reconciliation comparison.
func TestFilledQuantitiesAggregatesPerSymbol(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()

	if _, err := j.RecordFill(ctx, observation("o-1", "3")); err != nil {
		t.Fatal(err)
	}
	second := observation("o-2", "4")
	if _, err := j.RecordFill(ctx, second); err != nil {
		t.Fatal(err)
	}
	other := observation("o-3", "6")
	other.Symbol = "MSFT"
	if _, err := j.RecordFill(ctx, other); err != nil {
		t.Fatal(err)
	}

	totals, err := j.FilledQuantities(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if totals["AAPL"] != "7" {
		t.Fatalf("AAPL total = %q, want 7", totals["AAPL"])
	}
	if totals["MSFT"] != "6" {
		t.Fatalf("MSFT total = %q, want 6", totals["MSFT"])
	}
}

// TestRecordFillRejectsAnUnreadableQuantity refuses to store a number it cannot
// compare later.
func TestRecordFillRejectsAnUnreadableQuantity(t *testing.T) {
	j := openTestJournal(t)
	if _, err := j.RecordFill(context.Background(), observation("o-1", "many")); err == nil {
		t.Fatal("a non-decimal filled quantity must be refused")
	}
}

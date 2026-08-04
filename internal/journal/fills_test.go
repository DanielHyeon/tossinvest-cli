package journal

import (
	"context"
	"errors"
	"fmt"
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
		AccountRef:     "acct-1",
		Symbol:         "AAPL",
		Market:         "us",
		TradingDay:     "2026-03-30",
		Side:           "BUY",
		State:          "OPEN_PARTIALLY_FILLED",
		Quantity:       "10",
		FilledQuantity: filled,
		AveragePrice:   "213.4",
		ObservedAt:     "2026-03-30T00:30:00Z",
	}
}

func TestFillEventsRequireCanonicalScopeWhenOrderIDIsReused(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()
	firstScope := FillSnapshotScope{
		OrderID: "reused-order", AccountRef: "acct-1", Market: "us",
		TradingDay: "2026-03-30", Symbol: "AAPL", Side: "BUY",
	}
	secondScope := firstScope
	secondScope.TradingDay = "2026-03-31"
	recordConfirmedFillOrderScope(t, j, "reused-fill-first", "reused-fill-attempt-first",
		"reused-order", firstScope)
	recordConfirmedFillOrderScope(t, j, "reused-fill-second", "reused-fill-attempt-second",
		"reused-order", secondScope)

	first := observation("reused-order", "3")
	if _, err := j.RecordFill(ctx, first); err != nil {
		t.Fatal(err)
	}
	second := observation("reused-order", "2")
	second.TradingDay = "2026-03-31"
	second.AveragePrice = "220"
	if _, err := j.RecordFill(ctx, second); err != nil {
		t.Fatal(err)
	}

	if _, err := j.FillEvents(ctx, "reused-order"); !errors.Is(err, ErrFillScopeAmbiguous) {
		t.Fatalf("FillEvents(order only) err=%v, want ErrFillScopeAmbiguous", err)
	}
	got, err := j.FillEventsScoped(ctx, secondScope)
	if err != nil {
		t.Fatalf("FillEventsScoped: %v", err)
	}
	if len(got) != 1 || got[0].CumulativeQuantity != "2" || got[0].TradingDay != "2026-03-31" {
		t.Fatalf("scoped fills=%+v, want only the later trading-day event", got)
	}
}

func TestFillEventsScopedRequiresPreexistingUniqueConfirmedOwner(t *testing.T) {
	for _, tc := range []struct {
		name       string
		ownerCount int
		committed  string
	}{
		{name: "future owner in the same second", ownerCount: 1, committed: "2026-03-30T00:30:00Z"},
		{name: "two owners", ownerCount: 2, committed: "2026-03-30T00:31:00Z"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			j := openTestJournal(t)
			ctx := context.Background()
			scope := FillSnapshotScope{OrderID: "scoped-owner-boundary", AccountRef: "acct-1",
				Market: "us", TradingDay: "2026-03-30", Symbol: "AAPL", Side: "BUY"}
			if _, err := j.db.ExecContext(ctx, `INSERT INTO fill_events
				(order_id, account_ref, symbol, market, trading_day, side, delta_quantity,
				 cumulative_quantity, average_price, broker_visible_at, committed_at)
				VALUES (?,?,?,?,?,?,?,?,?,?,?)`, scope.OrderID, scope.AccountRef, scope.Symbol,
				scope.Market, scope.TradingDay, scope.Side, "1", "1", "200", "", tc.committed); err != nil {
				t.Fatal(err)
			}
			for i := 0; i < tc.ownerCount; i++ {
				recordConfirmedFillOrderScope(t, j, fmt.Sprintf("scoped-owner-%d", i),
					fmt.Sprintf("scoped-owner-attempt-%d", i), scope.OrderID, scope)
			}
			events, err := j.FillEventsScoped(ctx, scope)
			if err != nil || len(events) != 0 {
				t.Fatalf("scoped events=%+v err=%v, want unsafe evidence excluded", events, err)
			}
		})
	}
}

func recordConfirmedFillOrder(t *testing.T, j *Journal, intentID, attemptID, orderID string) {
	recordConfirmedFillOrderScope(t, j, intentID, attemptID, orderID, FillSnapshotScope{
		AccountRef: "acct-1", Market: "us", TradingDay: "2026-03-30", Symbol: "AAPL", Side: "BUY",
	})
}

func recordConfirmedFillOrderScope(t *testing.T, j *Journal, intentID, attemptID, orderID string,
	scope FillSnapshotScope) {
	t.Helper()
	ctx := context.Background()
	attempt, err := j.Prepare(ctx, PrepareRequest{
		Intent: Intent{
			ID: intentID, Market: scope.Market, TradingDay: scope.TradingDay, AccountRef: scope.AccountRef,
			Symbol: scope.Symbol, Side: scope.Side, OrderType: "LIMIT", Quantity: "10",
			Price: "200", Currency: "USD", Source: "engine", Fingerprint: "fp-" + intentID,
		},
		Kind: KindPlace, AttemptID: attemptID,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := attempt.MarkDispatchStarted(ctx); err != nil {
		t.Fatal(err)
	}
	if err := attempt.MarkAcked(ctx, orderID); err != nil {
		t.Fatal(err)
	}
	if err := attempt.Settle(ctx, StateConfirmed, ReasonBrokerAcknowledged, "acked"); err != nil {
		t.Fatal(err)
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

	recordConfirmedFillOrder(t, j, "intent-open", "attempt-open", "o-open")
	live := observation("o-open", "2")
	if _, err := j.RecordFill(ctx, live); err != nil {
		t.Fatal(err)
	}
	recordConfirmedFillOrder(t, j, "intent-done", "attempt-done", "o-done")
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

// TestTrackedFillOrdersExcludeExternalSnapshots pins the ownership boundary:
// storing a broker observation is evidence that the order existed, not evidence
// that this engine submitted it.
func TestTrackedFillOrdersExcludeExternalSnapshots(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()

	if _, err := j.RecordFill(ctx, observation("external-open", "0")); err != nil {
		t.Fatal(err)
	}
	if _, err := j.LookupFill(ctx, "external-open"); err != nil {
		t.Fatalf("external observation must remain durably readable: %v", err)
	}

	tracked, err := j.TrackedFillOrders(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(tracked) != 0 {
		t.Fatalf("tracked = %+v, want no locally owned order", tracked)
	}
}

func TestTrackedFillOrdersRejectSnapshotIdentityMismatch(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()
	recordConfirmedFillOrder(t, j, "intent-owned", "attempt-owned", "shared-order")
	mismatched := observation("shared-order", "0")
	mismatched.Symbol = "MSFT"
	if _, err := j.RecordFill(ctx, mismatched); err != nil {
		t.Fatal(err)
	}

	tracked, err := j.TrackedFillOrders(ctx, "acct-1")
	if !errors.Is(err, ErrTrackedFillIdentityConflict) || tracked != nil {
		t.Fatalf("tracked = %+v err = %v, want a durable identity-conflict refusal", tracked, err)
	}
	active, err := j.ActiveReconcileStates(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(active) != 1 || active[0].Cause != ReconcileCauseIdentifierConflict || active[0].Symbol != "" {
		t.Fatalf("active reconcile states = %+v, want account-wide identifier conflict", active)
	}
}

func TestTrackedFillOrdersScopeReusedOrderIDsByAccount(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()
	recordConfirmedFillOrder(t, j, "intent-one", "attempt-one", "account-scoped-id")
	attempt, err := j.Prepare(ctx, PrepareRequest{
		Intent: Intent{
			ID: "intent-two", Market: "us", TradingDay: "2026-03-30", AccountRef: "acct-2",
			Symbol: "AAPL", Side: "BUY", OrderType: "LIMIT", Quantity: "10",
			Price: "200", Currency: "USD", Source: "engine", Fingerprint: "fp-two",
		},
		Kind: KindPlace, AttemptID: "attempt-two",
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := attempt.MarkDispatchStarted(ctx); err != nil {
		t.Fatal(err)
	}
	if err := attempt.MarkAcked(ctx, "account-scoped-id"); err != nil {
		t.Fatal(err)
	}
	if err := attempt.Settle(ctx, StateConfirmed, ReasonBrokerAcknowledged, "acked"); err != nil {
		t.Fatal(err)
	}

	for _, account := range []string{"acct-1", "acct-2"} {
		tracked, err := j.TrackedFillOrders(ctx, account)
		if err != nil {
			t.Fatalf("account %s: %v", account, err)
		}
		if len(tracked) != 1 || tracked[0].OrderID != "account-scoped-id" {
			t.Fatalf("account %s tracked=%+v, want its scoped order", account, tracked)
		}
	}
}

func TestTrackedFillOrdersSelectLatestTradingDayWhenAnAccountReusesAnOrderID(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()
	confirm := func(intentID, attemptID, day, symbol string) {
		attempt, err := j.Prepare(ctx, PrepareRequest{
			Intent: Intent{
				ID: intentID, Market: "us", TradingDay: day, AccountRef: "acct-1",
				Symbol: symbol, Side: "BUY", OrderType: "LIMIT", Quantity: "1",
				Price: "100", Currency: "USD", Source: "engine", Fingerprint: "fp-" + intentID,
			},
			Kind: KindPlace, AttemptID: attemptID,
		})
		if err != nil {
			t.Fatal(err)
		}
		if err := attempt.MarkDispatchStarted(ctx); err != nil {
			t.Fatal(err)
		}
		if err := attempt.MarkAcked(ctx, "reused-by-day"); err != nil {
			t.Fatal(err)
		}
		if err := attempt.Settle(ctx, StateConfirmed, ReasonBrokerAcknowledged, "acked"); err != nil {
			t.Fatal(err)
		}
	}
	confirm("old-day", "old-attempt", "2026-03-29", "MSFT")
	prior := observation("reused-by-day", "1")
	prior.AccountRef, prior.TradingDay, prior.Side = "acct-1", "2026-03-29", "BUY"
	prior.Symbol, prior.Market, prior.Terminal, prior.State = "MSFT", "us", true, "FILLED"
	if _, err := j.RecordFill(ctx, prior); err != nil {
		t.Fatal(err)
	}
	confirm("new-day", "new-attempt", "2026-03-30", "AAPL")

	tracked, err := j.TrackedFillOrders(ctx, "acct-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(tracked) != 1 || tracked[0].OrderID != "reused-by-day" || tracked[0].Symbol != "AAPL" {
		t.Fatalf("tracked = %+v, want only the latest trading-day identity", tracked)
	}
	current := observation("reused-by-day", "1")
	current.AccountRef, current.TradingDay, current.Side = "acct-1", "2026-03-30", "BUY"
	res, err := j.RecordFill(ctx, current)
	if err != nil {
		t.Fatal(err)
	}
	if res.FailClosed || res.Delta != "1" {
		t.Fatalf("new trading-day cumulative sequence = %+v, want fresh delta 1", res)
	}
	stored, err := j.LookupFillScoped(ctx, FillSnapshotScope{
		OrderID: "reused-by-day", AccountRef: "acct-1", Market: "us",
		TradingDay: "2026-03-30", Symbol: "AAPL", Side: "BUY",
	})
	if err != nil {
		t.Fatal(err)
	}
	if stored.TradingDay != "2026-03-30" || stored.Symbol != "AAPL" {
		t.Fatalf("stored snapshot = %+v, want current trading-day scope", stored)
	}
}

func TestTrackedAndLiveOrdersKeepTwoNonterminalReusedTradingDays(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()
	confirm := func(intentID, attemptID, day string) {
		attempt, err := j.Prepare(ctx, PrepareRequest{
			Intent: Intent{
				ID: intentID, Market: "us", TradingDay: day, AccountRef: "acct-1",
				Symbol: "AAPL", Side: "BUY", OrderType: "LIMIT", Quantity: "1",
				Price: "100", Currency: "USD", Source: "engine", Fingerprint: "fp-" + intentID,
			},
			Kind: KindPlace, AttemptID: attemptID,
		})
		if err != nil {
			t.Fatal(err)
		}
		if err := attempt.MarkDispatchStarted(ctx); err != nil {
			t.Fatal(err)
		}
		if err := attempt.MarkAcked(ctx, "active-reused-day-id"); err != nil {
			t.Fatal(err)
		}
		if err := attempt.Settle(ctx, StateConfirmed, ReasonBrokerAcknowledged, "acked"); err != nil {
			t.Fatal(err)
		}
	}
	confirm("active-old-day", "active-old-attempt", "2026-03-29")
	confirm("active-new-day", "active-new-attempt", "2026-03-30")
	for _, day := range []string{"2026-03-29", "2026-03-30"} {
		obs := observation("active-reused-day-id", "0")
		obs.TradingDay = day
		if _, err := j.RecordFill(ctx, obs); err != nil {
			t.Fatalf("RecordFill(%s): %v", day, err)
		}
	}

	tracked, err := j.TrackedFillOrders(ctx, "acct-1")
	if err != nil {
		t.Fatal(err)
	}
	days := map[string]bool{}
	for _, order := range tracked {
		if order.OrderID == "active-reused-day-id" {
			days[order.TradingDay] = true
		}
	}
	if len(days) != 2 || !days["2026-03-29"] || !days["2026-03-30"] {
		t.Fatalf("tracked reused days = %+v (all=%+v), want both nonterminal scopes", days, tracked)
	}

	live, err := j.LiveOrdersForSymbol(ctx, "acct-1", "us", "AAPL")
	if err != nil {
		t.Fatal(err)
	}
	liveDays := map[string]bool{}
	for _, order := range live {
		if order.OrderID == "active-reused-day-id" {
			liveDays[order.TradingDay] = true
		}
	}
	if len(liveDays) != 2 {
		t.Fatalf("live reused days = %+v (all=%+v), want both nonterminal scopes", liveDays, live)
	}
}

func TestRecordFillRejectsPartialScopeWithoutOverwritingOrRepeatingDelta(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()
	trusted := observation("partial-scope-reused-id", "3")
	if _, err := j.RecordFill(ctx, trusted); err != nil {
		t.Fatal(err)
	}
	partial := trusted
	partial.TradingDay = ""
	partial.FilledQuantity = "5"
	for attempt := 1; attempt <= 2; attempt++ {
		if _, err := j.RecordFill(ctx, partial); !errors.Is(err, ErrInvalidRequest) {
			t.Fatalf("partial observation %d error = %v, want ErrInvalidRequest", attempt, err)
		}
	}
	stored, err := j.LookupFillScoped(ctx, fillSnapshotScopeOf(trusted))
	if err != nil {
		t.Fatal(err)
	}
	if stored.FilledQuantity != "3" || stored.TradingDay != trusted.TradingDay {
		t.Fatalf("trusted scoped snapshot was overwritten: %+v", stored)
	}
	var events int
	if err := j.db.QueryRowContext(ctx,
		`SELECT count(*) FROM fill_events WHERE order_id = ?`, trusted.OrderID).Scan(&events); err != nil {
		t.Fatal(err)
	}
	if events != 1 {
		t.Fatalf("fill events = %d, want one trusted delta and no repeated partial delta", events)
	}
}

func TestRecordFillKeepsReusedOrderIDSnapshotsInTheirCanonicalScopes(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()

	prior := observation("canonically-reused", "7")
	prior.AccountRef, prior.TradingDay = "acct-1", "2026-03-29"
	prior.Symbol, prior.Side, prior.Terminal, prior.State = "MSFT", "SELL", true, "FILLED"
	current := observation("canonically-reused", "2")
	current.AccountRef, current.TradingDay = "acct-2", "2026-03-30"
	current.Symbol, current.Side = "AAPL", "BUY"

	if res, err := j.RecordFill(ctx, prior); err != nil || res.Delta != "7" {
		t.Fatalf("RecordFill(prior) = %+v, %v", res, err)
	}
	if res, err := j.RecordFill(ctx, current); err != nil || res.Delta != "2" || res.FailClosed {
		t.Fatalf("RecordFill(current) = %+v, %v; want an independent cumulative sequence", res, err)
	}

	assertScoped := func(scope FillSnapshotScope, wantQuantity string, wantTerminal bool) {
		t.Helper()
		got, err := j.LookupFillScoped(ctx, scope)
		if err != nil {
			t.Fatalf("LookupFillScoped(%+v): %v", scope, err)
		}
		if got.FilledQuantity != wantQuantity || got.Terminal != wantTerminal {
			t.Fatalf("LookupFillScoped(%+v) = %+v, want quantity=%s terminal=%v",
				scope, got, wantQuantity, wantTerminal)
		}
	}
	assertScoped(FillSnapshotScope{
		OrderID: "canonically-reused", AccountRef: "acct-1", Market: "us",
		TradingDay: "2026-03-29", Symbol: "MSFT", Side: "SELL",
	}, "7", true)
	assertScoped(FillSnapshotScope{
		OrderID: "canonically-reused", AccountRef: "acct-2", Market: "us",
		TradingDay: "2026-03-30", Symbol: "AAPL", Side: "BUY",
	}, "2", false)

	if _, err := j.LookupFill(ctx, "canonically-reused"); !errors.Is(err, ErrFillScopeAmbiguous) {
		t.Fatalf("LookupFill(unscoped) error = %v, want ErrFillScopeAmbiguous", err)
	}
}

func TestScopedTerminalStateDoesNotHideAnotherAccountsLiveReusedOrder(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()
	recordConfirmedFillOrder(t, j, "acct-one-intent", "acct-one-attempt", "live-reused")
	attempt, err := j.Prepare(ctx, PrepareRequest{
		Intent: Intent{
			ID: "acct-two-intent", Market: "us", TradingDay: "2026-03-30", AccountRef: "acct-2",
			Symbol: "AAPL", Side: "BUY", OrderType: "LIMIT", Quantity: "10",
			Price: "200", Currency: "USD", Source: "engine", Fingerprint: "fp-acct-two",
		},
		Kind: KindPlace, AttemptID: "acct-two-attempt",
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := attempt.MarkDispatchStarted(ctx); err != nil {
		t.Fatal(err)
	}
	if err := attempt.MarkAcked(ctx, "live-reused"); err != nil {
		t.Fatal(err)
	}
	if err := attempt.Settle(ctx, StateConfirmed, ReasonBrokerAcknowledged, "acked"); err != nil {
		t.Fatal(err)
	}

	acctOne := observation("live-reused", "10")
	acctOne.Terminal, acctOne.State = true, "FILLED"
	if _, err := j.RecordFill(ctx, acctOne); err != nil {
		t.Fatal(err)
	}
	acctTwo := observation("live-reused", "1")
	acctTwo.AccountRef = "acct-2"
	if _, err := j.RecordFill(ctx, acctTwo); err != nil {
		t.Fatal(err)
	}

	trackedOne, err := j.TrackedFillOrders(ctx, "acct-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(trackedOne) != 0 {
		t.Fatalf("acct-1 tracked = %+v, want terminal scope omitted", trackedOne)
	}
	trackedTwo, err := j.TrackedFillOrders(ctx, "acct-2")
	if err != nil {
		t.Fatal(err)
	}
	if len(trackedTwo) != 1 || trackedTwo[0].IntentID != "acct-two-intent" {
		t.Fatalf("acct-2 tracked = %+v, want its independent live scope", trackedTwo)
	}
	liveTwo, err := j.LiveOrdersForSymbol(ctx, "acct-2", "us", "AAPL")
	if err != nil {
		t.Fatal(err)
	}
	if len(liveTwo) != 1 || liveTwo[0].IntentID != "acct-two-intent" {
		t.Fatalf("acct-2 live = %+v, want its independent live scope", liveTwo)
	}
}

func TestDirectLegacyBlankSnapshotIsNotAWildcardAcrossReusedDays(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()
	confirm := func(intentID, attemptID, day string) {
		attempt, err := j.Prepare(ctx, PrepareRequest{
			Intent: Intent{
				ID: intentID, Market: "us", TradingDay: day, AccountRef: "acct-1",
				Symbol: "AAPL", Side: "BUY", OrderType: "LIMIT", Quantity: "10",
				Price: "200", Currency: "USD", Source: "engine", Fingerprint: "fp-" + intentID,
			},
			Kind: KindPlace, AttemptID: attemptID,
		})
		if err != nil {
			t.Fatal(err)
		}
		if err := attempt.MarkDispatchStarted(ctx); err != nil {
			t.Fatal(err)
		}
		if err := attempt.MarkAcked(ctx, "legacy-direct-reused"); err != nil {
			t.Fatal(err)
		}
		if err := attempt.Settle(ctx, StateConfirmed, ReasonBrokerAcknowledged, "acked"); err != nil {
			t.Fatal(err)
		}
	}
	confirm("legacy-direct-old", "legacy-direct-old-attempt", "2026-03-29")
	confirm("legacy-direct-new", "legacy-direct-new-attempt", "2026-03-30")
	legacy := observation("legacy-direct-reused", "3")
	legacy.AccountRef, legacy.TradingDay, legacy.Side = "", "", ""
	if _, err := j.RecordFill(ctx, legacy); err != nil {
		t.Fatal(err)
	}

	for _, day := range []string{"2026-03-29", "2026-03-30"} {
		_, err := j.LookupFillScoped(ctx, FillSnapshotScope{
			OrderID: "legacy-direct-reused", AccountRef: "acct-1", Market: "us",
			TradingDay: day, Symbol: "AAPL", Side: "BUY",
		})
		if !errors.Is(err, ErrFillNotFound) {
			t.Fatalf("legacy blank snapshot bound to %s: %v", day, err)
		}
	}
	legacyStored, err := j.LookupFill(ctx, "legacy-direct-reused")
	if err != nil {
		t.Fatal(err)
	}
	if legacyStored.AccountRef != "" || legacyStored.TradingDay != "" {
		t.Fatalf("legacy snapshot was rewritten into a scope: %+v", legacyStored)
	}
}

func TestNetPositionsJoinsReusedOrderIDOnlyWithinCanonicalScope(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()
	confirm := func(intentID, attemptID, account, symbol, side string) {
		attempt, err := j.Prepare(ctx, PrepareRequest{
			Intent: Intent{
				ID: intentID, Market: "us", TradingDay: "2026-03-30", AccountRef: account,
				Symbol: symbol, Side: side, OrderType: "LIMIT", Quantity: "10",
				Price: "200", Currency: "USD", Source: "engine", Fingerprint: "fp-" + intentID,
			},
			Kind: KindPlace, AttemptID: attemptID,
		})
		if err != nil {
			t.Fatal(err)
		}
		if err := attempt.MarkDispatchStarted(ctx); err != nil {
			t.Fatal(err)
		}
		if err := attempt.MarkAcked(ctx, "net-reused"); err != nil {
			t.Fatal(err)
		}
		if err := attempt.Settle(ctx, StateConfirmed, ReasonBrokerAcknowledged, "acked"); err != nil {
			t.Fatal(err)
		}
	}
	confirm("net-buy", "net-buy-attempt", "acct-1", "AAPL", "BUY")
	confirm("net-sell", "net-sell-attempt", "acct-2", "MSFT", "SELL")
	buy := observation("net-reused", "3")
	if _, err := j.RecordFill(ctx, buy); err != nil {
		t.Fatal(err)
	}
	sell := observation("net-reused", "2")
	sell.AccountRef, sell.Symbol, sell.Side = "acct-2", "MSFT", "SELL"
	if _, err := j.RecordFill(ctx, sell); err != nil {
		t.Fatal(err)
	}

	net, err := j.NetPositions(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if net["AAPL"] != "3" || net["MSFT"] != "-2" {
		t.Fatalf("NetPositions = %+v, want only exact scoped joins AAPL=3 MSFT=-2", net)
	}
}

func TestTrackedAndLiveOrdersAllowReusedIDInDifferentSymbolAndSideScopes(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()
	confirm := func(intentID, attemptID, symbol, side string) {
		attempt, err := j.Prepare(ctx, PrepareRequest{
			Intent: Intent{
				ID: intentID, Market: "us", TradingDay: "2026-03-30", AccountRef: "acct-1",
				Symbol: symbol, Side: side, OrderType: "LIMIT", Quantity: "10",
				Price: "200", Currency: "USD", Source: "engine", Fingerprint: "fp-" + intentID,
			},
			Kind: KindPlace, AttemptID: attemptID,
		})
		if err != nil {
			t.Fatal(err)
		}
		if err := attempt.MarkDispatchStarted(ctx); err != nil {
			t.Fatal(err)
		}
		if err := attempt.MarkAcked(ctx, "symbol-side-reused"); err != nil {
			t.Fatal(err)
		}
		if err := attempt.Settle(ctx, StateConfirmed, ReasonBrokerAcknowledged, "acked"); err != nil {
			t.Fatal(err)
		}
	}
	confirm("reused-aapl-buy", "reused-aapl-buy-attempt", "AAPL", "BUY")
	confirm("reused-msft-sell", "reused-msft-sell-attempt", "MSFT", "SELL")
	buy := observation("symbol-side-reused", "1")
	if _, err := j.RecordFill(ctx, buy); err != nil {
		t.Fatal(err)
	}
	sell := observation("symbol-side-reused", "2")
	sell.Symbol, sell.Side = "MSFT", "SELL"
	if _, err := j.RecordFill(ctx, sell); err != nil {
		t.Fatal(err)
	}

	tracked, err := j.TrackedFillOrders(ctx, "acct-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(tracked) != 2 {
		t.Fatalf("tracked = %+v, want both distinct symbol/side scopes", tracked)
	}
	for _, symbol := range []string{"AAPL", "MSFT"} {
		live, err := j.LiveOrdersForSymbol(ctx, "acct-1", "us", symbol)
		if err != nil {
			t.Fatal(err)
		}
		if len(live) != 1 || live[0].Symbol != symbol {
			t.Fatalf("live %s = %+v, want its exact reused-id scope", symbol, live)
		}
	}
}

func TestLiveOrdersForSymbolKeepsAReusedCurrentDayOrderAfterThePriorDayTerminal(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()
	confirm := func(intentID, attemptID, day string) {
		attempt, err := j.Prepare(ctx, PrepareRequest{
			Intent: Intent{
				ID: intentID, Market: "us", TradingDay: day, AccountRef: "acct-1",
				Symbol: "AAPL", Side: "BUY", OrderType: "LIMIT", Quantity: "1",
				Price: "100", Currency: "USD", Source: "engine", Fingerprint: "fp-" + intentID,
			},
			Kind: KindPlace, AttemptID: attemptID,
		})
		if err != nil {
			t.Fatal(err)
		}
		if err := attempt.MarkDispatchStarted(ctx); err != nil {
			t.Fatal(err)
		}
		if err := attempt.MarkAcked(ctx, "reused-live-order"); err != nil {
			t.Fatal(err)
		}
		if err := attempt.Settle(ctx, StateConfirmed, ReasonBrokerAcknowledged, "acked"); err != nil {
			t.Fatal(err)
		}
	}

	confirm("prior-day-intent", "prior-day-attempt", "2026-03-29")
	prior := observation("reused-live-order", "1")
	prior.TradingDay, prior.Terminal, prior.State = "2026-03-29", true, "FILLED"
	if _, err := j.RecordFill(ctx, prior); err != nil {
		t.Fatal(err)
	}
	confirm("current-day-intent", "current-day-attempt", "2026-03-30")

	live, err := j.LiveOrdersForSymbol(ctx, "acct-1", "us", "AAPL")
	if err != nil {
		t.Fatal(err)
	}
	if len(live) != 1 || live[0].OrderID != "reused-live-order" ||
		live[0].IntentID != "current-day-intent" || live[0].TradingDay != "2026-03-30" {
		t.Fatalf("live orders = %+v, want the current-day reused order", live)
	}
}

func TestLiveOrdersForSymbolUsesLegacyTerminalOnlyWithinItsMarket(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()
	confirm := func(intentID, attemptID, market, day, symbol, currency string) {
		attempt, err := j.Prepare(ctx, PrepareRequest{
			Intent: Intent{
				ID: intentID, Market: market, TradingDay: day, AccountRef: "acct-1",
				Symbol: symbol, Side: "BUY", OrderType: "LIMIT", Quantity: "1",
				Price: "100", Currency: currency, Source: "engine", Fingerprint: "fp-" + intentID,
			},
			Kind: KindPlace, AttemptID: attemptID,
		})
		if err != nil {
			t.Fatal(err)
		}
		if err := attempt.MarkDispatchStarted(ctx); err != nil {
			t.Fatal(err)
		}
		if err := attempt.MarkAcked(ctx, "legacy-cross-market-id"); err != nil {
			t.Fatal(err)
		}
		if err := attempt.Settle(ctx, StateConfirmed, ReasonBrokerAcknowledged, "acked"); err != nil {
			t.Fatal(err)
		}
	}

	confirm("us-legacy-intent", "us-legacy-attempt", "us", "2026-03-30", "AAPL", "USD")
	terminal := observation("legacy-cross-market-id", "1")
	terminal.AccountRef, terminal.TradingDay, terminal.Side = "", "", ""
	terminal.Symbol, terminal.Market, terminal.Terminal, terminal.State = "AAPL", "us", true, "FILLED"
	if _, err := j.RecordFill(ctx, terminal); err != nil {
		t.Fatal(err)
	}
	confirm("kr-reused-intent", "kr-reused-attempt", "kr", "2026-03-31", "005930", "KRW")

	live, err := j.LiveOrdersForSymbol(ctx, "acct-1", "us", "AAPL")
	if err != nil {
		t.Fatal(err)
	}
	if len(live) != 0 {
		t.Fatalf("US live orders = %+v, want legacy US terminal to remain authoritative despite KR id reuse", live)
	}
}

func TestLegacySnapshotCannotTerminateOrOwnAFutureReusedOrder(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()
	if _, err := j.db.ExecContext(ctx, `INSERT INTO fill_snapshots
		(order_id, account_ref, symbol, market, trading_day, side, state, terminal,
		 filled_quantity, committed_at)
		VALUES ('future-snapshot','','AAPL','us','','','FILLED',1,'1',
		        '2026-03-29T20:00:00Z')`); err != nil {
		t.Fatal(err)
	}
	recordConfirmedFillOrder(t, j, "future-intent", "future-attempt", "future-snapshot")

	tracked, err := j.TrackedFillOrders(ctx, "acct-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(tracked) != 1 || tracked[0].OrderID != "future-snapshot" {
		t.Fatalf("tracked=%+v, want future order retained without inherited snapshot", tracked)
	}
	live, err := j.LiveOrdersForSymbol(ctx, "acct-1", "us", "AAPL")
	if err != nil {
		t.Fatal(err)
	}
	if len(live) != 1 || live[0].OrderID != "future-snapshot" {
		t.Fatalf("live=%+v, want future order not hidden by old terminal", live)
	}
	net, err := j.NetPositions(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if net["AAPL"] != "" {
		t.Fatalf("net=%+v, want old external snapshot excluded from future owner", net)
	}
}

func TestScopedExternalSnapshotCannotTerminateOrOwnAFutureOrder(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()
	if _, err := j.db.ExecContext(ctx, `INSERT INTO scoped_fill_snapshots
		(order_id, account_ref, symbol, market, trading_day, side, state, terminal,
		 filled_quantity, committed_at)
		VALUES ('future-scoped','acct-1','AAPL','us','2026-03-30','BUY','FILLED',1,'1',
		        '2026-03-29T20:00:00Z')`); err != nil {
		t.Fatal(err)
	}
	recordConfirmedFillOrder(t, j, "future-scoped-intent", "future-scoped-attempt", "future-scoped")

	tracked, err := j.TrackedFillOrders(ctx, "acct-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(tracked) != 1 || tracked[0].OrderID != "future-scoped" {
		t.Fatalf("tracked=%+v, want future order retained without inherited scoped snapshot", tracked)
	}
	live, err := j.LiveOrdersForSymbol(ctx, "acct-1", "us", "AAPL")
	if err != nil {
		t.Fatal(err)
	}
	if len(live) != 1 || live[0].OrderID != "future-scoped" {
		t.Fatalf("live=%+v, want future order not hidden by pre-owner scoped terminal", live)
	}
	net, err := j.NetPositions(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if net["AAPL"] != "" {
		t.Fatalf("net=%+v, want pre-owner scoped snapshot excluded", net)
	}
}

func TestOrderIDLatestDayPartitionKeepsBothMarketsInTheSameAccount(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()
	confirm := func(intentID, attemptID, market, day, symbol, currency string) {
		attempt, err := j.Prepare(ctx, PrepareRequest{
			Intent: Intent{
				ID: intentID, Market: market, TradingDay: day, AccountRef: "acct-1",
				Symbol: symbol, Side: "BUY", OrderType: "LIMIT", Quantity: "1",
				Price: "100", Currency: currency, Source: "engine", Fingerprint: "fp-" + intentID,
			},
			Kind: KindPlace, AttemptID: attemptID,
		})
		if err != nil {
			t.Fatal(err)
		}
		if err := attempt.MarkDispatchStarted(ctx); err != nil {
			t.Fatal(err)
		}
		if err := attempt.MarkAcked(ctx, "cross-market-id"); err != nil {
			t.Fatal(err)
		}
		if err := attempt.Settle(ctx, StateConfirmed, ReasonBrokerAcknowledged, "acked"); err != nil {
			t.Fatal(err)
		}
	}

	confirm("us-intent", "us-attempt", "us", "2026-03-30", "AAPL", "USD")
	confirm("kr-intent", "kr-attempt", "kr", "2026-03-31", "005930", "KRW")

	tracked, err := j.TrackedFillOrders(ctx, "acct-1")
	if err != nil {
		t.Fatal(err)
	}
	markets := map[string]bool{}
	for _, order := range tracked {
		markets[order.Market] = true
	}
	if len(tracked) != 2 || !markets["us"] || !markets["kr"] {
		t.Fatalf("tracked orders = %+v, want both market-scoped identities", tracked)
	}
	live, err := j.LiveOrdersForSymbol(ctx, "acct-1", "us", "AAPL")
	if err != nil {
		t.Fatal(err)
	}
	if len(live) != 1 || live[0].IntentID != "us-intent" {
		t.Fatalf("US live orders = %+v, want the US order despite the later KR day", live)
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

	recordConfirmedFillOrder(t, j, "intent-parent", "attempt-parent", "parent-1")
	if _, err := j.RecordFill(ctx, observation("parent-1", "2")); err != nil {
		t.Fatal(err)
	}
	amend, err := j.Prepare(ctx, PrepareRequest{
		Intent: Intent{
			ID: "intent-child", Market: "us", TradingDay: "2026-03-30", AccountRef: "acct-1",
			Symbol: "AAPL", Side: "BUY", OrderType: "LIMIT", Quantity: "8",
			Price: "200", Currency: "USD", Source: "engine", Fingerprint: "fp-intent-child",
		},
		Kind: KindAmend, AttemptID: "attempt-child", TargetOrderID: "parent-1",
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := amend.MarkDispatchStarted(ctx); err != nil {
		t.Fatal(err)
	}
	if err := amend.MarkAcked(ctx, "child-1"); err != nil {
		t.Fatal(err)
	}
	if err := amend.ResolveConfirmedWithLineage(ctx, LineageEdge{
		ParentOrderID: "parent-1", ChildOrderID: "child-1",
		ParentFilledQuantity: "2", RequestedQuantity: "8",
	}, ReasonResolvedFound, "amended"); err != nil {
		t.Fatal(err)
	}

	tracked, err := j.TrackedFillOrders(ctx)
	if err != nil {
		t.Fatal(err)
	}
	var foundParent, foundChild bool
	for _, tr := range tracked {
		if tr.OrderID == "parent-1" {
			foundParent = true
			if tr.SuccessorOrderID != "child-1" {
				t.Fatalf("successor = %q, want child-1", tr.SuccessorOrderID)
			}
		}
		foundChild = foundChild || tr.OrderID == "child-1"
	}
	if !foundParent || !foundChild {
		t.Fatalf("tracked = %+v, want live parent and confirmed unseen child", tracked)
	}

	done := observation("child-1", "8")
	done.Terminal = true
	done.State = "FILLED"
	if _, err := j.RecordFill(ctx, done); err != nil {
		t.Fatal(err)
	}
	tracked, err = j.TrackedFillOrders(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for _, tr := range tracked {
		if tr.OrderID == "child-1" {
			t.Fatalf("terminal successor remained tracked: %+v", tracked)
		}
	}
}

func TestTrackedFillOrdersScopeReusedLineageEndpointByAccount(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()
	recordConfirmedFillOrder(t, j, "intent-parent", "attempt-parent", "reused-parent")
	recordConfirmedFillOrder(t, j, "intent-child", "attempt-child", "owned-child")
	if _, err := j.RecordFill(ctx, observation("reused-parent", "1")); err != nil {
		t.Fatal(err)
	}
	if _, err := j.db.ExecContext(ctx,
		`INSERT INTO lineage_edges
		   (parent_order_id, child_order_id, relation, parent_filled_quantity,
		    requested_quantity, intent_id, attempt_id, created_at)
		 VALUES ('reused-parent','owned-child','replaces','1','9','intent-parent','attempt-parent',
		         '2026-03-30T00:30:00Z')`); err != nil {
		t.Fatal(err)
	}
	attempt, err := j.Prepare(ctx, PrepareRequest{
		Intent: Intent{
			ID: "other-parent-intent", Market: "us", TradingDay: "2026-03-30", AccountRef: "acct-2",
			Symbol: "AAPL", Side: "BUY", OrderType: "LIMIT", Quantity: "10",
			Price: "200", Currency: "USD", Source: "engine", Fingerprint: "other-parent-fp",
		},
		Kind: KindPlace, AttemptID: "other-parent-attempt",
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := attempt.MarkDispatchStarted(ctx); err != nil {
		t.Fatal(err)
	}
	if err := attempt.MarkAcked(ctx, "reused-parent"); err != nil {
		t.Fatal(err)
	}
	if err := attempt.Settle(ctx, StateConfirmed, ReasonBrokerAcknowledged, "acked"); err != nil {
		t.Fatal(err)
	}

	tracked, err := j.TrackedFillOrders(ctx, "acct-1")
	if err != nil {
		t.Fatal(err)
	}
	var parent, child bool
	for _, order := range tracked {
		parent = parent || order.OrderID == "reused-parent"
		child = child || order.OrderID == "owned-child"
	}
	if !parent || !child {
		t.Fatalf("selected account lineage was lost because another account reused an id: %+v", tracked)
	}
	if active, err := j.ActiveReconcileStates(ctx); err != nil || len(active) != 0 {
		t.Fatalf("cross-account scope raised a false durable conflict: states=%+v err=%v", active, err)
	}
}

func TestTrackedFillOrdersRejectMalformedLegacyLineageOwnership(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()
	recordConfirmedFillOrder(t, j, "owned-intent", "owned-place-attempt", "owned-parent")
	external := observation("external-child", "0")
	if _, err := j.RecordFill(ctx, external); err != nil {
		t.Fatal(err)
	}
	// A pre-v16/corrupt edge tied to a PLACE attempt is not amendment ownership:
	// the attempt names neither parent as its target nor child as its broker id.
	if _, err := j.db.ExecContext(ctx, `
		INSERT INTO lineage_edges
		  (parent_order_id, child_order_id, relation, parent_filled_quantity,
		   requested_quantity, intent_id, attempt_id, created_at)
		VALUES ('owned-parent','external-child','replaces','0','10',
		        'owned-intent','owned-place-attempt','2026-03-30T00:30:00Z')`); err != nil {
		t.Fatal(err)
	}

	tracked, err := j.TrackedFillOrders(ctx, "acct-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(tracked) != 1 || tracked[0].OrderID != "owned-parent" || tracked[0].SuccessorOrderID != "" {
		t.Fatalf("tracked = %+v, want only the confirmed parent without malformed successor", tracked)
	}
}

func TestTrackedFillOrdersUseScopedLineageWhenTheSamePairIsReused(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()
	other := OrderLineageScope{
		AccountRef: "acct-2", Market: "us", TradingDay: "2026-03-29", Symbol: "AAPL", Side: "BUY",
	}
	selected := other
	selected.AccountRef, selected.TradingDay = "acct-1", "2026-03-30"
	recordConfirmedReplacement(t, j, other, "reused-parent", "reused-child", "tracked-other")
	recordConfirmedReplacement(t, j, selected, "reused-parent", "reused-child", "tracked-selected")
	parent := observation("reused-parent", "0")
	parent.AccountRef, parent.TradingDay = selected.AccountRef, selected.TradingDay
	if _, err := j.RecordFill(ctx, parent); err != nil {
		t.Fatal(err)
	}

	tracked, err := j.TrackedFillOrders(ctx, selected.AccountRef)
	if err != nil {
		t.Fatal(err)
	}
	foundParent := false
	for _, order := range tracked {
		if order.OrderID == "reused-parent" {
			foundParent = true
			if order.SuccessorOrderID != "reused-child" || order.TradingDay != selected.TradingDay {
				t.Fatalf("selected parent = %+v, want scoped successor on %s", order, selected.TradingDay)
			}
		}
	}
	if !foundParent {
		t.Fatalf("tracked = %+v, want lineage-owned parent from scoped v16 evidence", tracked)
	}
}

func TestTrackedFillOrdersDoNotBindLegacyLineageSnapshotAcrossReusedDays(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()
	oldScope := OrderLineageScope{
		AccountRef: "acct-1", Market: "us", TradingDay: "2026-03-29", Symbol: "AAPL", Side: "BUY",
	}
	newScope := oldScope
	newScope.TradingDay = "2026-03-30"
	recordConfirmedReplacement(t, j, oldScope, "reused-legacy-parent", "reused-legacy-child", "legacy-old")
	recordConfirmedReplacement(t, j, newScope, "reused-legacy-parent", "reused-legacy-child", "legacy-new")
	legacy := observation("reused-legacy-parent", "0")
	// Reproduce a genuine pre-v16 row: those snapshots had market and symbol,
	// but no account, trading day, or side. A half-populated v16 identity is now
	// rejected because it cannot be made idempotent when order ids are reused.
	legacy.AccountRef, legacy.TradingDay, legacy.Side = "", "", ""
	if _, err := j.RecordFill(ctx, legacy); err != nil {
		t.Fatal(err)
	}

	if _, err := j.TrackedFillOrders(ctx, "acct-1"); !errors.Is(err, ErrTrackedFillIdentityConflict) {
		t.Fatalf("TrackedFillOrders err=%v, want a durable ambiguity conflict", err)
	}
}

func TestLegacySnapshotWithCrossAccountOrSideOwnersFailsClosed(t *testing.T) {
	for _, tc := range []struct {
		name   string
		second FillSnapshotScope
	}{
		{name: "account", second: FillSnapshotScope{
			AccountRef: "acct-2", Market: "us", TradingDay: "2026-03-30", Symbol: "AAPL", Side: "BUY",
		}},
		{name: "side", second: FillSnapshotScope{
			AccountRef: "acct-1", Market: "us", TradingDay: "2026-03-30", Symbol: "AAPL", Side: "SELL",
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			j := openTestJournal(t)
			ctx := context.Background()
			first := FillSnapshotScope{
				AccountRef: "acct-1", Market: "us", TradingDay: "2026-03-30", Symbol: "AAPL", Side: "BUY",
			}
			recordConfirmedFillOrderScope(t, j, "legacy-owner-a", "legacy-attempt-a",
				"legacy-ambiguous", first)
			recordConfirmedFillOrderScope(t, j, "legacy-owner-b", "legacy-attempt-b",
				"legacy-ambiguous", tc.second)
			legacy := observation("legacy-ambiguous", "1")
			legacy.AccountRef, legacy.TradingDay, legacy.Side = "", "", ""
			if _, err := j.RecordFill(ctx, legacy); err != nil {
				t.Fatal(err)
			}
			if _, err := j.TrackedFillOrders(ctx, "acct-1"); !errors.Is(err, ErrTrackedFillIdentityConflict) {
				t.Fatalf("TrackedFillOrders err=%v, want ambiguity conflict", err)
			}
			net, err := j.NetPositions(ctx)
			if err != nil {
				t.Fatal(err)
			}
			if net["AAPL"] != "" {
				t.Fatalf("net=%+v, want ambiguous legacy snapshot excluded", net)
			}
		})
	}
}

func TestLiveOrdersForSymbolRejectsDuplicateCanonicalOwner(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()
	scope := FillSnapshotScope{
		AccountRef: "acct-1", Market: "us", TradingDay: "2026-03-30", Symbol: "AAPL", Side: "BUY",
	}
	recordConfirmedFillOrderScope(t, j, "live-owner-a", "live-attempt-a", "live-ambiguous", scope)
	recordConfirmedFillOrderScope(t, j, "live-owner-b", "live-attempt-b", "live-ambiguous", scope)

	if _, err := j.LiveOrdersForSymbol(ctx, scope.AccountRef, scope.Market, scope.Symbol); !errors.Is(err, ErrTrackedFillIdentityConflict) {
		t.Fatalf("LiveOrdersForSymbol err=%v, want durable identity conflict", err)
	}
}

func TestConfirmedCancelDoesNotBecomeASecondOrderOwner(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()
	scope := FillSnapshotScope{
		AccountRef: "acct-1", Market: "us", TradingDay: "2026-03-30", Symbol: "AAPL", Side: "BUY",
	}
	recordConfirmedFillOrderScope(t, j, "place-owner", "place-owner-attempt", "cancelled-order", scope)
	if _, err := j.RecordFill(ctx, observation("cancelled-order", "0")); err != nil {
		t.Fatal(err)
	}
	cancel, err := j.Prepare(ctx, PrepareRequest{
		Intent: Intent{ID: "cancel-intent", Market: scope.Market, TradingDay: scope.TradingDay,
			AccountRef: scope.AccountRef, Symbol: scope.Symbol, Side: scope.Side, OrderType: "LIMIT",
			Quantity: "10", Price: "200", Currency: "USD", Source: "engine", Fingerprint: "cancel-fp"},
		Kind: KindCancel, AttemptID: "cancel-attempt", TargetOrderID: "cancelled-order",
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := cancel.MarkDispatchStarted(ctx); err != nil {
		t.Fatal(err)
	}
	if err := cancel.MarkAcked(ctx, "cancelled-order"); err != nil {
		t.Fatal(err)
	}
	if err := cancel.Settle(ctx, StateConfirmed, ReasonBrokerAcknowledged, "cancel accepted"); err != nil {
		t.Fatal(err)
	}

	live, err := j.LiveOrdersForSymbol(ctx, scope.AccountRef, scope.Market, scope.Symbol)
	if err != nil {
		t.Fatalf("confirmed CANCEL poisoned live ownership: %v", err)
	}
	if len(live) != 1 || live[0].IntentID != "place-owner" {
		t.Fatalf("live=%+v, want only the PLACE owner", live)
	}
}

func TestPostOwnerFillStartsANewCumulativeSequenceAfterExternalEvidence(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()
	scope := FillSnapshotScope{OrderID: "reused-external-scope", AccountRef: "acct-1", Market: "us",
		TradingDay: "2026-03-30", Symbol: "AAPL", Side: "BUY"}
	if _, err := j.db.ExecContext(ctx, `INSERT INTO scoped_fill_snapshots
		(account_ref, market, trading_day, symbol, side, order_id, state, terminal,
		 fail_closed, quantity, filled_quantity, average_price, committed_at)
		VALUES (?,?,?,?,?,?,'FILLED',1,0,'10','10','190','2026-03-30T00:30:00Z')`,
		scope.AccountRef, scope.Market, scope.TradingDay, scope.Symbol, scope.Side, scope.OrderID); err != nil {
		t.Fatal(err)
	}
	recordConfirmedFillOrderScope(t, j, "post-external-owner", "post-external-attempt", scope.OrderID, scope)
	var projectedDelta string
	if err := j.SetApplyHooks(ApplyHooks{
		Project: func(_ context.Context, _ *ApplyTx, fill AppliedFill) error {
			projectedDelta = fill.Delta
			return nil
		},
		Exit: func(context.Context, *ApplyTx, AppliedFill) error { return nil },
	}); err != nil {
		t.Fatal(err)
	}
	obs := observation(scope.OrderID, "12")
	obs.State, obs.Terminal = "FILLED", true
	res, err := j.RecordFill(ctx, obs)
	if err != nil {
		t.Fatal(err)
	}
	if res.Delta != "12" || projectedDelta != "12" {
		t.Fatalf("result delta=%s projected=%s, want new owner sequence to start at 12",
			res.Delta, projectedDelta)
	}
	events, err := j.FillEventsScoped(ctx, scope)
	if err != nil || len(events) != 1 || events[0].DeltaQuantity != "12" {
		t.Fatalf("post-owner events=%+v err=%v, want one full local cumulative delta", events, err)
	}
}

func TestLateParentFillStartsANewSequenceAfterExternalEvidenceAndReplacement(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()
	scope := FillSnapshotScope{OrderID: "replaced-external-parent", AccountRef: "acct-1", Market: "us",
		TradingDay: "2026-03-30", Symbol: "AAPL", Side: "BUY"}
	if _, err := j.db.ExecContext(ctx, `INSERT INTO scoped_fill_snapshots
		(account_ref, market, trading_day, symbol, side, order_id, state, terminal,
		 fail_closed, quantity, filled_quantity, average_price, committed_at)
		VALUES (?,?,?,?,?,?,'FILLED',1,0,'10','10','190','2026-03-30T00:30:00Z')`,
		scope.AccountRef, scope.Market, scope.TradingDay, scope.Symbol, scope.Side, scope.OrderID); err != nil {
		t.Fatal(err)
	}
	recordConfirmedFillOrderScope(t, j, "replaced-parent-owner", "replaced-parent-attempt", scope.OrderID, scope)
	recordConfirmedReplacement(t, j, OrderLineageScope{
		AccountRef: scope.AccountRef, Market: scope.Market, TradingDay: scope.TradingDay,
		Symbol: scope.Symbol, Side: scope.Side,
	}, scope.OrderID, "replaced-external-child", "external-parent")

	var projectedDelta string
	if err := j.SetApplyHooks(ApplyHooks{
		Project: func(_ context.Context, _ *ApplyTx, fill AppliedFill) error {
			projectedDelta = fill.Delta
			return nil
		},
		Exit: func(context.Context, *ApplyTx, AppliedFill) error { return nil },
	}); err != nil {
		t.Fatal(err)
	}
	obs := observation(scope.OrderID, "12")
	obs.State, obs.Terminal = "FILLED", true
	res, err := j.RecordFill(ctx, obs)
	if err != nil {
		t.Fatal(err)
	}
	if res.Delta != "12" || projectedDelta != "12" {
		t.Fatalf("result delta=%s projected=%s, want full parent cumulative after direct-owner precedence",
			res.Delta, projectedDelta)
	}
	events, err := j.FillEventsScoped(ctx, scope)
	if err != nil || len(events) != 1 || events[0].DeltaQuantity != "12" {
		t.Fatalf("late parent events=%+v err=%v, want one full local cumulative delta", events, err)
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

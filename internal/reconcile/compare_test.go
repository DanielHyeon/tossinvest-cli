package reconcile_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/reconcile"
)

// Comparison tests (harden-execution-base task 3.4).
//
// Three requirements meet here: quantity tolerance zero, an average-price epsilon
// that is reported and never blocks, and external provenance for anything the
// broker has that the engine never asked for.

func snapshotWith(holdings []reconcile.Holding, orders []reconcile.BrokerOrder) reconcile.Snapshot {
	return reconcile.Snapshot{
		AsOf: asOf, CompletedAt: asOf, AccountRef: "acct-7",
		Holdings: holdings, OpenOrders: orders,
	}
}

// TestQuantitiesHaveNoBusinessTolerance: one share of disagreement is a
// disagreement, and it blocks new entries.
func TestQuantitiesHaveNoBusinessTolerance(t *testing.T) {
	snap := snapshotWith([]reconcile.Holding{{Symbol: "AAPL", Quantity: "10"}}, nil)
	local := reconcile.LocalState{
		AccountRef: "acct-7",
		Positions:  map[string]string{"AAPL": "9"},
	}

	diff := reconcile.Comparer{}.Compare(snap, local)
	if len(diff.Quantities) != 1 {
		t.Fatalf("quantity mismatches = %+v, want 1", diff.Quantities)
	}
	if !diff.BlocksEntry() {
		t.Fatal("a quantity mismatch must block new entries")
	}
	if got := diff.Quantities[0].Authority(); got != "10" {
		t.Fatalf("authority = %q, want the account's 10 — the account always wins", got)
	}
}

// TestFractionalQuantitiesSurviveTheFloatRoundTrip: 0.1 + 0.2 comes back as
// 0.30000000000000004, and blocking trading on that would be an outage caused by
// binary arithmetic rather than by a real discrepancy.
func TestFractionalQuantitiesSurviveTheFloatRoundTrip(t *testing.T) {
	snap := snapshotWith([]reconcile.Holding{{Symbol: "AAPL", Quantity: "0.3"}}, nil)
	local := reconcile.LocalState{
		Positions: map[string]string{"AAPL": "0.30000000000000004"},
	}
	diff := reconcile.Comparer{}.Compare(snap, local)
	if diff.BlocksEntry() {
		t.Fatalf("a float round-trip artefact must not block: %+v", diff.Quantities)
	}
	if diff.Matched != 1 {
		t.Fatalf("matched = %d, want 1", diff.Matched)
	}
}

// TestAveragePriceDeviationIsReportedButNeverBlocks is the spec's
// "평균단가는 … 진입 차단 판정에서 제외".
func TestAveragePriceDeviationIsReportedButNeverBlocks(t *testing.T) {
	snap := snapshotWith([]reconcile.Holding{
		{Symbol: "AAPL", Quantity: "10", AveragePrice: "213.4"},
	}, nil)

	deviations := reconcile.Comparer{}.ComparePositionPrices(snap,
		map[string]string{"AAPL": "213.9"})
	if len(deviations) != 1 {
		t.Fatalf("deviations = %+v, want 1", deviations)
	}
	if deviations[0].Blocking {
		t.Fatal("an average-price deviation must never be marked blocking")
	}

	diff := reconcile.Comparer{}.Compare(snap, reconcile.LocalState{
		Positions: map[string]string{"AAPL": "10"},
	})
	diff.Prices = deviations
	if diff.BlocksEntry() {
		t.Fatal("a diff whose only finding is a price deviation must not block entries")
	}
	if diff.Clean() {
		t.Fatal("it must still be reported")
	}
}

// TestAveragePriceWithinEpsilonIsNotADeviation: the two systems round a weighted
// mean differently, and a rounding difference is not news.
func TestAveragePriceWithinEpsilonIsNotADeviation(t *testing.T) {
	snap := snapshotWith([]reconcile.Holding{
		{Symbol: "005930", Quantity: "10", AveragePrice: "70000"},
	}, nil)
	deviations := reconcile.Comparer{}.ComparePositionPrices(snap,
		map[string]string{"005930": "70000.03"})
	if len(deviations) != 0 {
		t.Fatalf("deviations = %+v, want none inside the documented epsilon", deviations)
	}
}

// TestIdenticalPriceStringsShortCircuit honours the spec's "decimal 문자열 비교"
// wording: two identical strings cannot differ numerically, so no float is
// consulted at all.
func TestIdenticalPriceStringsShortCircuit(t *testing.T) {
	snap := snapshotWith([]reconcile.Holding{
		{Symbol: "AAPL", Quantity: "1", AveragePrice: "0.1"},
	}, nil)
	got := reconcile.Comparer{}.ComparePositionPrices(snap, map[string]string{"AAPL": "0.1"})
	if len(got) != 0 {
		t.Fatalf("identical decimal strings must not deviate: %+v", got)
	}
}

// TestExternalOrderIsClassifiedNotBlocked is the spec's "외부 수동 주문 발견"
// scenario: the owner trading their own account by hand is a fact to record and
// alert on, not a malfunction to stop the engine for.
func TestExternalOrderIsClassifiedNotBlocked(t *testing.T) {
	snap := snapshotWith(nil, []reconcile.BrokerOrder{
		{OrderID: "manual-1", Symbol: "AAPL", Side: "BUY", Quantity: "3"},
	})
	diff := reconcile.Comparer{}.Compare(snap, reconcile.LocalState{
		OpenOrders: map[string]reconcile.LocalOrder{},
	})

	if len(diff.ExternalOrd) != 1 {
		t.Fatalf("external orders = %+v, want 1", diff.ExternalOrd)
	}
	if diff.ExternalOrd[0].Provenance != reconcile.ProvenanceExternal {
		t.Fatalf("provenance = %q, want external", diff.ExternalOrd[0].Provenance)
	}
	if diff.BlocksEntry() {
		t.Fatal("an external order must not block the engine's own entries")
	}
	if diff.Clean() {
		t.Fatal("an external order must still be reported")
	}
}

// TestExternalPositionIsClassified covers the holdings side of the same idea.
func TestExternalPositionIsClassified(t *testing.T) {
	snap := snapshotWith([]reconcile.Holding{{Symbol: "TSLA", Quantity: "4"}}, nil)
	diff := reconcile.Comparer{}.Compare(snap, reconcile.LocalState{
		Positions: map[string]string{},
	})
	if len(diff.ExternalPos) != 1 || diff.ExternalPos[0].Symbol != "TSLA" {
		t.Fatalf("external positions = %+v, want TSLA", diff.ExternalPos)
	}
	if diff.BlocksEntry() {
		t.Fatal("an external position must not block entries")
	}
}

// TestAnOrderTheAccountDoesNotShowBlocks: the engine thinks it has exposure the
// broker has never heard of. That is the engine not knowing its own position,
// which is exactly what must stop new entries.
func TestAnOrderTheAccountDoesNotShowBlocks(t *testing.T) {
	diff := reconcile.Comparer{}.Compare(snapshotWith(nil, nil), reconcile.LocalState{
		OpenOrders: map[string]reconcile.LocalOrder{
			"o-1": {OrderID: "o-1", Symbol: "AAPL"},
		},
	})
	if len(diff.MissingOrders) != 1 {
		t.Fatalf("missing orders = %+v, want 1", diff.MissingOrders)
	}
	if !diff.BlocksEntry() {
		t.Fatal("an order the account does not show must block new entries")
	}
}

// TestAgreementIsClean is the ordinary case.
func TestAgreementIsClean(t *testing.T) {
	snap := snapshotWith(
		[]reconcile.Holding{{Symbol: "AAPL", Quantity: "10", AveragePrice: "200"}},
		[]reconcile.BrokerOrder{{OrderID: "o-1", Symbol: "AAPL", Side: "SELL", Quantity: "10"}},
	)
	diff := reconcile.Comparer{}.Compare(snap, reconcile.LocalState{
		Positions:  map[string]string{"AAPL": "10"},
		OpenOrders: map[string]reconcile.LocalOrder{"o-1": {OrderID: "o-1", Symbol: "AAPL"}},
	})
	if !diff.Clean() {
		t.Fatalf("expected a clean diff, got %s", diff.Summary())
	}
	if diff.Matched != 1 {
		t.Fatalf("matched = %d, want 1", diff.Matched)
	}
}

// TestZeroOnBothSidesIsNotAFinding keeps a symbol the engine has closed out of
// the report entirely.
func TestZeroOnBothSidesIsNotAFinding(t *testing.T) {
	snap := snapshotWith([]reconcile.Holding{{Symbol: "AAPL", Quantity: "0"}}, nil)
	diff := reconcile.Comparer{}.Compare(snap, reconcile.LocalState{
		Positions: map[string]string{"AAPL": "0"},
	})
	if !diff.Clean() {
		t.Fatalf("a flat symbol on both sides is nothing to report: %s", diff.Summary())
	}
}

// --- journal-backed local state ---------------------------------------------

func openJournal(t *testing.T) *journal.Journal {
	t.Helper()
	j, err := journal.Open(context.Background(), journal.Options{
		Path:  filepath.Join(t.TempDir(), "journal.db"),
		Clock: clock.NewFake(asOf),
		// This repository lives on ntfs; the guard has its own tests in
		// internal/journal.
		FSProber: journal.FixedFSProber(journal.FSInfo{Name: "ext4", Magic: journal.MagicExt}),
	})
	if err != nil {
		t.Fatalf("journal.Open: %v", err)
	}
	t.Cleanup(func() { _ = j.Close() })
	return j
}

// confirmedOrder records an intent, an attempt and the broker order it produced.
func confirmedOrder(t *testing.T, j *journal.Journal, intentID, attemptID, orderID, symbol, side string) {
	t.Helper()
	ctx := context.Background()
	attempt, err := j.Prepare(ctx, journal.PrepareRequest{
		Intent: journal.Intent{
			ID: intentID, Market: "us", TradingDay: "2026-03-30", AccountRef: "acct-7",
			Symbol: symbol, Side: side, OrderType: "LIMIT", Quantity: "10", Price: "200",
			Currency: "USD", Source: "engine", Fingerprint: "fp-" + intentID,
		},
		Kind: journal.KindPlace, AttemptID: attemptID,
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
	if err := attempt.Settle(ctx, journal.StateConfirmed, journal.ReasonBrokerAcknowledged, "acked"); err != nil {
		t.Fatal(err)
	}
}

// TestLocalStateNetsBuysAgainstSells: comparing gross fills against a holding
// would report a mismatch on every completed round trip.
func TestLocalStateNetsBuysAgainstSells(t *testing.T) {
	j := openJournal(t)
	ctx := context.Background()

	confirmedOrder(t, j, "intent-buy", "attempt-buy", "o-buy", "AAPL", "BUY")
	confirmedOrder(t, j, "intent-sell", "attempt-sell", "o-sell", "AAPL", "SELL")

	for _, obs := range []journal.FillObservation{
		{OrderID: "o-buy", Symbol: "AAPL", Market: "us", State: "FILLED", Terminal: true,
			Quantity: "10", FilledQuantity: "10", ObservedAt: "2026-03-30T01:30:00Z"},
		{OrderID: "o-sell", Symbol: "AAPL", Market: "us", State: "FILLED", Terminal: true,
			Quantity: "4", FilledQuantity: "4", ObservedAt: "2026-03-30T01:30:00Z"},
	} {
		if _, err := j.RecordFill(ctx, obs); err != nil {
			t.Fatal(err)
		}
	}

	local, err := reconcile.LocalStateFromJournal(ctx, j, "acct-7")
	if err != nil {
		t.Fatal(err)
	}
	if local.Positions["AAPL"] != "6" {
		t.Fatalf("net position = %q, want 6 (10 bought minus 4 sold)", local.Positions["AAPL"])
	}
}

// TestLocalStateResolvesLineageBeforeComparing is the comparison key's lineage
// step. Without it an amended order reads as one missing order plus one external
// order — one order turned into two wrong findings.
func TestLocalStateResolvesLineageBeforeComparing(t *testing.T) {
	j := openJournal(t)
	ctx := context.Background()

	confirmedOrder(t, j, "intent-1", "attempt-1", "order-original", "AAPL", "BUY")
	if _, err := j.RecordFill(ctx, journal.FillObservation{
		OrderID: "order-original", Symbol: "AAPL", Market: "us",
		State: "OPEN_UNFILLED", Quantity: "10", FilledQuantity: "0",
		ObservedAt: "2026-03-30T01:30:00Z",
	}); err != nil {
		t.Fatal(err)
	}

	// An amendment closed the original and created a successor.
	amend, err := j.Prepare(ctx, journal.PrepareRequest{
		Intent: journal.Intent{
			ID: "intent-2", Market: "us", TradingDay: "2026-03-30", AccountRef: "acct-7",
			Symbol: "AAPL", Side: "BUY", OrderType: "LIMIT", Quantity: "10", Price: "205",
			Currency: "USD", Source: "engine", Fingerprint: "fp-2",
		},
		Kind: journal.KindAmend, AttemptID: "attempt-2", TargetOrderID: "order-original",
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := amend.MarkDispatchStarted(ctx); err != nil {
		t.Fatal(err)
	}
	if err := amend.MarkInDoubt(ctx, "x", "y"); err != nil {
		t.Fatal(err)
	}
	if err := amend.ResolveConfirmedWithLineage(ctx, journal.LineageEdge{
		ParentOrderID: "order-original", ChildOrderID: "order-replacement",
		Relation: journal.RelationReplaces, ParentFilledQuantity: "0", RequestedQuantity: "10",
	}, journal.ReasonResolvedFound, "found the successor"); err != nil {
		t.Fatal(err)
	}

	local, err := reconcile.LocalStateFromJournal(ctx, j, "acct-7")
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := local.OpenOrders["order-replacement"]; !ok {
		t.Fatalf("local open orders = %+v, want the lineage-resolved replacement", local.OpenOrders)
	}

	// The broker shows only the replacement. With lineage resolved, that agrees.
	snap := snapshotWith(nil, []reconcile.BrokerOrder{
		{OrderID: "order-replacement", Symbol: "AAPL", Side: "BUY", Quantity: "10"},
	})
	diff := reconcile.Comparer{}.Compare(snap, local)
	if diff.BlocksEntry() {
		t.Fatalf("an amended order must reconcile against its successor: %s", diff.Summary())
	}
	if len(diff.ExternalOrd) != 0 {
		t.Fatalf("the successor is the engine's own order, not external: %+v", diff.ExternalOrd)
	}
}

// TestFailClosedSnapshotsAreExcludedFromLocalBelief: a quantity the ledger
// refused is not a quantity the engine may claim to know.
func TestFailClosedSnapshotsAreExcludedFromLocalBelief(t *testing.T) {
	j := openJournal(t)
	ctx := context.Background()
	confirmedOrder(t, j, "intent-1", "attempt-1", "o-1", "AAPL", "BUY")

	if _, err := j.RecordFill(ctx, journal.FillObservation{
		OrderID: "o-1", Symbol: "AAPL", Market: "us", State: "UNKNOWN_BROKER_STATE",
		FailClosed: true, Reason: "unknown_status", Detail: "unreadable",
		Quantity: "10", FilledQuantity: "5", ObservedAt: "2026-03-30T01:30:00Z",
	}); err != nil {
		t.Fatal(err)
	}

	local, err := reconcile.LocalStateFromJournal(ctx, j, "acct-7")
	if err != nil {
		t.Fatal(err)
	}
	if qty, ok := local.Positions["AAPL"]; ok && qty != "0" {
		t.Fatalf("local position = %q, want nothing claimed from a refused snapshot", qty)
	}
}

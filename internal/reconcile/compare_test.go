package reconcile_test

import (
	"context"
	"path/filepath"
	"testing"

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

// openJournal opens a journal with the position projection bound.
//
// Binding the hook is not test scaffolding since task 6.3: the local half of the
// comparison *is* the projection (position-ledger: reconciliation의 로컬 상태는
// 이 투영을 소비한다 SHALL), so a journal recording fills without projecting them
// is a journal whose local belief is empty. Every test below that used to reach
// the fill ledger directly now reaches it through the row a fill wrote here.
func openJournal(t *testing.T) *journal.Journal {
	t.Helper()
	j := openJournalAt(t, filepath.Join(t.TempDir(), "journal.db"))
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
//
// Since task 6.3 the netting is the projection's, not a second sum computed
// here (position-ledger: fills-only 파생과 별도의 두 번째 포지션 계산을 두지
// 않는다 SHALL NOT). The number is the same; where it comes from is the point.
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
//
// Since task 6.3 the exclusion is structural rather than a filter in the query:
// the apply hooks are not called for a refused snapshot (apply_hook.go rule 3),
// so the projection never sees it and there is no row to exclude.
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

// --- the projection is the source (task 6.3) ---------------------------------

// TestLocalStateReadsTheProjectionNotTheFills is the rewiring itself: an
// adjustment moves the projection without touching a single fill, and the
// comparison has to follow the projection (position-ledger: 로컬 포지션 상태의
// 출처는 Position 투영이며 SHALL).
//
// The fill ledger still says 10 after this, so a local state derived from fills
// would keep reporting a disagreement the adjustment already settled — and the
// engine would stay blocked on a difference that no longer exists.
func TestLocalStateReadsTheProjectionNotTheFills(t *testing.T) {
	j := openJournal(t)
	ctx := context.Background()

	confirmedOrder(t, j, "intent-1", "attempt-1", "o-1", "AAPL", "BUY")
	if _, err := j.RecordFill(ctx, journal.FillObservation{
		OrderID: "o-1", Symbol: "AAPL", Market: "us", State: "FILLED", Terminal: true,
		Quantity: "10", FilledQuantity: "10", ObservedAt: "2026-03-30T01:30:00Z",
	}); err != nil {
		t.Fatal(err)
	}

	watermark, err := j.FillWatermark(ctx, "AAPL")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := j.ApplyPositionAdjustment(ctx, journal.AdjustmentRequest{
		AccountRef: "acct-7", Market: "us", Symbol: "AAPL", Kind: journal.AdjustmentUnknown,
		ExpectedPrevQuantity: "10", ExpectedFillWatermark: watermark, NewQuantity: "4",
		BrokerAsOf: "2026-03-30T01:31:00Z", Evidence: "the account says 4",
	}); err != nil {
		t.Fatalf("ApplyPositionAdjustment: %v", err)
	}

	// The fills are untouched — that is what makes this a real test of the source.
	net, err := j.NetPositions(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if net["AAPL"] != "10" {
		t.Fatalf("precondition: the fill ledger still says %q, want the untouched 10", net["AAPL"])
	}

	local, err := reconcile.LocalStateFromJournal(ctx, j, "acct-7")
	if err != nil {
		t.Fatal(err)
	}
	if local.Positions["AAPL"] != "4" {
		t.Fatalf("local position = %q, want the projection's adjusted 4", local.Positions["AAPL"])
	}
	// And the comparison it feeds now agrees with the account.
	diff := reconcile.Comparer{}.Compare(
		snapshotWith([]reconcile.Holding{{Symbol: "AAPL", Quantity: "4", Market: "us"}}, nil), local)
	if diff.BlocksEntry() {
		t.Fatalf("the adjustment converged the projection; the comparison must agree: %s", diff.Summary())
	}
}

// TestClosedInstancesAreNotHeld: CLOSED is final, so a closed instance
// contributes nothing to what the engine believes it holds. Counting it would
// report exposure on every symbol the engine has ever traded.
func TestClosedInstancesAreNotHeld(t *testing.T) {
	j := openJournal(t)
	ctx := context.Background()

	confirmedOrder(t, j, "intent-buy", "attempt-buy", "o-buy", "AAPL", "BUY")
	confirmedOrder(t, j, "intent-sell", "attempt-sell", "o-sell", "AAPL", "SELL")
	for _, obs := range []journal.FillObservation{
		{OrderID: "o-buy", Symbol: "AAPL", Market: "us", State: "FILLED", Terminal: true,
			Quantity: "10", FilledQuantity: "10", ObservedAt: "2026-03-30T01:30:00Z"},
		{OrderID: "o-sell", Symbol: "AAPL", Market: "us", State: "FILLED", Terminal: true,
			Quantity: "10", FilledQuantity: "10", ObservedAt: "2026-03-30T01:31:00Z"},
	} {
		if _, err := j.RecordFill(ctx, obs); err != nil {
			t.Fatal(err)
		}
	}

	current, err := j.CurrentPosition(ctx, "acct-7", "us", "AAPL")
	if err != nil {
		t.Fatal(err)
	}
	if current.State != journal.PositionClosed {
		t.Fatalf("precondition: state = %s, want CLOSED", current.State)
	}

	local, err := reconcile.LocalStateFromJournal(ctx, j, "acct-7")
	if err != nil {
		t.Fatal(err)
	}
	if qty, ok := local.Positions["AAPL"]; ok && qty != "0" {
		t.Fatalf("local position = %q, want a closed instance to hold nothing", qty)
	}
}

// TestTheProjectionIsSummedAtSymbolLevel is the delta's reduction rule
// (SHALL — 비교는 심볼 수준에서 수행하고 투영은 비-CLOSED 인스턴스의 합으로
// 축약한다). The holdings snapshot's market dimension is [미측정], so two
// instances of one symbol are one comparison unit until it is measured.
func TestTheProjectionIsSummedAtSymbolLevel(t *testing.T) {
	j := openJournal(t)
	ctx := context.Background()

	// One instance from a fill, a second from an adjustment on another market.
	confirmedOrder(t, j, "intent-1", "attempt-1", "o-1", "AAPL", "BUY")
	if _, err := j.RecordFill(ctx, journal.FillObservation{
		OrderID: "o-1", Symbol: "AAPL", Market: "us", State: "FILLED", Terminal: true,
		Quantity: "10", FilledQuantity: "10", ObservedAt: "2026-03-30T01:30:00Z",
	}); err != nil {
		t.Fatal(err)
	}
	watermark, err := j.FillWatermark(ctx, "AAPL")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := j.ApplyPositionAdjustment(ctx, journal.AdjustmentRequest{
		AccountRef: "acct-7", Market: "kr", Symbol: "AAPL", Kind: journal.AdjustmentExternal,
		ExpectedPrevQuantity: "0", ExpectedFillWatermark: watermark, NewQuantity: "2.5",
		BrokerAsOf: "2026-03-30T01:31:00Z", Evidence: "held on another venue",
	}); err != nil {
		t.Fatalf("ApplyPositionAdjustment: %v", err)
	}

	local, err := reconcile.LocalStateFromJournal(ctx, j, "acct-7")
	if err != nil {
		t.Fatal(err)
	}
	if local.Positions["AAPL"] != "12.5" {
		t.Fatalf("symbol total = %q, want the 12.5 both instances add up to", local.Positions["AAPL"])
	}
}

// TestAnotherAccountsProjectionIsNotThisAccountsBelief: the projection is keyed
// by account, and reading somebody else's rows would compare one account's
// snapshot against another's positions.
func TestAnotherAccountsProjectionIsNotThisAccountsBelief(t *testing.T) {
	j := openJournal(t)
	ctx := context.Background()

	confirmedOrder(t, j, "intent-1", "attempt-1", "o-1", "AAPL", "BUY")
	if _, err := j.RecordFill(ctx, journal.FillObservation{
		OrderID: "o-1", Symbol: "AAPL", Market: "us", State: "FILLED", Terminal: true,
		Quantity: "10", FilledQuantity: "10", ObservedAt: "2026-03-30T01:30:00Z",
	}); err != nil {
		t.Fatal(err)
	}

	local, err := reconcile.LocalStateFromJournal(ctx, j, "acct-other")
	if err != nil {
		t.Fatal(err)
	}
	if qty, ok := local.Positions["AAPL"]; ok && qty != "0" {
		t.Fatalf("acct-other believes it holds %q of a position acct-7 opened", qty)
	}
}

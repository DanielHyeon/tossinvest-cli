package reconcile_test

import (
	"context"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/reconcile"
)

// TestExternalOrderObservationCannotBecomeMissingLocalExposure follows the
// false-block path across cycles. Persisting the first broker observation must
// not make the order local when it later disappears from the broker open list.
func TestExternalOrderObservationCannotBecomeMissingLocalExposure(t *testing.T) {
	j := openJournal(t)
	ctx := context.Background()
	const orderID = "external-order"

	if _, err := j.RecordFill(ctx, journal.FillObservation{
		OrderID: orderID, Symbol: "AAPL", Market: "us", State: "OPEN_UNFILLED",
		Quantity: "10", FilledQuantity: "0", ObservedAt: "2026-03-30T01:30:00Z",
	}); err != nil {
		t.Fatal(err)
	}

	local, err := reconcile.LocalStateFromJournal(ctx, j, "acct-7")
	if err != nil {
		t.Fatal(err)
	}
	if len(local.OpenOrders) != 0 {
		t.Fatalf("local open orders = %+v, want the broker-only order to remain external", local.OpenOrders)
	}

	tracker := &reconcile.Tracker{AccountRef: "acct-7", MaxFailures: 3}
	first := reconcile.Comparer{}.Compare(reconcile.Snapshot{
		AccountRef: "acct-7",
		OpenOrders: []reconcile.BrokerOrder{{OrderID: orderID, Symbol: "AAPL", Side: "BUY", Quantity: "10"}},
	}, local)
	if first.BlocksEntry() || len(first.ExternalOrd) != 1 {
		t.Fatalf("first diff = %+v, want one non-blocking external order", first)
	}
	if _, err := tracker.Observe(ctx, first); err != nil {
		t.Fatal(err)
	}

	second := reconcile.Comparer{}.Compare(reconcile.Snapshot{AccountRef: "acct-7"}, local)
	if len(second.MissingOrders) != 0 || second.BlocksEntry() {
		t.Fatalf("second diff = %+v, want no missing local order after the external order disappears", second)
	}
	if _, err := tracker.Observe(ctx, second); err != nil {
		t.Fatal(err)
	}
	if tracker.Failures() != 0 {
		t.Fatalf("reconcile failures = %d, want 0 for an external order lifecycle", tracker.Failures())
	}
}

func TestAnotherAccountsLiveOrderCannotEnterThisAccountsLocalState(t *testing.T) {
	j := openJournal(t)
	ctx := context.Background()
	attempt, err := j.Prepare(ctx, journal.PrepareRequest{
		Intent: journal.Intent{
			ID: "other-account-intent", Market: "us", TradingDay: "2026-03-30",
			AccountRef: "acct-other", Symbol: "MSFT", Side: "BUY", OrderType: "LIMIT",
			Quantity: "2", Price: "400", Currency: "USD", Source: "engine",
			Fingerprint: "other-account-fingerprint",
		},
		Kind: journal.KindPlace, AttemptID: "other-account-attempt",
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := attempt.MarkDispatchStarted(ctx); err != nil {
		t.Fatal(err)
	}
	if err := attempt.MarkAcked(ctx, "other-account-order"); err != nil {
		t.Fatal(err)
	}
	if err := attempt.Settle(ctx, journal.StateConfirmed, journal.ReasonBrokerAcknowledged, "acked"); err != nil {
		t.Fatal(err)
	}
	if _, err := j.RecordFill(ctx, journal.FillObservation{
		OrderID: "other-account-order", Symbol: "MSFT", Market: "us", State: "OPEN_UNFILLED",
		Quantity: "2", FilledQuantity: "0", ObservedAt: "2026-03-30T01:30:00Z",
	}); err != nil {
		t.Fatal(err)
	}

	local, err := reconcile.LocalStateFromJournal(ctx, j, "acct-7")
	if err != nil {
		t.Fatal(err)
	}
	if len(local.OpenOrders) != 0 {
		t.Fatalf("local open orders = %+v, want another account excluded", local.OpenOrders)
	}
	other, err := reconcile.LocalStateFromJournal(ctx, j, "acct-other")
	if err != nil {
		t.Fatal(err)
	}
	if len(other.OpenOrders) != 1 {
		t.Fatalf("other account open orders = %+v, want its owned order", other.OpenOrders)
	}
}

func TestPriorTradingDayLineageCannotMakeRecoveryComparisonFalseClean(t *testing.T) {
	j := openJournal(t)
	ctx := context.Background()

	priorAmend, err := j.Prepare(ctx, journal.PrepareRequest{
		Intent: journal.Intent{
			ID: "prior-amend-intent", Market: "us", TradingDay: "2026-03-29",
			AccountRef: "acct-7", Symbol: "AAPL", Side: "BUY", OrderType: "LIMIT",
			Quantity: "1", Price: "100", Currency: "USD", Source: "engine",
			Fingerprint: "prior-amend-fingerprint",
		},
		Kind: journal.KindAmend, AttemptID: "prior-amend-attempt", TargetOrderID: "reused-parent",
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := priorAmend.MarkDispatchStarted(ctx); err != nil {
		t.Fatal(err)
	}
	if err := priorAmend.MarkAcked(ctx, "prior-child"); err != nil {
		t.Fatal(err)
	}
	if err := priorAmend.ResolveConfirmedWithLineage(ctx, journal.LineageEdge{
		ParentOrderID: "reused-parent", ChildOrderID: "prior-child",
		ParentFilledQuantity: "0", RequestedQuantity: "1",
	}, journal.ReasonResolvedFound, "prior-day replacement"); err != nil {
		t.Fatal(err)
	}

	current, err := j.Prepare(ctx, journal.PrepareRequest{
		Intent: journal.Intent{
			ID: "current-intent", Market: "us", TradingDay: "2026-03-30",
			AccountRef: "acct-7", Symbol: "AAPL", Side: "BUY", OrderType: "LIMIT",
			Quantity: "1", Price: "100", Currency: "USD", Source: "engine",
			Fingerprint: "current-fingerprint",
		},
		Kind: journal.KindPlace, AttemptID: "current-attempt",
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := current.MarkDispatchStarted(ctx); err != nil {
		t.Fatal(err)
	}
	if err := current.MarkAcked(ctx, "reused-parent"); err != nil {
		t.Fatal(err)
	}
	if err := current.Settle(ctx, journal.StateConfirmed, journal.ReasonBrokerAcknowledged, "current order"); err != nil {
		t.Fatal(err)
	}

	local, err := reconcile.LocalStateFromJournal(ctx, j, "acct-7")
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := local.OpenOrders["reused-parent"]; !ok {
		t.Fatalf("local open orders = %+v, want current-day parent to remain distinct", local.OpenOrders)
	}
	diff := reconcile.Comparer{}.Compare(reconcile.Snapshot{
		AccountRef: "acct-7",
		OpenOrders: []reconcile.BrokerOrder{{
			OrderID: "prior-child", Symbol: "AAPL", Side: "BUY", Quantity: "1",
		}},
	}, local)
	if !diff.BlocksEntry() || len(diff.MissingOrders) == 0 {
		t.Fatalf("diff = %+v, want the current-day parent missing rather than a false-clean release", diff)
	}
}

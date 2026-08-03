package reconcile_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/reconcile"
)

func TestOfficialOrderScopeIsPreservedInSnapshot(t *testing.T) {
	collector, _, orders, _, _ := newCollector(t)
	orders.pages = map[string]execgw.OrderPage{"": {Orders: []json.RawMessage{json.RawMessage(
		`{"orderId":"scoped-id","accountRef":"acct-7","market":"US","orderDate":"2026-03-30","orderedAt":"2026-03-31T00:30:00+09:00","symbol":"AAPL","side":"BUY","status":"OPEN","quantity":"1","price":"100","execution":{"filledQuantity":"0"}}`,
	)}}}

	snapshot, err := collector.Collect(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(snapshot.OpenOrders) != 1 {
		t.Fatalf("open orders = %+v, want one", snapshot.OpenOrders)
	}
	got := snapshot.OpenOrders[0]
	if got.AccountRef != "acct-7" || got.Market != "us" || got.TradingDay != "2026-03-30" ||
		got.OrderDate != "2026-03-30" || got.OrderedAt != "2026-03-31T00:30:00+09:00" {
		t.Fatalf("broker identity evidence = %+v, want raw account/market/date/time preserved", got)
	}
}

func TestOrderedAtWithoutMarketUsesCandidateMarketTradingDay(t *testing.T) {
	j := openJournal(t)
	confirmScopedOrder(t, j, "intent-us", "attempt-us", "acct-7", "us", "2026-03-30", "AAPL", "BUY", "reused-id")
	local, err := reconcile.LocalStateFromJournal(context.Background(), j, "acct-7")
	if err != nil {
		t.Fatal(err)
	}

	// The official endpoint timestamps orders in +09:00 but does not currently
	// promise a market field. This instant is March 31 in Seoul and March 30 in
	// New York, so the local candidate's market must define the canonical day.
	diff := reconcile.Comparer{}.Compare(reconcile.Snapshot{
		AccountRef: "acct-7",
		OpenOrders: []reconcile.BrokerOrder{{
			OrderID: "reused-id", TradingDay: "2026-03-31",
			OrderedAt: "2026-03-31T00:30:00+09:00", Symbol: "AAPL", Side: "BUY", Quantity: "1",
		}},
	}, local)
	if diff.BlocksEntry() || len(diff.ExternalOrd) != 0 {
		t.Fatalf("diff = %+v, want orderedAt interpreted in the candidate's US market", diff)
	}
}

func confirmScopedOrder(t *testing.T, j *journal.Journal, intentID, attemptID, account, market, day, symbol, side, orderID string) {
	t.Helper()
	ctx := context.Background()
	currency := "USD"
	if market == "kr" {
		currency = "KRW"
	}
	attempt, err := j.Prepare(ctx, journal.PrepareRequest{
		Intent: journal.Intent{
			ID: intentID, AccountRef: account, Market: market, TradingDay: day,
			Symbol: symbol, Side: side, OrderType: "LIMIT", Quantity: "1", Price: "100",
			Currency: currency, Source: "engine", Fingerprint: "fp-" + intentID,
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

// TestBrokerOrderInAnotherScopeCannotHideMissingOwnedOrder is the end-to-end
// journal projection regression. The opaque broker id is deliberately reused;
// symbol and side are sufficient evidence that the broker row is not the local
// order and therefore cannot make recovery false-clean.
func TestBrokerOrderInAnotherScopeCannotHideMissingOwnedOrder(t *testing.T) {
	j := openJournal(t)
	confirmScopedOrder(t, j, "intent-local", "attempt-local", "acct-7", "us", "2026-03-30", "AAPL", "BUY", "reused-id")

	local, err := reconcile.LocalStateFromJournal(context.Background(), j, "acct-7")
	if err != nil {
		t.Fatal(err)
	}
	diff := reconcile.Comparer{}.Compare(reconcile.Snapshot{
		AccountRef: "acct-7",
		OpenOrders: []reconcile.BrokerOrder{{
			OrderID: "reused-id", Symbol: "MSFT", Side: "SELL", Quantity: "1",
		}},
	}, local)

	if !diff.BlocksEntry() || len(diff.MissingOrders) != 1 || len(diff.ExternalOrd) != 1 {
		t.Fatalf("diff = %+v, want the foreign-scope row external and the owned order missing", diff)
	}
	if got := diff.MissingOrders[0]; got.Symbol != "AAPL" || got.Side != "BUY" || got.TradingDay != "2026-03-30" {
		t.Fatalf("missing identity = %+v, want the journal's complete canonical scope", got)
	}
}

func TestEveryCanonicalScopeDimensionRejectsAReusedID(t *testing.T) {
	j := openJournal(t)
	confirmScopedOrder(t, j, "intent-local", "attempt-local", "acct-7", "us", "2026-03-30", "AAPL", "BUY", "reused-id")
	local, err := reconcile.LocalStateFromJournal(context.Background(), j, "acct-7")
	if err != nil {
		t.Fatal(err)
	}

	base := reconcile.BrokerOrder{
		OrderID: "reused-id", AccountRef: "acct-7", Market: "us", TradingDay: "2026-03-30",
		Symbol: "AAPL", Side: "BUY", Quantity: "1",
	}
	tests := []struct {
		name   string
		mutate func(*reconcile.BrokerOrder)
	}{
		{name: "account", mutate: func(order *reconcile.BrokerOrder) { order.AccountRef = "acct-other" }},
		{name: "market", mutate: func(order *reconcile.BrokerOrder) { order.Market = "kr" }},
		{name: "trading day", mutate: func(order *reconcile.BrokerOrder) { order.TradingDay = "2026-03-31" }},
		{name: "symbol", mutate: func(order *reconcile.BrokerOrder) { order.Symbol = "MSFT" }},
		{name: "side", mutate: func(order *reconcile.BrokerOrder) { order.Side = "SELL" }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			broker := base
			test.mutate(&broker)
			diff := reconcile.Comparer{}.Compare(reconcile.Snapshot{
				AccountRef: "acct-7", OpenOrders: []reconcile.BrokerOrder{broker},
			}, local)
			if !diff.BlocksEntry() || len(diff.MissingOrders) != 1 || len(diff.ExternalOrd) != 1 {
				t.Fatalf("diff = %+v, want mismatched %s external plus owned order missing", diff, test.name)
			}
		})
	}
}

// TestPartialBrokerEvidenceCannotResolveTwoScopedOwners proves the second
// false-clean shape: when market/day evidence is absent, matching the shared id,
// symbol and side still leaves two market-scoped local owners. Ambiguity is not
// agreement; both local orders remain missing and the broker row stays external.
func TestPartialBrokerEvidenceCannotResolveTwoScopedOwners(t *testing.T) {
	j := openJournal(t)
	confirmScopedOrder(t, j, "intent-us", "attempt-us", "acct-7", "us", "2026-03-30", "AAPL", "BUY", "reused-id")
	confirmScopedOrder(t, j, "intent-kr", "attempt-kr", "acct-7", "kr", "2026-03-30", "AAPL", "BUY", "reused-id")

	local, err := reconcile.LocalStateFromJournal(context.Background(), j, "acct-7")
	if err != nil {
		t.Fatal(err)
	}
	diff := reconcile.Comparer{}.Compare(reconcile.Snapshot{
		AccountRef: "acct-7",
		OpenOrders: []reconcile.BrokerOrder{{
			OrderID: "reused-id", Symbol: "AAPL", Side: "BUY", Quantity: "1",
		}},
	}, local)

	if !diff.BlocksEntry() || len(diff.MissingOrders) != 2 || len(diff.ExternalOrd) != 1 {
		t.Fatalf("diff = %+v, want ambiguous broker evidence to remain external with both owners missing", diff)
	}
}

// TestCompleteBrokerScopeSelectsExactlyOneReusedIdentity keeps reuse usable when
// the official payload supplies enough evidence. The selected market/day agrees and
// only the other locally-owned order is missing.
func TestCompleteBrokerScopeSelectsExactlyOneReusedIdentity(t *testing.T) {
	j := openJournal(t)
	confirmScopedOrder(t, j, "intent-us", "attempt-us", "acct-7", "us", "2026-03-30", "AAPL", "BUY", "reused-id")
	confirmScopedOrder(t, j, "intent-kr", "attempt-kr", "acct-7", "kr", "2026-03-30", "AAPL", "BUY", "reused-id")

	local, err := reconcile.LocalStateFromJournal(context.Background(), j, "acct-7")
	if err != nil {
		t.Fatal(err)
	}
	diff := reconcile.Comparer{}.Compare(reconcile.Snapshot{
		AccountRef: "acct-7",
		OpenOrders: []reconcile.BrokerOrder{{
			OrderID: "reused-id", Market: "us", TradingDay: "2026-03-30",
			Symbol: "AAPL", Side: "BUY", Quantity: "1",
		}},
	}, local)

	if !diff.BlocksEntry() || len(diff.MissingOrders) != 1 || len(diff.ExternalOrd) != 0 {
		t.Fatalf("diff = %+v, want one exact scoped match and the other owner missing", diff)
	}
	if got := diff.MissingOrders[0].Market; got != "kr" {
		t.Fatalf("missing market = %q, want kr", got)
	}
}

func TestOlderNonterminalReusedTradingDayCannotDisappearFromLocalComparison(t *testing.T) {
	j := openJournal(t)
	confirmScopedOrder(t, j, "intent-old-day", "attempt-old-day", "acct-7", "us", "2026-03-29", "AAPL", "BUY", "active-reused-id")
	confirmScopedOrder(t, j, "intent-new-day", "attempt-new-day", "acct-7", "us", "2026-03-30", "AAPL", "BUY", "active-reused-id")

	local, err := reconcile.LocalStateFromJournal(context.Background(), j, "acct-7")
	if err != nil {
		t.Fatal(err)
	}
	if len(local.ScopedOpenOrders) != 2 {
		t.Fatalf("scoped local orders = %+v, want both active trading-day identities", local.ScopedOpenOrders)
	}
	diff := reconcile.Comparer{}.Compare(reconcile.Snapshot{
		AccountRef: "acct-7",
		OpenOrders: []reconcile.BrokerOrder{{
			OrderID: "active-reused-id", AccountRef: "acct-7", Market: "us",
			TradingDay: "2026-03-30", Symbol: "AAPL", Side: "BUY", Quantity: "1",
		}},
	}, local)
	if !diff.BlocksEntry() || len(diff.MissingOrders) != 1 ||
		diff.MissingOrders[0].TradingDay != "2026-03-29" {
		t.Fatalf("diff = %+v, want older active canonical order reported missing", diff)
	}
}

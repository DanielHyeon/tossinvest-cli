package execgw_test

// Query-fallback correction tests (extend-execution-contract task 2.4).
//
// Both corrections fix a way the P1 procedure reached a *confident* wrong
// answer, which is the only kind of wrong answer that matters here:
//
//   - a partially filled order is returned by both the OPEN and the CLOSED
//     query (the parameter is a group label and PARTIAL_FILLED is in both
//     groups — openapi), so the most common in-doubt case matched twice and
//     parked forever;
//   - the balance/holding cross-check compares against a pre-dispatch baseline,
//     so another mutation on the same symbol inside the window makes "nothing
//     moved" meaningless — and it was being read as proof of absence.

import (
	"context"
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
)

// TestPartialFillInBothListsResolvesByDedup is the spec's "부분 체결 주문의 양쪽
// 목록 출현": one order, two groups, one match.
func TestPartialFillInBothListsResolvesByDedup(t *testing.T) {
	f := newDoubtFixture(t)
	attemptID := f.placeInDoubt(t, &execgw.Baseline{BuyingPower: 1_000_000, Holding: 0, Currency: "KRW"})

	// The same order id, in both groups, exactly as the broker returns a
	// partially filled order.
	const partial = "O-partial"
	f.orders.set("OPEN", []string{
		orderJSON(partial, "005930", "BUY", "PARTIAL_FILLED", "2", "70000", "2026-03-30T10:30:00+09:00"),
	})
	f.orders.set("CLOSED", []string{
		orderJSON(partial, "005930", "BUY", "PARTIAL_FILLED", "2", "70000", "2026-03-30T10:30:00+09:00"),
	})

	res, err := resolveAsync(t, f.resolver(), f.clk, attemptID)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if res.State != journal.StateConfirmed {
		t.Fatalf("state: got %s, want CONFIRMED — the double listing is one order, not two (%s)",
			res.State, res.Detail)
	}
	if res.BrokerOrderID != partial {
		t.Errorf("broker order id: got %q, want %q", res.BrokerOrderID, partial)
	}
	// The detail still tells an operator the truth about where it was seen.
	if !strings.Contains(res.Detail, "OPEN+CLOSED") {
		t.Errorf("detail does not record the double listing: %s", res.Detail)
	}

	stored, err := f.journal.LookupAttempt(context.Background(), attemptID)
	if err != nil {
		t.Fatalf("LookupAttempt: %v", err)
	}
	if stored.State != journal.StateConfirmed || stored.BrokerOrderID != partial {
		t.Errorf("journal: state=%s brokerOrderID=%q", stored.State, stored.BrokerOrderID)
	}
}

// TestTwoDistinctOrdersStillPark: the dedup removes an artefact of how the
// question was asked, not the evidence that our model of the account is wrong.
// Two *different* ids matching one fingerprint is still a stop condition.
func TestTwoDistinctOrdersStillPark(t *testing.T) {
	f := newDoubtFixture(t)
	attemptID := f.placeInDoubt(t, &execgw.Baseline{BuyingPower: 1_000_000, Holding: 0, Currency: "KRW"})

	f.orders.set("OPEN", []string{
		orderJSON("O-one", "005930", "BUY", "PENDING", "2", "70000", "2026-03-30T10:30:00+09:00"),
		orderJSON("O-two", "005930", "BUY", "PENDING", "2", "70000", "2026-03-30T10:30:00+09:00"),
	})

	res, err := resolveAsync(t, f.resolver(), f.clk, attemptID)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if res.State != journal.StateUnresolvedInDoubt {
		t.Fatalf("state: got %s, want UNRESOLVED_IN_DOUBT (%s)", res.State, res.Detail)
	}
}

// TestContaminatedWindowParksInsteadOfFailing is the spec's "관측 창 오염": the
// lists show nothing and the account looks unchanged, but another mutation on
// the same symbol was dispatched while we were watching — so "unchanged" proves
// nothing, and the attempt must not be retired as FAILED_CONFIRMED.
func TestContaminatedWindowParksInsteadOfFailing(t *testing.T) {
	f := newDoubtFixture(t)
	attemptID := f.placeInDoubt(t, &execgw.Baseline{BuyingPower: 1_000_000, Holding: 0, Currency: "KRW"})

	// A second mutation on the same symbol, dispatched inside the window. It is
	// written straight to the journal because the gateway would refuse it — the
	// per-symbol latch is exactly what makes this rare — but a restart, an
	// operator or a saga can leave one behind, and the absence judgement has to
	// survive that.
	ctx := context.Background()
	other, err := f.journal.Prepare(ctx, journal.PrepareRequest{
		Intent: journal.Intent{
			ID: "intent-other", Market: "kr", TradingDay: "2026-03-30", AccountRef: "acct-7",
			Symbol: "005930", Side: "SELL", OrderType: "LIMIT", Quantity: "1", Price: "71000",
			Currency: "KRW", Source: "operator", Fingerprint: "fp-other",
		},
		Kind: journal.KindPlace, AttemptID: "attempt-other",
	})
	if err != nil {
		t.Fatalf("Prepare(other): %v", err)
	}
	if err := other.MarkDispatchStarted(ctx); err != nil {
		t.Fatalf("MarkDispatchStarted(other): %v", err)
	}
	if err := other.MarkAcked(ctx, "O-other"); err != nil {
		t.Fatalf("MarkAcked(other): %v", err)
	}
	if err := other.Settle(ctx, journal.StateConfirmed, "test", "someone else's order landed"); err != nil {
		t.Fatalf("Settle(other): %v", err)
	}

	res, err := resolveAsync(t, f.resolver(), f.clk, attemptID)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if res.State == journal.StateFailedConfirmed {
		t.Fatal("a contaminated observation window must never produce an automatic FAILED_CONFIRMED")
	}
	if res.State != journal.StateUnresolvedInDoubt {
		t.Fatalf("state: got %s, want UNRESOLVED_IN_DOUBT (%s)", res.State, res.Detail)
	}
	if !strings.Contains(res.Detail, "attempt-other") {
		t.Errorf("the park detail does not name the contaminating mutation: %s", res.Detail)
	}
	if f.gate.CheckEntry() == nil {
		t.Error("a parked attempt must latch the entry gate")
	}
}

// TestAnUnsentMutationDoesNotContaminate: a refusal raised after the dispatch
// transition committed carries a dispatch timestamp while provably having sent
// nothing. Counting it would make absence unprovable for no reason.
func TestAnUnsentMutationDoesNotContaminate(t *testing.T) {
	f := newDoubtFixture(t)
	attemptID := f.placeInDoubt(t, &execgw.Baseline{BuyingPower: 1_000_000, Holding: 0, Currency: "KRW"})

	ctx := context.Background()
	refused, err := f.journal.Prepare(ctx, journal.PrepareRequest{
		Intent: journal.Intent{
			ID: "intent-refused", Market: "kr", TradingDay: "2026-03-30", AccountRef: "acct-7",
			Symbol: "005930", Side: "BUY", OrderType: "LIMIT", Quantity: "1", Price: "70000",
			Currency: "KRW", Source: "engine", Fingerprint: "fp-refused",
		},
		Kind: journal.KindPlace, AttemptID: "attempt-refused",
	})
	if err != nil {
		t.Fatalf("Prepare(refused): %v", err)
	}
	if err := refused.MarkDispatchStarted(ctx); err != nil {
		t.Fatalf("MarkDispatchStarted(refused): %v", err)
	}
	if err := refused.Settle(ctx, journal.StateNotDispatched, "test", "refused before a byte left"); err != nil {
		t.Fatalf("Settle(refused): %v", err)
	}

	res, err := resolveAsync(t, f.resolver(), f.clk, attemptID)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if res.State != journal.StateFailedConfirmed {
		t.Fatalf("state: got %s, want FAILED_CONFIRMED — nothing else was ever sent (%s)",
			res.State, res.Detail)
	}
}

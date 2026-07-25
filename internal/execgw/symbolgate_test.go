package execgw_test

// symbolgate_test.go covers EntryGate's symbol dimension (task 4.2).

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/domain"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/orderintent"
	"github.com/JungHoonGhae/tossinvest-cli/internal/trading"
)

var gateNow = time.Date(2026, 7, 26, 9, 0, 0, 0, time.UTC)

// noStaleness builds a gate with no freshness requirements, so the tests observe
// latch behaviour and nothing else.
func noStaleness(t *testing.T) (*execgw.EntryGate, *clock.Fake) {
	t.Helper()
	clk := clock.NewFake(gateNow)
	return execgw.NewEntryGate(clk, map[execgw.RequiredQuery]time.Duration{}), clk
}

// TestSymbolBlockStopsOnlyThatSymbol is the whole point of the dimension: before
// it existed, a disagreement about one symbol had to latch the whole account.
func TestSymbolBlockStopsOnlyThatSymbol(t *testing.T) {
	gate, _ := noStaleness(t)
	gate.BlockSymbol("us", "AAPL", execgw.ReasonBrokerStateUnknown, "cumulative fill went backwards")

	rejected := gate.CheckEntryFor("us", "AAPL")
	if rejected == nil || rejected.Reason != execgw.ReasonBrokerStateUnknown {
		t.Fatalf("AAPL = %v, want blocked with unknown_broker_state", rejected)
	}
	if rejected.Detail != "cumulative fill went backwards" {
		t.Errorf("detail = %q, want the observation that raised it", rejected.Detail)
	}
	if rejected := gate.CheckEntryFor("us", "MSFT"); rejected != nil {
		t.Errorf("MSFT = %v, want tradable", rejected)
	}
	if rejected := gate.CheckEntryFor("kr", "AAPL"); rejected != nil {
		t.Errorf("a kr-market block was never raised, got %v", rejected)
	}
}

// TestCheckEntryStaysAccountWide pins the compatibility half: every existing
// caller of CheckEntry() keeps its existing answer, because the narrowing happens
// by asking a narrower question rather than by changing the old one.
func TestCheckEntryStaysAccountWide(t *testing.T) {
	gate, _ := noStaleness(t)
	gate.BlockSymbol("us", "AAPL", execgw.ReasonReconcileMismatch, "quantity disagreement")

	if rejected := gate.CheckEntry(); rejected != nil {
		t.Errorf("CheckEntry() = %v; a symbol block is not an account block", rejected)
	}

	gate.Block(execgw.ReasonBrokerAuthRejected, "401 from the broker")
	if rejected := gate.CheckEntry(); rejected == nil {
		t.Error("CheckEntry() must still report account-wide latches")
	}
}

// TestAccountBlockCoversEveryTradableSymbol: an account-wide latch is still
// account-wide when asked the narrow question.
func TestAccountBlockCoversEveryTradableSymbol(t *testing.T) {
	gate, _ := noStaleness(t)
	gate.Block(execgw.ReasonBrokerAuthRejected, "401 from the broker")

	for _, symbol := range []string{"AAPL", "MSFT", "005930"} {
		rejected := gate.CheckEntryFor("us", symbol)
		if rejected == nil || rejected.Reason != execgw.ReasonBrokerAuthRejected {
			t.Errorf("%s = %v, want the account-wide latch", symbol, rejected)
		}
	}
}

// TestAccountReasonWinsOverSymbolReason: the widest applicable reason is the
// actionable one. Fixing "005930 disagrees" changes nothing while the credential
// is rejected.
func TestAccountReasonWinsOverSymbolReason(t *testing.T) {
	gate, _ := noStaleness(t)
	gate.BlockSymbol("kr", "005930", execgw.ReasonReconcileMismatch, "symbol level")
	gate.Block(execgw.ReasonBrokerAuthRejected, "account level")

	rejected := gate.CheckEntryFor("kr", "005930")
	if rejected == nil || rejected.Reason != execgw.ReasonBrokerAuthRejected {
		t.Fatalf("reason = %v, want the account-wide one", rejected)
	}
}

// TestMarketlessBlockAppliesEverywhere. A producer that knows the symbol but not
// the market (reconcile's quantity mismatch is one) must still block it.
func TestMarketlessBlockAppliesEverywhere(t *testing.T) {
	gate, _ := noStaleness(t)
	gate.BlockSymbol("", "AAPL", execgw.ReasonReconcileMismatch, "no market known")

	for _, market := range []string{"us", "kr", ""} {
		if rejected := gate.CheckEntryFor(market, "AAPL"); rejected == nil {
			t.Errorf("market %q: want AAPL blocked", market)
		}
	}
	if rejected := gate.CheckEntryFor("us", "MSFT"); rejected != nil {
		t.Errorf("MSFT = %v, want tradable", rejected)
	}
}

// TestBlockSymbolWithoutASymbolLatchesTheAccount. A symbol block with no symbol
// would silently block nothing, and the caller observed *something*.
func TestBlockSymbolWithoutASymbolLatchesTheAccount(t *testing.T) {
	gate, _ := noStaleness(t)
	gate.BlockSymbol("us", "  ", execgw.ReasonBrokerStateUnknown, "no symbol in the payload")

	if rejected := gate.CheckEntry(); rejected == nil {
		t.Fatal("want the account latched rather than a block that stops nothing")
	}
}

// TestReblockingKeepsTheOriginalObservation: a block whose "since" advances every
// poll looks new forever, and an operator cannot tell how long it has been there.
func TestReblockingKeepsTheOriginalObservation(t *testing.T) {
	gate, clk := noStaleness(t)
	gate.BlockSymbol("us", "AAPL", execgw.ReasonReconcileMismatch, "first observation")

	clk.Advance(time.Hour)
	gate.BlockSymbol("us", "AAPL", execgw.ReasonReconcileMismatch, "second observation")

	blocks := gate.SymbolBlocks()
	if len(blocks) != 1 {
		t.Fatalf("blocks = %+v, want one", blocks)
	}
	if !blocks[0].Since.Equal(gateNow) {
		t.Errorf("since = %v, want the first observation at %v", blocks[0].Since, gateNow)
	}
	if blocks[0].Detail != "first observation" {
		t.Errorf("detail = %q, want the first observation", blocks[0].Detail)
	}
}

// TestClearSymbolReleasesOnlyThatBlock.
func TestClearSymbolReleasesOnlyThatBlock(t *testing.T) {
	gate, _ := noStaleness(t)
	gate.BlockSymbol("us", "AAPL", execgw.ReasonReconcileMismatch, "a")
	gate.BlockSymbol("us", "MSFT", execgw.ReasonReconcileMismatch, "b")
	gate.BlockSymbol("us", "AAPL", execgw.ReasonBrokerStateUnknown, "c")

	gate.ClearSymbol("us", "AAPL", execgw.ReasonReconcileMismatch)

	if rejected := gate.CheckEntryFor("us", "AAPL"); rejected == nil ||
		rejected.Reason != execgw.ReasonBrokerStateUnknown {
		t.Errorf("AAPL = %v, want the remaining unknown-state block", rejected)
	}
	if rejected := gate.CheckEntryFor("us", "MSFT"); rejected == nil {
		t.Error("MSFT's own block must survive")
	}
}

// TestClearSymbolReason releases a whole reason across symbols, which is what a
// clean sweep does.
func TestClearSymbolReason(t *testing.T) {
	gate, _ := noStaleness(t)
	gate.BlockSymbol("us", "AAPL", execgw.ReasonReconcileMismatch, "a")
	gate.BlockSymbol("us", "MSFT", execgw.ReasonReconcileMismatch, "b")
	gate.BlockSymbol("us", "AAPL", execgw.ReasonBrokerStateUnknown, "c")

	gate.ClearSymbolReason(execgw.ReasonReconcileMismatch)

	if rejected := gate.CheckEntryFor("us", "MSFT"); rejected != nil {
		t.Errorf("MSFT = %v, want released", rejected)
	}
	if rejected := gate.CheckEntryFor("us", "AAPL"); rejected == nil ||
		rejected.Reason != execgw.ReasonBrokerStateUnknown {
		t.Errorf("AAPL = %v, want the untouched reason still blocking", rejected)
	}
}

// TestStalenessStillAppliesToTheNarrowQuestion: freshness is an account property,
// so a stale required read blocks every symbol.
func TestStalenessStillAppliesToTheNarrowQuestion(t *testing.T) {
	clk := clock.NewFake(gateNow)
	gate := execgw.NewEntryGate(clk, map[execgw.RequiredQuery]time.Duration{
		execgw.QueryOpenOrders: 20 * time.Second,
	})

	if rejected := gate.CheckEntryFor("us", "AAPL"); rejected == nil ||
		rejected.Reason != execgw.ReasonQueryStale {
		t.Fatalf("a never-observed read must block, got %v", rejected)
	}

	gate.RecordSuccess(execgw.QueryOpenOrders)
	if rejected := gate.CheckEntryFor("us", "AAPL"); rejected != nil {
		t.Fatalf("a fresh read must release, got %v", rejected)
	}

	clk.Advance(21 * time.Second)
	if rejected := gate.CheckEntryFor("us", "AAPL"); rejected == nil {
		t.Fatal("a stale read must block again")
	}
}

// TestGatewayAsksTheSymbolQuestion is the wiring half of task 4.2: a block on one
// symbol must refuse a place in that symbol and permit one in another, through
// the gateway rather than only at the gate.
func TestGatewayAsksTheSymbolQuestion(t *testing.T) {
	clk := clock.NewFake(fixedNow)
	j := openJournal(t, clk)
	gate := execgw.NewEntryGate(clk, map[execgw.RequiredQuery]time.Duration{})
	broker := &fakeBroker{result: domain.MutationResult{Kind: "place", Status: "accepted", OrderID: "O-1"}}

	gw, err := execgw.New(execgw.Options{
		Journal:    j,
		Trading:    trading.NewService(openPolicy(), broker),
		Clock:      clk,
		AccountRef: "acct-7",
		Source:     "test",
		Entry:      gate,
	})
	if err != nil {
		t.Fatalf("execgw.New: %v", err)
	}

	gate.BlockSymbol("kr", "005930", execgw.ReasonBrokerStateUnknown, "the broker snapshot contradicted itself")

	blocked := placeIntent() // 005930, kr
	_, err = gw.Place(context.Background(), execgw.PlaceRequest{
		Intent:   blocked,
		Decision: goodDecision(t, execgw.PlaceHash(blocked), clk),
	})
	var rejected *execgw.RejectedError
	if !errors.As(err, &rejected) || rejected.Reason != execgw.ReasonBrokerStateUnknown {
		t.Fatalf("blocked symbol: err = %v, want a refusal with unknown_broker_state", err)
	}

	other := placeIntent()
	other.Symbol = "000660"
	out, err := gw.Place(context.Background(), execgw.PlaceRequest{
		Intent:   other,
		Decision: goodDecision(t, execgw.PlaceHash(other), clk),
	})
	if err != nil {
		t.Fatalf("an unrelated symbol must still trade: %v", err)
	}
	if out.State != journal.StateConfirmed {
		t.Errorf("state = %s, want CONFIRMED (%s)", out.State, out.Detail)
	}
}

// TestGatewayNeverGatesAnExitOnASymbolBlock is §0.3 for the new dimension: a
// symbol can be shut for entries and must stay open for cancels.
func TestGatewayNeverGatesAnExitOnASymbolBlock(t *testing.T) {
	clk := clock.NewFake(fixedNow)
	j := openJournal(t, clk)
	gate := execgw.NewEntryGate(clk, map[execgw.RequiredQuery]time.Duration{})
	broker := &fakeBroker{result: domain.MutationResult{Kind: "cancel", Status: "accepted", OrderID: "O-9"}}

	gw, err := execgw.New(execgw.Options{
		Journal:    j,
		Trading:    trading.NewService(openPolicy(), broker),
		Clock:      clk,
		AccountRef: "acct-7",
		Source:     "test",
		Entry:      gate,
	})
	if err != nil {
		t.Fatalf("execgw.New: %v", err)
	}
	gate.BlockSymbol("kr", "005930", execgw.ReasonBrokerStateUnknown, "shut for entries")

	intent := orderintent.CancelIntent{OrderID: "O-1", Symbol: "005930"}
	out, err := gw.Cancel(context.Background(), execgw.CancelRequest{
		Intent: intent,
		Order:  execgw.OrderRef{Market: "kr", Side: "BUY", Quantity: 2, Price: 70000, Currency: "KRW"},
		Decision: execgw.GuardianDecision{
			Nonce:      "cancel-nonce",
			IntentHash: execgw.CancelHash(intent),
			IssuedAt:   clk.Now(),
			ExpiresAt:  clk.Now().Add(30 * time.Second),
			Authority:  "test-guardian",
		},
	})
	if err != nil {
		t.Fatalf("a blocked symbol must still be exitable (§0.3): %v", err)
	}
	if out.State != journal.StateConfirmed {
		t.Errorf("state = %s, want CONFIRMED (%s)", out.State, out.Detail)
	}
}

// TestSymbolBlocksAreConcurrencySafe covers the -race requirement: fill
// detection, reconciliation and the gateway all touch this gate from different
// goroutines.
func TestSymbolBlocksAreConcurrencySafe(t *testing.T) {
	gate, _ := noStaleness(t)

	done := make(chan struct{})
	for i := 0; i < 4; i++ {
		go func(i int) {
			defer func() { done <- struct{}{} }()
			for n := 0; n < 200; n++ {
				gate.BlockSymbol("us", "AAPL", execgw.ReasonReconcileMismatch, "x")
				gate.CheckEntryFor("us", "AAPL")
				gate.SymbolBlocks()
				gate.ClearSymbol("us", "AAPL", execgw.ReasonReconcileMismatch)
			}
		}(i)
	}
	for i := 0; i < 4; i++ {
		<-done
	}
}

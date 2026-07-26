package reconcile_test

import (
	"context"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/config"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/orderintent"
	"github.com/JungHoonGhae/tossinvest-cli/internal/reconcile"
	"github.com/JungHoonGhae/tossinvest-cli/internal/trading"
)

// Mismatch-handling tests (harden-execution-base task 3.6).
//
// The requirements: entries block and exits stay open, re-reconciliation waits
// out a minimum interval, three consecutive failures make the mismatch permanent,
// a success resets the counter, and every block has a scope and a documented
// release condition.

func mismatchDiff(symbol, local, broker string) reconcile.Diff {
	return reconcile.Diff{
		AccountRef: "acct-7",
		Quantities: []reconcile.QuantityMismatch{
			{Symbol: symbol, Local: local, Broker: broker},
		},
	}
}

func newTracker(clk clock.Clock, gate *execgw.EntryGate) *reconcile.Tracker {
	return &reconcile.Tracker{
		Clock:       clk,
		Gate:        gate,
		MinInterval: 30 * time.Second,
		MaxFailures: 3,
		AccountRef:  "acct-7",
	}
}

// TestQuantityMismatchBlocksEntries is the spec's "수량 불일치 감지" scenario.
func TestQuantityMismatchBlocksEntries(t *testing.T) {
	clk := clock.NewFake(asOf)
	gate := execgw.NewEntryGate(clk, map[execgw.RequiredQuery]time.Duration{})
	tracker := newTracker(clk, gate)

	out := tracker.Observe(mismatchDiff("AAPL", "10", "4"))
	if !out.Blocked || out.Failures != 1 {
		t.Fatalf("outcome = %+v, want a block on the first failure", out)
	}
	if len(out.Added) != 1 || out.Added[0].Scope != reconcile.ScopeSymbol {
		t.Fatalf("added = %+v, want one symbol-scoped block", out.Added)
	}
	if out.Added[0].Symbol != "AAPL" {
		t.Fatalf("block symbol = %q, want AAPL", out.Added[0].Symbol)
	}
	// The block is symbol-scoped, and since task 4.2 the gate can say so: AAPL
	// is shut, everything else on the account keeps trading. Before the gate had
	// a symbol dimension this had to latch the whole account (issues.md,
	// 2026-07-26) — safe, but it meant one disagreement stopped the engine
	// entirely.
	rejected := gate.CheckEntryFor("us", "AAPL")
	if rejected == nil || rejected.Reason != execgw.ReasonReconcileMismatch {
		t.Fatalf("AAPL entries must be blocked, got %v", rejected)
	}
	if rejected := gate.CheckEntryFor("us", "MSFT"); rejected != nil {
		t.Fatalf("an unrelated symbol must stay tradable, got %v", rejected)
	}
	if rejected := gate.CheckEntry(); rejected != nil {
		t.Fatalf("the account-wide question must stay open for a symbol-scoped block, got %v", rejected)
	}
	if !tracker.ExitAllowed("us", "AAPL") {
		t.Fatal("exits must stay open (§0.3)")
	}
}

// TestSuccessResetsTheFailureCounter: the counter measures *consecutive* failure,
// which is what separates a stuck account from a busy one.
func TestSuccessResetsTheFailureCounter(t *testing.T) {
	clk := clock.NewFake(asOf)
	gate := execgw.NewEntryGate(clk, map[execgw.RequiredQuery]time.Duration{})
	tracker := newTracker(clk, gate)

	tracker.Observe(mismatchDiff("AAPL", "10", "4"))
	clk.Advance(30 * time.Second)
	tracker.Observe(mismatchDiff("AAPL", "10", "4"))
	if tracker.Failures() != 2 {
		t.Fatalf("failures = %d, want 2", tracker.Failures())
	}

	clk.Advance(30 * time.Second)
	out := tracker.Observe(reconcile.Diff{AccountRef: "acct-7", Matched: 1})
	if out.Failures != 0 {
		t.Fatalf("failures after a clean reconcile = %d, want 0", out.Failures)
	}
	if len(out.Cleared) != 1 {
		t.Fatalf("cleared = %+v, want the mismatch block released", out.Cleared)
	}
	if rejected := gate.CheckEntry(); rejected != nil {
		t.Fatalf("a clean reconcile must release the gate, got %v", rejected)
	}

	// And a fresh failure starts counting from one again, not from three.
	clk.Advance(30 * time.Second)
	if out := tracker.Observe(mismatchDiff("AAPL", "10", "4")); out.Failures != 1 {
		t.Fatalf("failures after the reset = %d, want 1", out.Failures)
	}
}

// TestThreeConsecutiveFailuresBecomePermanent: "look again" has now been shown
// not to fix it, so looping forever is not a plan.
func TestThreeConsecutiveFailuresBecomePermanent(t *testing.T) {
	clk := clock.NewFake(asOf)
	gate := execgw.NewEntryGate(clk, map[execgw.RequiredQuery]time.Duration{})
	tracker := newTracker(clk, gate)

	var out reconcile.Outcome
	for i := 0; i < 3; i++ {
		out = tracker.Observe(mismatchDiff("AAPL", "10", "4"))
		clk.Advance(30 * time.Second)
	}
	if !out.Permanent || out.Failures != 3 {
		t.Fatalf("outcome after three failures = %+v, want permanent", out)
	}

	blocks := tracker.Blocks()
	var permanent *reconcile.Block
	for i := range blocks {
		if blocks[i].Permanent {
			permanent = &blocks[i]
			break
		}
	}
	if permanent == nil {
		t.Fatalf("blocks = %+v, want a permanent one", blocks)
	}
	if permanent.Scope != reconcile.ScopeAccount {
		t.Fatalf("permanent scope = %q, want account — the reconciliation process is what is broken",
			permanent.Scope)
	}
	if permanent.Release != reconcile.ReleaseOperatorOnly {
		t.Fatalf("release = %q, want operator only", permanent.Release)
	}
	rejected := gate.CheckEntry()
	if rejected == nil || rejected.Reason != execgw.ReasonReconcilePermanent {
		t.Fatalf("gate reason = %v, want the permanent one to take precedence", rejected)
	}
}

// TestACleanReconcileDoesNotReleaseAPermanentMismatch: the whole point of
// "permanent" is that observation has stopped being the answer.
func TestACleanReconcileDoesNotReleaseAPermanentMismatch(t *testing.T) {
	clk := clock.NewFake(asOf)
	gate := execgw.NewEntryGate(clk, map[execgw.RequiredQuery]time.Duration{})
	tracker := newTracker(clk, gate)

	for i := 0; i < 3; i++ {
		tracker.Observe(mismatchDiff("AAPL", "10", "4"))
		clk.Advance(30 * time.Second)
	}
	tracker.Observe(reconcile.Diff{AccountRef: "acct-7", Matched: 1})

	if !tracker.Permanent() {
		t.Fatal("a clean reconcile must not un-mark a permanent mismatch")
	}
	rejected := gate.CheckEntry()
	if rejected == nil || rejected.Reason != execgw.ReasonReconcilePermanent {
		t.Fatalf("gate = %v, want the permanent block still held", rejected)
	}
}

// TestOperatorResolutionClearsAPermanentMismatch, and demands the audit trail
// that makes overriding the machine accountable.
func TestOperatorResolutionClearsAPermanentMismatch(t *testing.T) {
	clk := clock.NewFake(asOf)
	gate := execgw.NewEntryGate(clk, map[execgw.RequiredQuery]time.Duration{})
	tracker := newTracker(clk, gate)

	for i := 0; i < 3; i++ {
		tracker.Observe(mismatchDiff("AAPL", "10", "4"))
		clk.Advance(30 * time.Second)
	}

	if err := tracker.Resolve("", "checked"); err == nil {
		t.Fatal("resolving without an identity must be refused")
	}
	if err := tracker.Resolve("daniel", ""); err == nil {
		t.Fatal("resolving without a note must be refused")
	}
	if err := tracker.Resolve("daniel", "compared the app to the journal by hand"); err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if tracker.Permanent() || len(tracker.Blocks()) != 0 {
		t.Fatalf("after an operator resolution: permanent=%v blocks=%+v",
			tracker.Permanent(), tracker.Blocks())
	}
	if rejected := gate.CheckEntry(); rejected != nil {
		t.Fatalf("the gate must reopen, got %v", rejected)
	}
}

// TestReReconciliationWaitsOutTheMinimumInterval: reconciling in a tight loop
// against a settlement in flight burns the query budget and makes the
// fill-detection reads stale, which blocks entries for a second reason and hides
// the first.
func TestReReconciliationWaitsOutTheMinimumInterval(t *testing.T) {
	clk := clock.NewFake(asOf)
	tracker := newTracker(clk, nil)

	if !tracker.Due(clk.Now()) {
		t.Fatal("a tracker that has never reconciled must be due immediately")
	}
	tracker.Observe(mismatchDiff("AAPL", "10", "4"))

	if tracker.Due(clk.Now()) {
		t.Fatal("a reconciliation must not be due immediately after one ran")
	}
	clk.Advance(29 * time.Second)
	if tracker.Due(clk.Now()) {
		t.Fatal("29s into a 30s interval is not due")
	}
	clk.Advance(time.Second)
	if !tracker.Due(clk.Now()) {
		t.Fatal("the interval elapsed; the reconciliation is due")
	}
	if want := asOf.Add(30 * time.Second); !tracker.NextDueAt().Equal(want) {
		t.Fatalf("NextDueAt = %s, want %s", tracker.NextDueAt(), want)
	}
}

// TestDefaultIntervalMatchesTheSpec pins the documented default rather than
// leaving it to whatever the last edit happened to leave behind.
func TestDefaultIntervalMatchesTheSpec(t *testing.T) {
	if reconcile.DefaultReconcileInterval != 30*time.Second {
		t.Fatalf("default reconcile interval = %s, want the spec's 30s",
			reconcile.DefaultReconcileInterval)
	}
	if reconcile.DefaultMaxFailures != 3 {
		t.Fatalf("default max failures = %d, want the spec's 3", reconcile.DefaultMaxFailures)
	}
	if reconcile.DefaultStabilisationInterval != 2*time.Second {
		t.Fatalf("default stabilisation interval = %s, want the spec's 2s",
			reconcile.DefaultStabilisationInterval)
	}
}

// TestBlockScopes: a symbol block stops that symbol, an account block stops
// everything, and an unrelated symbol is unaffected by the narrow one.
func TestBlockScopes(t *testing.T) {
	clk := clock.NewFake(asOf)
	tracker := newTracker(clk, nil)
	tracker.Observe(mismatchDiff("AAPL", "10", "4"))

	if rejected := tracker.EntryAllowed("us", "AAPL"); rejected == nil {
		t.Fatal("the mismatched symbol must be blocked")
	}
	if rejected := tracker.EntryAllowed("us", "MSFT"); rejected != nil {
		t.Fatalf("an unrelated symbol must not be blocked by a symbol-scoped mismatch: %v", rejected)
	}

	// Escalation to account scope stops everything.
	clk.Advance(30 * time.Second)
	tracker.Observe(mismatchDiff("AAPL", "10", "4"))
	clk.Advance(30 * time.Second)
	tracker.Observe(mismatchDiff("AAPL", "10", "4"))

	if rejected := tracker.EntryAllowed("us", "MSFT"); rejected == nil {
		t.Fatal("a permanent, account-scoped mismatch must block every symbol")
	}
	if !tracker.ExitAllowed("us", "MSFT") {
		t.Fatal("exits stay open even under an account-scoped block (§0.3)")
	}
}

// TestEveryReasonHasAStateTableRow is the spec's "차단 범위와 해제 조건은
// reason-code와 함께 상태표로 정의한다", enforced: a block whose release nobody
// decided is how an engine ends up permanently and inexplicably shut.
func TestEveryReasonHasAStateTableRow(t *testing.T) {
	for _, reason := range []execgw.ReasonCode{
		execgw.ReasonRecoveryIncomplete,
		execgw.ReasonReconcileMismatch,
		execgw.ReasonReconcilePermanent,
		execgw.ReasonBrokerStateUnknown,
	} {
		rule, ok := reconcile.RuleFor(reason)
		if !ok {
			t.Fatalf("reason %q has no row in the state table", reason)
		}
		if rule.Scope == "" || rule.Auto == "" || rule.Condition == "" {
			t.Fatalf("the row for %q is incomplete: %+v", reason, rule)
		}
	}
}

// TestStateTableRowsAreDistinct catches a copy-paste that would make two
// conditions share a release policy by accident.
func TestStateTableRowsAreDistinct(t *testing.T) {
	seen := map[execgw.ReasonCode]bool{}
	for _, rule := range reconcile.BlockRules() {
		if seen[rule.Reason] {
			t.Fatalf("reason %q appears twice in the state table", rule.Reason)
		}
		seen[rule.Reason] = true
	}
	if len(seen) != 4 {
		t.Fatalf("state table has %d rows, want the four documented conditions", len(seen))
	}
}

// TestMissingOrderRaisesASymbolBlock covers the other blocking finding.
func TestMissingOrderRaisesASymbolBlock(t *testing.T) {
	clk := clock.NewFake(asOf)
	tracker := newTracker(clk, nil)
	out := tracker.Observe(reconcile.Diff{
		AccountRef: "acct-7",
		MissingOrders: []reconcile.LocalOrder{
			{OrderID: "o-1", Symbol: "AAPL", Market: "us"},
		},
	})
	if len(out.Added) != 1 || out.Added[0].Symbol != "AAPL" {
		t.Fatalf("added = %+v, want a symbol block for AAPL", out.Added)
	}
	if out.Added[0].Reason != execgw.ReasonReconcileMismatch {
		t.Fatalf("reason = %q, want reconciliation_mismatch", out.Added[0].Reason)
	}
}

// TestExternalFindingsDoNotBlock: the owner trading their own account by hand is
// not a malfunction, and blocking on it would make the engine unusable next to
// its owner.
func TestExternalFindingsDoNotBlock(t *testing.T) {
	clk := clock.NewFake(asOf)
	gate := execgw.NewEntryGate(clk, map[execgw.RequiredQuery]time.Duration{})
	tracker := newTracker(clk, gate)

	out := tracker.Observe(reconcile.Diff{
		AccountRef: "acct-7",
		ExternalOrd: []reconcile.ExternalOrder{
			{BrokerOrder: reconcile.BrokerOrder{OrderID: "manual-1", Symbol: "TSLA"},
				Provenance: reconcile.ProvenanceExternal},
		},
		ExternalPos: []reconcile.ExternalPosition{
			{Holding: reconcile.Holding{Symbol: "TSLA", Quantity: "3"},
				Provenance: reconcile.ProvenanceExternal},
		},
	})
	if out.Blocked || out.Failures != 0 {
		t.Fatalf("outcome = %+v, want external findings not to block or count as failure", out)
	}
	if rejected := gate.CheckEntry(); rejected != nil {
		t.Fatalf("the gate must stay open, got %v", rejected)
	}
}

// TestGatewayKeepsExitsOpenUnderAMismatch is §0.3 driven through the real
// gateway: with the mismatch latch held, a buy is refused and a cancel is not.
func TestGatewayKeepsExitsOpenUnderAMismatch(t *testing.T) {
	j := openJournal(t)
	clk := clock.NewFake(asOf)
	gate := execgw.NewEntryGate(clk, map[execgw.RequiredQuery]time.Duration{})
	tracker := newTracker(clk, gate)
	broker := &stubBroker{}

	gw, err := execgw.New(execgw.Options{
		Journal: j,
		Trading: trading.NewService(config.Trading{
			Place: true, Sell: true, Cancel: true, Amend: true, AllowLiveOrderActions: true,
		}, broker),
		Clock: clk, AccountRef: "acct-7", Source: "test", Entry: gate,
	})
	if err != nil {
		t.Fatal(err)
	}

	tracker.Observe(mismatchDiff("AAPL", "10", "4"))

	buy := orderintent.PlaceIntent{
		Symbol: "AAPL", Market: "us", Side: "buy", OrderType: "limit",
		Quantity: 1, Price: 200, CurrencyMode: "USD",
	}
	// The issuer records each decision before the gateway is called; the gateway
	// verifies against the row, never against these values.
	issuer := &execgw.Issuer{Journal: j, Clock: clk, AccountRef: "acct-7", TTL: time.Minute}
	buyDecision, err := issuer.IssueEntry(context.Background(), execgw.EntryRequest{
		Market: "us", Symbol: "AAPL", Side: "buy", Quantity: 1, EntryPrice: 200,
		StopPrice: 180, PolicyVersion: "test/v1",
		Limits: execgw.Limits{MaxQuantity: 10, MaxNotional: 100000, Currency: "USD"},
	})
	if err != nil {
		t.Fatal(err)
	}
	out, err := gw.Place(context.Background(), execgw.PlaceRequest{
		Intent:   buy,
		Decision: buyDecision,
	})
	if err == nil {
		t.Fatal("a buy under a mismatch must be refused")
	}
	if out.Reason != execgw.ReasonReconcileMismatch {
		t.Fatalf("reason = %q, want reconciliation_mismatch", out.Reason)
	}
	if broker.places != 0 {
		t.Fatalf("broker places = %d, want 0", broker.places)
	}

	// The exit is not gated: liquidation is how the disagreement gets smaller.
	cancelIntent := orderintent.CancelIntent{OrderID: "broker-1", Symbol: "AAPL"}
	cancelDecision, err := issuer.IssueReduction(context.Background(), execgw.ReductionRequest{
		Kind: journal.KindCancel, Market: "us", Symbol: "AAPL", Side: "BUY",
		MaxQuantity: 1, Reason: "exit under a mismatch",
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := gw.Cancel(context.Background(), execgw.CancelRequest{
		Intent:   cancelIntent,
		Order:    execgw.OrderRef{Market: "us", Side: "BUY", Quantity: 1, Price: 200, Currency: "USD"},
		Decision: cancelDecision,
	}); err != nil {
		t.Fatalf("a cancel under a mismatch must go through: %v", err)
	}
	if broker.cancels != 1 {
		t.Fatalf("broker cancels = %d, want 1", broker.cancels)
	}

	// A sell reduces exposure too, and must not be gated either.
	sell := orderintent.PlaceIntent{
		Symbol: "MSFT", Market: "us", Side: "sell", OrderType: "limit",
		Quantity: 1, Price: 400, CurrencyMode: "USD",
	}
	sellDecision, err := issuer.IssueReduction(context.Background(), execgw.ReductionRequest{
		Kind: journal.KindPlace, Market: "us", Symbol: "MSFT", Side: "SELL",
		MaxQuantity: 1, Reason: "exit under a mismatch",
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := gw.Place(context.Background(), execgw.PlaceRequest{
		Intent:   sell,
		Decision: sellDecision,
	}); err != nil {
		t.Fatalf("a sell reduces exposure and must not be gated: %v", err)
	}
	if broker.places != 1 {
		t.Fatalf("broker places = %d, want the sell to have gone through", broker.places)
	}
}

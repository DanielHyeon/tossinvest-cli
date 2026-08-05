package engine_test

// reconcileloop_test.go covers the reconciliation driver loop and the adoption
// judgement it ends with (change adopt-external-positions tasks 2.1, 2.2, 2.3
// and 2.5).
//
// Every test drives RunOnce rather than Run, because what is under test is the
// cycle — with one exception, which exists to prove the loop stops when its
// context does.
//
// The journal is real and the exit policy is real. Only the broker surfaces are
// faked, and the fake holdings reader deliberately implements the *raw* path so
// the cost basis that reaches `position_adoptions` is the string the broker
// spelled rather than a re-rendered float.

import (
	"context"
	"encoding/json"
	"errors"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/app/engine"
	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/config"
	"github.com/JungHoonGhae/tossinvest-cli/internal/domain"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
	"github.com/JungHoonGhae/tossinvest-cli/internal/reconcile"
)

var reconcileLoopNow = time.Date(2026, 3, 30, 1, 0, 0, 0, time.UTC)

const reconcileAccount = "acct-reconcile"

// --- broker fakes -------------------------------------------------------------

// fakeOrderPager answers the open-order half of a snapshot. One page, because
// what the loop's tests are about is the cycle and not the pagination the
// collector already has its own tests for.
type fakeOrderPager struct {
	orders []json.RawMessage
	err    error
	calls  int
}

func (f *fakeOrderPager) OrdersPageRaw(context.Context, execgw.OrderQuery, string) (execgw.OrderPage, error) {
	f.calls++
	if f.err != nil {
		return execgw.OrderPage{}, f.err
	}
	return execgw.OrderPage{Orders: f.orders}, nil
}

// fakeHoldings answers the holdings half, through the raw path. It satisfies
// both PositionsReader and RawPositionsReader, so the collector takes the raw
// one and CostBasisRaw survives to the adoption record.
type fakeHoldings struct {
	items []reconcile.RawHolding
	err   error
	calls int
}

func (f *fakeHoldings) Positions(context.Context) ([]domain.Position, error) {
	f.calls++
	if f.err != nil {
		return nil, f.err
	}
	out := make([]domain.Position, 0, len(f.items))
	for _, h := range f.items {
		out = append(out, domain.Position{Symbol: h.Symbol, MarketType: h.Market})
	}
	return out, nil
}

func (f *fakeHoldings) PositionsRaw(context.Context) ([]reconcile.RawHolding, error) {
	f.calls++
	if f.err != nil {
		return nil, f.err
	}
	return append([]reconcile.RawHolding(nil), f.items...), nil
}

type fakeBalance struct{ err error }

func (f *fakeBalance) BuyingPower(context.Context, string) (float64, error) {
	if f.err != nil {
		return 0, f.err
	}
	return 1_000_000, nil
}

type refusingReconcileRelease struct {
	reconcile.ReconcileStore
}

func (s *refusingReconcileRelease) ReleaseReconcile(context.Context,
	journal.ReleaseReconcileRequest) (journal.ReconcileState, bool, error) {
	return journal.ReconcileState{}, false, nil
}

// --- harness ------------------------------------------------------------------

type driverHarness struct {
	t        *testing.T
	journal  *journal.Journal
	clk      *clock.Fake
	orders   *fakeOrderPager
	holdings *fakeHoldings
	balance  *fakeBalance
	prices   *fakePrices
	alerts   *fakeAlerts
	tracker  *reconcile.Tracker
	driver   *engine.ReconcileDriver
}

func newDriverHarness(t *testing.T, mutate func(*engine.ReconcileDriverOptions)) *driverHarness {
	t.Helper()
	clk := clock.NewFake(reconcileLoopNow)
	j, err := journal.Open(context.Background(), journal.Options{
		Path:     filepath.Join(t.TempDir(), "journal.db"),
		Clock:    clk,
		FSProber: journal.FixedFSProber(journal.FSInfo{Name: "ext4", Magic: journal.MagicExt}),
	})
	if err != nil {
		t.Fatalf("journal.Open: %v", err)
	}
	t.Cleanup(func() { _ = j.Close() })
	if err := j.SetApplyHooks(journal.ApplyHooks{
		Project: journal.ProjectPosition, Exit: journal.ApplyExitFill,
	}); err != nil {
		t.Fatalf("SetApplyHooks: %v", err)
	}

	gate := execgw.NewEntryGate(clk, nil)
	tracker := &reconcile.Tracker{Clock: clk, Gate: gate, Journal: j, AccountRef: reconcileAccount}

	h := &driverHarness{
		t: t, journal: j, clk: clk,
		orders:   &fakeOrderPager{},
		holdings: &fakeHoldings{},
		balance:  &fakeBalance{},
		prices:   &fakePrices{last: map[string]float64{}},
		alerts:   &fakeAlerts{},
		tracker:  tracker,
	}

	opts := engine.ReconcileDriverOptions{
		Journal: j,
		Collector: &reconcile.Collector{
			Orders: h.orders, Positions: h.holdings, Balance: h.balance,
			Currencies: []string{"KRW"}, AccountRef: reconcileAccount, Clock: clk,
		},
		Tracker: tracker,
		Ingest: &reconcile.Ingestor{
			Journal: j, AccountRef: reconcileAccount, DefaultMarket: "kr",
		},
		Converge: &reconcile.Converger{
			Journal: j, Credit: tracker, AccountRef: reconcileAccount,
		},
		Prices:        h.prices,
		Alerts:        h.alerts,
		AccountRef:    reconcileAccount,
		Clock:         clk,
		DefaultMarket: "kr",
		Adoption: config.Adoption{
			Enabled: true, DefaultStopPct: 0.05,
		},
	}
	if mutate != nil {
		mutate(&opts)
	}
	driver, err := engine.NewReconcileDriver(opts)
	if err != nil {
		t.Fatalf("NewReconcileDriver: %v", err)
	}
	h.driver = driver
	return h
}

// holds puts a holding in the account and a price on the tape.
func (h *driverHarness) holds(symbol, quantity, costBasis string, last float64) {
	h.holdsMarket("kr", symbol, quantity, costBasis, last, "KRW")
}

func (h *driverHarness) holdsMarket(market, symbol, quantity, costBasis string, last float64,
	currency string) {
	h.holdings.items = append(h.holdings.items, reconcile.RawHolding{
		Symbol: symbol, Market: market, Quantity: quantity, AveragePrice: costBasis,
	})
	h.prices.last[symbol] = last
	if h.prices.currencies == nil {
		h.prices.currencies = map[string]string{}
	}
	h.prices.currencies[symbol] = currency
}

// cycle runs one pass, advancing the fake clock through the stabilisation wait.
//
// The collector stamps as-of from the same fake clock, so the two snapshots are
// genuinely the stabilisation interval apart and the Stabiliser counts them —
// which is what makes these tests exercise the real two-collection rule rather
// than a version of it with the interval removed.
func (h *driverHarness) cycle() engine.ReconcileCycle {
	h.t.Helper()
	done := make(chan engine.ReconcileCycle, 1)
	go func() { done <- h.driver.RunOnce(context.Background()) }()
	waitForSleeper(h.t, h.clk)
	h.clk.Advance(reconcile.DefaultStabilisationInterval)
	return <-done
}

// waitForSleeper blocks until the driver has entered its stabilisation wait, so
// the fake clock's advance lands on a loop that is actually waiting rather than
// racing past it.
func waitForSleeper(t *testing.T, clk *clock.Fake) {
	t.Helper()
	if !clk.WaitForSleepers(1, 5*time.Second) {
		t.Fatal("the driver never entered its stabilisation wait")
	}
}

func (h *driverHarness) position(symbol string) journal.Position {
	return h.positionMarket("kr", symbol)
}

func (h *driverHarness) positionMarket(market, symbol string) journal.Position {
	h.t.Helper()
	p, err := h.journal.CurrentPosition(context.Background(), reconcileAccount, market, symbol)
	if err != nil {
		h.t.Fatalf("CurrentPosition(%s): %v", symbol, err)
	}
	return p
}

// --- the loop -----------------------------------------------------------------

// TestTheDriverFoldsAndAdoptsInOneCycle is the whole path: a holding the engine
// never bought is folded into the projection, adopted, and given an exit state
// seeded from the price observed at adoption.
func TestTheDriverFoldsAndAdoptsInOneCycle(t *testing.T) {
	h := newDriverHarness(t, nil)
	h.holds("005930", "10", "55000.0000", 70000)

	cycle := h.cycle()
	if cycle.Err != nil {
		t.Fatalf("cycle: %v", cycle.Err)
	}
	if cycle.Collected != 2 {
		t.Errorf("collections = %d, want 2 (the Stabiliser needs two)", cycle.Collected)
	}
	if !cycle.Stable {
		t.Fatal("two identical snapshots an interval apart must be stable")
	}
	if cycle.Folded != 1 || cycle.Adopted != 1 {
		t.Fatalf("folded = %d, adopted = %d, want 1 and 1", cycle.Folded, cycle.Adopted)
	}
	if cycle.Unmanaged != 0 {
		t.Errorf("unmanaged = %d; an adopted holding is managed and must not also be reported "+
			"as unprotected", cycle.Unmanaged)
	}

	p := h.position("005930")
	if !p.Adopted() || !p.ExitEligible() {
		t.Fatalf("position after the cycle = %+v, want adopted and eligible", p)
	}
	adoption, err := h.journal.AdoptionOf(context.Background(), p.ID)
	if err != nil {
		t.Fatalf("AdoptionOf: %v", err)
	}
	if adoption.ObservedPrice != "70000" {
		t.Errorf("observed price = %q, want the tape's 70000", adoption.ObservedPrice)
	}
	if adoption.SyntheticStop != "66500" {
		t.Errorf("synthetic stop = %q, want 70000 × (1 − 0.05)", adoption.SyntheticStop)
	}
	// The broker's raw string, trailing zeros and all. It never enters the R
	// formula; it is stored so the 2b fee measurement has the original to work
	// from.
	if adoption.CostBasis != "55000.0000" {
		t.Errorf("cost basis = %q, want the broker's raw string", adoption.CostBasis)
	}
	if adoption.CostBasisSource != journal.CostBasisBrokerAvg {
		t.Errorf("cost basis source = %q", adoption.CostBasisSource)
	}

	state, err := h.journal.ExitState(context.Background(), p.ID)
	if err != nil {
		t.Fatalf("ExitState: %v", err)
	}
	if state.EntryPrice != "70000" || state.InitialStop != "66500" || state.Baseline != "66500" {
		t.Errorf("exit state t0 = %+v, want the manage-forward pair", state)
	}
	if state.HighWater != "70000" || state.RatchetLevel != journal.RatchetNone {
		t.Errorf("exit state at t0 = %+v, want R=0 with the watermark at the entry", state)
	}

	// The adoption event replaces the unmanaged alert (design A4).
	if got := h.alerts.count(obs.EventExitPositionAdopted); got != 1 {
		t.Errorf("adoption events = %d, want 1", got)
	}
	if got := h.alerts.count(obs.EventExitPositionUnmanaged); got != 0 {
		t.Errorf("unmanaged alerts = %d, want 0: the adoption event replaces it", got)
	}
}

// TestTheAdoptionTransactionProposesNothing is design A2's SHALL NOT and the
// first half of task 2.3: adopting protects a position, it never sells one.
func TestTheAdoptionTransactionProposesNothing(t *testing.T) {
	h := newDriverHarness(t, nil)
	// Deep under water against the cost basis, and far above it in the other
	// test below. Either way the adoption itself proposes nothing: a cost-basis
	// anchored t0 would liquidate this on the spot.
	h.holds("005930", "10", "140000", 70000)

	cycle := h.cycle()
	if cycle.Adopted != 1 {
		t.Fatalf("adopted = %d, want 1 (%v)", cycle.Adopted, cycle.Err)
	}

	p := h.position("005930")
	events, err := h.journal.ExitEvents(context.Background(), p.ID)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range events {
		if strings.TrimSpace(e.Action) != "" && e.Action != journal.ExitEventOpened {
			t.Errorf("the adoption transaction recorded %q; it must record the opening and nothing "+
				"else — no exit judgement is part of adopting", e.Action)
		}
		if strings.TrimSpace(e.ProposedIntentID) != "" {
			t.Errorf("the adoption transaction proposed intent %q; adopting proposes nothing",
				e.ProposedIntentID)
		}
	}
	state, err := h.journal.ExitState(context.Background(), p.ID)
	if err != nil {
		t.Fatal(err)
	}
	if state.Pending() {
		t.Errorf("the adoption armed a proposal: %+v", state)
	}
	if state.TakenRatioTotal != "0" {
		t.Errorf("taken ratio = %q after adoption, want 0", state.TakenRatioTotal)
	}
}

// TestAdoptionIsSilentUnderReconcile is the transition-state rule: a symbol the
// engine and the account disagree about is neither adopted nor alerted on, and
// the next cycle re-evaluates it.
func TestAdoptionIsSilentUnderReconcile(t *testing.T) {
	h := newDriverHarness(t, nil)
	h.holds("005930", "10", "55000", 70000)

	// A durable symbol block, restored the way a real one would be.
	if _, _, err := h.journal.EnterReconcile(context.Background(), journal.EnterReconcileRequest{
		AccountRef: reconcileAccount, Symbol: "005930",
		Cause: journal.ReconcileCauseQuantityMismatch, Evidence: "local 3, broker 4",
	}); err != nil {
		t.Fatalf("EnterReconcile: %v", err)
	}
	cycle := h.cycle()
	if cycle.Adopted != 0 {
		t.Errorf("adopted = %d under RECONCILE; adopting into a disagreement freezes a t0 against "+
			"a quantity nobody agrees on", cycle.Adopted)
	}
	if cycle.Unmanaged != 0 {
		t.Errorf("unmanaged = %d; RECONCILE is a transition state and transition states are silent",
			cycle.Unmanaged)
	}
	if got := h.alerts.count(obs.EventExitPositionUnmanaged); got != 0 {
		t.Errorf("unmanaged alerts = %d, want 0 during a transition state", got)
	}
}

func TestActiveForeignCauseIsCountedAndBlocksBeforePriceRead(t *testing.T) {
	h := newDriverHarness(t, nil)
	h.holds("005930", "10", "55000", 70000)
	if _, _, err := h.journal.EnterReconcile(context.Background(), journal.EnterReconcileRequest{
		AccountRef: reconcileAccount, Symbol: "005930",
		Cause: journal.ReconcileCauseIdentifierConflict, Evidence: "identifier conflict",
	}); err != nil {
		t.Fatalf("EnterReconcile: %v", err)
	}
	cycle := h.cycle()
	if cycle.Err != nil {
		t.Fatalf("cycle: %v", cycle.Err)
	}
	if cycle.Blocked != 1 {
		t.Fatalf("blocked = %d, want the one active durable block (not only new additions)", cycle.Blocked)
	}
	if cycle.Adopted != 0 || h.prices.calls != 0 {
		t.Fatalf("adopted = %d, price reads = %d under identifier conflict", cycle.Adopted, h.prices.calls)
	}
}

func TestTrackerReleaseFailureStopsBeforePriceAndAdoption(t *testing.T) {
	h := newDriverHarness(t, nil)
	h.holds("005930", "10", "55000", 70000)
	if _, _, err := h.journal.EnterReconcile(context.Background(), journal.EnterReconcileRequest{
		AccountRef: reconcileAccount, Symbol: "005930",
		Cause: journal.ReconcileCauseQuantityMismatch, Evidence: "quantity differs",
	}); err != nil {
		t.Fatalf("EnterReconcile: %v", err)
	}
	if err := h.tracker.Restore(context.Background()); err != nil {
		t.Fatalf("Restore: %v", err)
	}
	h.tracker.Journal = &refusingReconcileRelease{ReconcileStore: h.journal}
	// Stamped before the cycle's own comparison so the credit is genuinely
	// spendable by it; what this test asserts is the refused durable release, not
	// the a083 fail-closed stamp check.
	h.tracker.AdjustmentApplied(h.clk.Now().UTC().Format(time.RFC3339), "005930")

	cycle := h.cycle()
	if cycle.Err == nil {
		t.Fatal("cycle must fail when the durable release is refused")
	}
	if cycle.Blocked != 1 || cycle.Released != 0 {
		t.Fatalf("cycle = %+v, want the active block preserved and no visible release", cycle)
	}
	if cycle.Adopted != 0 || cycle.Unmanaged != 0 || h.prices.calls != 0 {
		t.Fatalf("cycle continued after tracker failure: adopted=%d unmanaged=%d prices=%d",
			cycle.Adopted, cycle.Unmanaged, h.prices.calls)
	}
}

func TestIncludeOnlyAdoptionRequiresAPriceReader(t *testing.T) {
	clk := clock.NewFake(reconcileLoopNow)
	j, err := journal.Open(context.Background(), journal.Options{
		Path:     filepath.Join(t.TempDir(), "journal.db"),
		Clock:    clk,
		FSProber: journal.FixedFSProber(journal.FSInfo{Name: "ext4", Magic: journal.MagicExt}),
	})
	if err != nil {
		t.Fatalf("journal.Open: %v", err)
	}
	t.Cleanup(func() { _ = j.Close() })

	_, err = engine.NewReconcileDriver(engine.ReconcileDriverOptions{
		Journal: j, Collector: &reconcile.Collector{}, Tracker: &reconcile.Tracker{},
		Ingest: &reconcile.Ingestor{}, Converge: &reconcile.Converger{},
		AccountRef: reconcileAccount,
		Adoption:   config.Adoption{IncludeSymbols: []string{"005930"}, DefaultStopPct: 0.05},
	})
	if !errors.Is(err, engine.ErrReconcileDriverUnavailable) {
		t.Fatalf("NewReconcileDriver include-only without prices = %v, want unavailable", err)
	}
}

// TestAnUnstableAccountDefersSilently: two disagreeing snapshots are the owner
// trading right now, which is a transition state and not a finding.
func TestAnUnstableAccountDefersSilently(t *testing.T) {
	h := newDriverHarness(t, nil)
	h.holds("005930", "10", "55000", 70000)

	done := make(chan engine.ReconcileCycle, 1)
	go func() { done <- h.driver.RunOnce(context.Background()) }()
	waitForSleeper(h.t, h.clk)
	// The account moves between the two collections.
	h.holdings.items[0].Quantity = "12"
	h.clk.Advance(reconcile.DefaultStabilisationInterval)
	cycle := <-done

	if cycle.Stable {
		t.Fatal("two disagreeing snapshots must not be stable")
	}
	if cycle.Adopted != 0 || cycle.Unmanaged != 0 {
		t.Errorf("cycle = %+v, want nothing adopted and nothing reported", cycle)
	}
	if len(h.alerts.events) != 0 {
		t.Errorf("an unstable account raised %d alert(s); Stabiliser non-convergence is a transition "+
			"state", len(h.alerts.events))
	}
}

// TestTheUnmanagedAlertSurvivesAdoptionBeingOff is design A4's regression guard
// and §0.2: the toggle decides whether the engine *acts*, never whether the
// operator is told the engine is trading beside a position it will not protect.
func TestTheUnmanagedAlertSurvivesAdoptionBeingOff(t *testing.T) {
	h := newDriverHarness(t, func(o *engine.ReconcileDriverOptions) {
		o.Adoption = config.Adoption{}
	})
	h.holds("005930", "10", "55000", 70000)

	cycle := h.cycle()
	if cycle.Adopted != 0 {
		t.Fatalf("adopted = %d with adoption off", cycle.Adopted)
	}
	if cycle.Unmanaged != 1 {
		t.Fatalf("unmanaged = %d, want 1", cycle.Unmanaged)
	}
	alert, ok := h.alerts.first(obs.EventExitPositionUnmanaged)
	if !ok {
		t.Fatal("adoption being off silenced the unmanaged alert; that is the §0.2 regression " +
			"design A4 exists to prevent")
	}
	if alert.Fields["adoption_enabled"] != false {
		t.Errorf("the alert must say whether adoption was on: %+v", alert.Fields)
	}

	// Latched: the census still counts the holding every cycle, but the operator
	// is told once. An alert that repeats every minute is an alert nobody reads.
	if second := h.cycle(); second.Unmanaged != 1 {
		t.Errorf("unmanaged census = %d on the second cycle, want the holding still counted",
			second.Unmanaged)
	}
	if got := h.alerts.count(obs.EventExitPositionUnmanaged); got != 1 {
		t.Errorf("unmanaged alerts after two cycles = %d, want 1: the alert is latched per position", got)
	}
}

// TestAnExcludedSymbolIsReportedNotAdopted: the exclusion list is fine-grained
// control inside enabled, and an excluded holding is still unprotected.
func TestAnExcludedSymbolIsReportedNotAdopted(t *testing.T) {
	h := newDriverHarness(t, func(o *engine.ReconcileDriverOptions) {
		o.Adoption = config.Adoption{
			Enabled: true, DefaultStopPct: 0.05, ExcludeSymbols: []string{"005930"},
		}
	})
	h.holds("005930", "10", "55000", 70000)

	cycle := h.cycle()
	if cycle.Adopted != 0 {
		t.Errorf("adopted = %d, want 0 for an excluded symbol", cycle.Adopted)
	}
	if cycle.Unmanaged != 1 {
		t.Errorf("unmanaged = %d, want 1: an excluded holding is deliberately unprotected and the "+
			"operator is still told", cycle.Unmanaged)
	}
	alert, ok := h.alerts.first(obs.EventExitPositionUnmanaged)
	if !ok {
		t.Fatal("no unmanaged alert for the excluded symbol")
	}
	if !strings.Contains(alert.Body, "exclude_symbols") {
		t.Errorf("the alert must name the reason: %q", alert.Body)
	}
	// And no price was spent on a symbol that was never a candidate.
	if h.prices.calls != 0 {
		t.Errorf("price reads = %d for an excluded symbol; the batch must carry candidates only",
			h.prices.calls)
	}
}

// TestAdoptionPricesEveryCandidateInOneCall is design A6's batching rule: a
// fan-out that scaled with the portfolio would make the §0.4 budget a function
// of a number nobody bounded.
func TestAdoptionPricesEveryCandidateInOneCall(t *testing.T) {
	h := newDriverHarness(t, nil)
	h.holds("005930", "10", "55000", 70000)
	h.holds("000660", "5", "120000", 150000)
	h.holds("035420", "3", "200000", 210000)

	cycle := h.cycle()
	if cycle.Adopted != 3 {
		t.Fatalf("adopted = %d, want 3 (%v)", cycle.Adopted, cycle.Err)
	}
	if h.prices.calls != 1 {
		t.Errorf("price reads = %d for three candidates, want exactly 1", h.prices.calls)
	}
	if len(h.prices.asked) != 1 || len(h.prices.asked[0]) != 3 {
		t.Errorf("the batch asked %v, want all three symbols together", h.prices.asked)
	}
}

// TestACandidateWithNoQuoteIsDeferred: there is no t0 to freeze, so the holding
// waits rather than being adopted against a price nobody observed.
func TestACandidateWithNoQuoteIsDeferred(t *testing.T) {
	h := newDriverHarness(t, nil)
	h.holds("005930", "10", "55000", 70000)
	delete(h.prices.last, "005930")

	cycle := h.cycle()
	if cycle.Adopted != 0 || cycle.Deferred != 1 {
		t.Fatalf("cycle = %+v, want nothing adopted and one deferred", cycle)
	}
	p := h.position("005930")
	if p.Adopted() {
		t.Error("a holding with no observation was adopted anyway")
	}
	if _, err := h.journal.ExitState(context.Background(), p.ID); !errors.Is(err, journal.ErrExitStateNotFound) {
		t.Errorf("an exit state exists for a deferred candidate: %v", err)
	}
	// A candidate the engine could not adopt is a position it is not protecting.
	if got := h.alerts.count(obs.EventExitPositionUnmanaged); got != 1 {
		t.Errorf("unmanaged alerts = %d, want 1 for a failed adoption", got)
	}
}

// TestAStalePriceDefersTheAdoption is the exit-policy SHALL: an observation
// older than the bound is not a t0.
//
// The staleness is produced the way a real one would be — a price read that took
// longer than the bound — rather than by shrinking the bound to nothing.
func TestAStalePriceDefersTheAdoption(t *testing.T) {
	var h *driverHarness
	h = newDriverHarness(t, func(o *engine.ReconcileDriverOptions) {
		o.Prices = &slowPrices{
			inner: &fakePrices{last: map[string]float64{"005930": 70000}},
			delay: engine.DefaultAdoptionPriceStaleness + time.Second,
			clock: func() *clock.Fake { return h.clk },
		}
	})
	h.holds("005930", "10", "55000", 70000)

	cycle := h.cycle()
	if cycle.Adopted != 0 {
		t.Fatalf("adopted = %d against a stale observation; a synthetic stop frozen from a price "+
			"the market has left behind is an instant liquidation", cycle.Adopted)
	}
	if h.position("005930").Adopted() {
		t.Error("the position was adopted from a stale observation")
	}
}

// TestReReconciliationRecognisesAnAdoptedPosition is the crash-recovery and
// re-reconciliation half: a second cycle over the same holding neither
// re-adopts it nor trips the fold guard.
func TestReReconciliationRecognisesAnAdoptedPosition(t *testing.T) {
	h := newDriverHarness(t, nil)
	h.holds("005930", "10", "55000", 70000)

	first := h.cycle()
	if first.Adopted != 1 {
		t.Fatalf("first cycle adopted = %d (%v)", first.Adopted, first.Err)
	}
	adoptionID := h.position("005930").AdoptionID

	second := h.cycle()
	if second.Err != nil {
		t.Fatalf("the second cycle failed over an already-adopted position: %v", second.Err)
	}
	if second.Adopted != 0 {
		t.Errorf("adopted = %d on re-reconciliation, want 0", second.Adopted)
	}
	if got := h.position("005930").AdoptionID; got != adoptionID {
		t.Errorf("adoption id moved from %q to %q on re-reconciliation", adoptionID, got)
	}
	if h.prices.calls != 1 {
		t.Errorf("price reads = %d; an already-adopted holding is not a candidate", h.prices.calls)
	}
}

// TestCrashBetweenAdoptionAndExitStateIsCompleted: the adoption commits first,
// so a process that dies before the exit state opens leaves a record the next
// pass finishes from — without re-adopting.
//
// The crash window is built rather than simulated by deleting rows: the holding
// is folded with adoption off (so a position exists and is unmanaged), the
// adoption transaction is then run on its own, and that is exactly the on-disk
// state a process dying between the two commits leaves behind.
func TestCrashBetweenAdoptionAndExitStateIsCompleted(t *testing.T) {
	ctx := context.Background()
	h := newDriverHarness(t, func(o *engine.ReconcileDriverOptions) {
		o.Adoption = config.Adoption{}
	})
	h.holds("005930", "10", "55000", 70000)
	if cycle := h.cycle(); cycle.Folded != 1 {
		t.Fatalf("folded = %d (%v)", cycle.Folded, cycle.Err)
	}
	p := h.position("005930")

	adoption, err := h.journal.AdoptPosition(ctx, journal.AdoptionRequest{
		PositionID: p.ID, Symbol: p.Symbol, Market: p.Market, Quantity: p.Quantity,
		CostBasis: "55000", ObservedPrice: "70000", SyntheticStop: "66500",
		ObservedAt: journal.RFC3339(reconcileLoopNow),
	})
	if err != nil {
		t.Fatalf("AdoptPosition: %v", err)
	}
	// … and here the process dies. The record and the pointer are durable; the
	// exit state does not exist.
	if _, err := h.journal.ExitState(ctx, p.ID); !errors.Is(err, journal.ErrExitStateNotFound) {
		t.Fatalf("the crash window must leave no exit state: %v", err)
	}

	state, err := h.journal.OpenAdoptedExitState(ctx, p.ID)
	if err != nil {
		t.Fatalf("recovering the exit state from the persisted adoption: %v", err)
	}
	if state.EntryPrice != "70000" || state.InitialStop != "66500" {
		t.Errorf("the recovered t0 = %+v, want the values the adoption record froze", state)
	}
	if got := h.position("005930").AdoptionID; got != adoption.ID {
		t.Errorf("recovery moved the adoption id to %q; the record on disk is the one to complete", got)
	}
}

// TestAPartialSnapshotEndsTheCycle: a snapshot is all or none, and a cycle that
// could not read the account writes nothing.
func TestAPartialSnapshotEndsTheCycle(t *testing.T) {
	h := newDriverHarness(t, nil)
	h.holds("005930", "10", "55000", 70000)
	h.balance.err = errors.New("balance unavailable")

	cycle := h.driver.RunOnce(context.Background())
	if cycle.Err == nil {
		t.Fatal("a partial snapshot must fail the cycle")
	}
	if cycle.Stable || cycle.Folded != 0 || cycle.Adopted != 0 {
		t.Errorf("cycle = %+v, want nothing written", cycle)
	}
	if !errors.Is(cycle.Err, reconcile.ErrPartialSnapshot) {
		t.Errorf("cycle error = %v, want ErrPartialSnapshot", cycle.Err)
	}
}

// TestTheDriverCountsConsecutiveCycleFailures is the supervision seam
// (add-engine-runtime task 1.3): the loop still retries every period — that
// decision is untouched — but a loop that is alive and getting nowhere is now
// distinguishable from a healthy one, which is what the runtime's degradation
// threshold reads.
func TestTheDriverCountsConsecutiveCycleFailures(t *testing.T) {
	h := newDriverHarness(t, nil)
	h.holds("005930", "10", "55000", 70000)

	if got := h.driver.Health(); got.Consecutive != 0 || got.Cycles != 0 {
		t.Fatalf("a fresh driver reports %+v", got)
	}

	h.balance.err = errors.New("balance unavailable")
	for i := 1; i <= 3; i++ {
		_ = h.driver.RunOnce(context.Background())
		health := h.driver.Health()
		if health.Consecutive != i {
			t.Fatalf("after %d failed cycle(s) the count is %d", i, health.Consecutive)
		}
		if health.LastError == nil || health.Since.IsZero() {
			t.Fatalf("a failing run reports %+v with no error or start", health)
		}
		if health.Cycles != i {
			t.Fatalf("cycles = %d, want %d — every cycle is counted, failed or not", health.Cycles, i)
		}
	}
	if got := h.driver.ConsecutiveFailures(); got != 3 {
		t.Fatalf("ConsecutiveFailures = %d, want 3 — the runtime reads this one", got)
	}

	// One good cycle clears the run: the threshold is about a *continuous*
	// outage, and a loop that recovered is not one.
	h.balance.err = nil
	if cycle := h.cycle(); cycle.Err != nil {
		t.Fatalf("the recovering cycle failed: %v", cycle.Err)
	}
	if got := h.driver.Health(); got.Consecutive != 0 || got.LastError != nil || !got.Since.IsZero() {
		t.Fatalf("a recovered driver reports %+v", got)
	}
}

// TestTheDriverRefusesAnUnverifiedGate is the execution predicate. The loop
// writes to the ledger and starts protecting positions with real orders; an
// engine whose master switch is off must do neither unattended.
func TestTheDriverRefusesAnUnverifiedGate(t *testing.T) {
	dir := isolate(t)
	writeCredentials(t, dir, "test-api-key-000000", "test-secret")
	srv, _ := interlockServer(t, "123-45")
	writeGateConfig(t, dir, config.AutomationGate{})

	eng, err := openGateEngine(t, dir, srv, nil)
	if err != nil {
		t.Fatalf("start: %v", err)
	}
	if _, err := eng.ReconcileDriver(engine.ReconcileDriverOptions{
		Collector: &reconcile.Collector{},
	}); !errors.Is(err, engine.ErrReconcileDriverUnavailable) {
		t.Fatalf("ReconcileDriver on a gate-off engine: %v, want ErrReconcileDriverUnavailable", err)
	}
}

// TestRunStopsWithItsContext proves the loop leaves no goroutine behind.
func TestRunStopsWithItsContext(t *testing.T) {
	h := newDriverHarness(t, nil)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- h.driver.Run(ctx) }()

	waitForSleeper(h.t, h.clk)
	cancel()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Errorf("Run returned %v, want context.Canceled", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Run did not return when its context was cancelled")
	}
}

// TestAnExternalIncreaseAfterAdoptionIsReported is design A8: the frozen t0
// stays frozen and the operator is told, because recomputing the denominator
// would rewrite every R already reported for the position.
func TestAnExternalIncreaseAfterAdoptionIsReported(t *testing.T) {
	h := newDriverHarness(t, nil)
	h.holds("005930", "10", "55000", 70000)
	if cycle := h.cycle(); cycle.Adopted != 1 {
		t.Fatalf("adopted = %d (%v)", cycle.Adopted, cycle.Err)
	}
	before, err := h.journal.ExitState(context.Background(), h.position("005930").ID)
	if err != nil {
		t.Fatal(err)
	}

	// The owner buys five more by hand.
	h.holdings.items[0].Quantity = "15"
	if cycle := h.cycle(); cycle.Err != nil {
		t.Fatalf("cycle: %v", cycle.Err)
	}

	after, err := h.journal.ExitState(context.Background(), h.position("005930").ID)
	if err != nil {
		t.Fatal(err)
	}
	if after.EntryPrice != before.EntryPrice || after.InitialRisk != before.InitialRisk {
		t.Errorf("t0 moved from %s/%s to %s/%s; the frozen denominator must not be recomputed",
			before.EntryPrice, before.InitialRisk, after.EntryPrice, after.InitialRisk)
	}
	found := false
	for _, e := range h.alerts.events {
		if e.Type == obs.EventExitPositionUnmanaged && strings.Contains(e.Title, "grew") {
			found = true
			if e.Fields["adopted_quantity"] != "10" {
				t.Errorf("the alert must name the quantity the adoption froze: %+v", e.Fields)
			}
		}
	}
	if !found {
		t.Errorf("no alert about the external increase; events = %v", eventTypes(h.alerts.events))
	}
}

// slowPrices is a price read that takes longer than the staleness bound.
type slowPrices struct {
	inner *fakePrices
	delay time.Duration
	clock func() *clock.Fake
}

func (s *slowPrices) Prices(ctx context.Context, symbols []string) ([]domain.Quote, error) {
	out, err := s.inner.Prices(ctx, symbols)
	s.clock().Advance(s.delay)
	return out, err
}

func eventTypes(events []obs.Event) []string {
	out := make([]string, 0, len(events))
	for _, e := range events {
		out = append(out, string(e.Type))
	}
	return out
}

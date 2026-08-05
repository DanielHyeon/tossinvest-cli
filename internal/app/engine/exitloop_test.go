package engine_test

// exitloop_test.go covers the exit observation loop (tasks 7.4 and 7.5).
//
// Every test drives ObserveOnce rather than Run, because what is under test is
// the cycle and not the sleeping between cycles — with one exception, which
// exists precisely to prove the loop stops when its context does and leaves no
// goroutine behind.
//
// The Guardian is the real one. A fake issuer would let a proposal reach the
// gateway without the reduce-only check the spec routes it through, and the one
// thing an exit can get wrong is selling more than the account holds.

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/app/engine"
	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/costs"
	"github.com/JungHoonGhae/tossinvest-cli/internal/domain"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/exitpolicy"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
	"github.com/JungHoonGhae/tossinvest-cli/internal/risk"
	"github.com/JungHoonGhae/tossinvest-cli/internal/riskcalc"
)

var exitNow = time.Date(2026, 3, 30, 1, 0, 0, 0, time.UTC)

const exitAccount = "acct-exit"

// --- fakes --------------------------------------------------------------------

type fakePrices struct {
	last       map[string]float64
	currencies map[string]string
	err        error
	calls      int
	asked      [][]string
}

func (f *fakePrices) Prices(_ context.Context, symbols []string) ([]domain.Quote, error) {
	f.calls++
	f.asked = append(f.asked, append([]string(nil), symbols...))
	if f.err != nil {
		return nil, f.err
	}
	out := make([]domain.Quote, 0, len(symbols))
	for _, s := range symbols {
		if v, ok := f.last[s]; ok {
			currency := "KRW"
			if configured, ok := f.currencies[s]; ok {
				currency = configured
			}
			out = append(out, domain.Quote{Symbol: s, Last: v, Currency: currency})
		}
	}
	return out, nil
}

// fakeSubmitter stands in for the gateway's transport, not for the gateway's
// bookkeeping: an accepted place is *recorded in the journal* through `record`,
// so the order it created is a working order the loop can find and cancel and a
// fill can later be applied to. A submitter that only counted calls would make
// every test about the working-order path vacuous.
type fakeSubmitter struct {
	places  []execgw.PlaceRequest
	cancels []execgw.CancelRequest
	// placeOutcome, when set, replaces the default CONFIRMED answer.
	placeOutcome *execgw.Outcome
	cancelFails  bool
	nextOrder    int
	record       func(execgw.PlaceRequest) string
	settle       func(orderID string)
}

func (f *fakeSubmitter) Place(_ context.Context, req execgw.PlaceRequest) (execgw.Outcome, error) {
	f.places = append(f.places, req)
	if f.placeOutcome != nil {
		return *f.placeOutcome, nil
	}
	f.nextOrder++
	orderID := fmt.Sprintf("O-exit-%d", f.nextOrder)
	if f.record != nil {
		orderID = f.record(req)
	}
	return execgw.Outcome{
		IntentID: req.IntentID, State: journal.StateConfirmed, BrokerOrderID: orderID,
	}, nil
}

func (f *fakeSubmitter) Cancel(_ context.Context, req execgw.CancelRequest) (execgw.Outcome, error) {
	f.cancels = append(f.cancels, req)
	if f.cancelFails {
		return execgw.Outcome{Reason: execgw.ReasonSymbolInFlight, Detail: "busy"},
			errors.New("cancel refused")
	}
	if f.settle != nil {
		f.settle(req.Intent.OrderID)
	}
	return execgw.Outcome{State: journal.StateConfirmed}, nil
}

type fakeAlerts struct{ events []obs.Event }

func (f *fakeAlerts) Notify(_ context.Context, e obs.Event) error {
	f.events = append(f.events, e)
	return nil
}

func (f *fakeAlerts) count(t obs.EventType) int {
	n := 0
	for _, e := range f.events {
		if e.Type == t {
			n++
		}
	}
	return n
}

func (f *fakeAlerts) first(t obs.EventType) (obs.Event, bool) {
	for _, e := range f.events {
		if e.Type == t {
			return e, true
		}
	}
	return obs.Event{}, false
}

type fakeFloor struct {
	quantity string
	bound    string
	applies  bool
	err      error
}

func (f *fakeFloor) ConfirmedFloor(context.Context, string, string) (riskcalc.ConfirmedFloor, bool, error) {
	if f.err != nil {
		return riskcalc.ConfirmedFloor{}, true, f.err
	}
	return riskcalc.ConfirmedFloor{Quantity: f.quantity, Bound: f.bound}, f.applies, nil
}

type fakeSLO struct{ behind bool }

func (f *fakeSLO) FillDetectionBehind() bool { return f.behind }

// --- harness ------------------------------------------------------------------

type exitHarness struct {
	t *testing.T
	// dbPath is the journal file, so a test that has to inspect or age a row the
	// journal API deliberately does not expose can open it directly.
	dbPath   string
	journal  *journal.Journal
	clk      *clock.Fake
	gate     *execgw.EntryGate
	prices   *fakePrices
	submit   *fakeSubmitter
	alerts   *fakeAlerts
	floor    *fakeFloor
	slo      *fakeSLO
	observer *engine.ExitObserver
	ids      int
}

// newExitHarness assembles the loop over a real journal, a real Guardian and a
// real entry gate, with only the broker surfaces faked.
func newExitHarness(t *testing.T, mutate func(*engine.ExitObserverOptions)) *exitHarness {
	t.Helper()
	clk := clock.NewFake(exitNow)
	dbPath := filepath.Join(t.TempDir(), "journal.db")
	j, err := journal.Open(context.Background(), journal.Options{
		Path:     dbPath,
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

	// A gate that only knows about the price query, so "is the entry blocked"
	// answers the one question these tests ask.
	gate := execgw.NewEntryGate(clk, map[execgw.RequiredQuery]time.Duration{
		execgw.QueryPrice: 15 * time.Second,
	})
	guardian, err := execgw.NewRiskGuardian(execgw.RiskGuardianOptions{
		Journal: j, Clock: clk, AccountRef: exitAccount,
		Policy: exitPolicy(), Costs: costs.DefaultModel(), PolicyVersion: "test/v1",
	})
	if err != nil {
		t.Fatalf("NewRiskGuardian: %v", err)
	}

	h := &exitHarness{
		t: t, dbPath: dbPath, journal: j, clk: clk, gate: gate,
		prices: &fakePrices{last: map[string]float64{}},
		submit: &fakeSubmitter{},
		alerts: &fakeAlerts{},
		floor:  &fakeFloor{},
		slo:    &fakeSLO{},
	}
	opts := engine.ExitObserverOptions{
		Journal: j,
		Prices:  h.prices,
		Retrier: &execgw.Retrier{
			Clock: clk, Gate: gate,
			Policy: execgw.RetryPolicy{MaxAttempts: 1, Budget: time.Second},
		},
		Issuer:     guardian,
		Submit:     h.submit,
		Alerts:     h.alerts,
		Costs:      costs.DefaultModel(),
		Floor:      h.floor,
		SLO:        h.slo,
		Escalate:   j,
		AccountRef: exitAccount,
		Clock:      clk,
		NewID:      func() string { h.ids++; return fmt.Sprintf("exit-intent-%d", h.ids) },
	}
	if mutate != nil {
		mutate(&opts)
	}
	observer, err := engine.NewExitObserver(opts)
	if err != nil {
		t.Fatalf("NewExitObserver: %v", err)
	}
	h.observer = observer
	h.submit.record = h.recordSubmittedSell
	h.submit.settle = h.settleCancelled
	return h
}

// recordSubmittedSell journals the order an accepted place created, under the
// intent id the proposal armed.
func (h *exitHarness) recordSubmittedSell(req execgw.PlaceRequest) string {
	h.t.Helper()
	ctx := context.Background()
	h.ids++
	orderID := fmt.Sprintf("O-sell-%d", h.ids)
	attempt, err := h.journal.Prepare(ctx, journal.PrepareRequest{
		Intent: journal.Intent{
			ID: req.IntentID, Market: "kr", TradingDay: "2026-03-30", AccountRef: exitAccount,
			Symbol: req.Intent.Symbol, Side: "SELL", OrderType: "LIMIT", TimeInForce: "DAY",
			Quantity: decimalText(req.Intent.Quantity), Price: decimalText(req.Intent.Price),
			Currency: "KRW", Source: "engine/test", Fingerprint: "fp-" + req.IntentID,
		},
		Kind: journal.KindPlace, AttemptID: "a-" + orderID,
		AccountRef: exitAccount, DecisionID: req.Decision.ID,
		SafetyClass:   journal.SafetyClassRiskReducing,
		ClientOrderID: journal.DeriveClientOrderID(req.Decision.ID, 0),
	})
	if err != nil {
		h.t.Fatalf("recording the submitted sell: %v", err)
	}
	if err := attempt.MarkDispatchStarted(ctx); err != nil {
		h.t.Fatalf("MarkDispatchStarted: %v", err)
	}
	if err := attempt.MarkAcked(ctx, orderID); err != nil {
		h.t.Fatalf("MarkAcked: %v", err)
	}
	// CONFIRMED is what makes the order a *working* one as far as the journal is
	// concerned, and the gateway settles it there on a broker acknowledgement.
	if err := attempt.Settle(ctx, journal.StateConfirmed, "broker_accepted", ""); err != nil {
		h.t.Fatalf("Settle: %v", err)
	}
	return orderID
}

// settleCancelled records the terminal snapshot a cancelled order produces, so
// the order stops being a working one.
func (h *exitHarness) settleCancelled(orderID string) {
	h.t.Helper()
	if _, err := h.journal.RecordFill(context.Background(), journal.FillObservation{
		OrderID: orderID, Symbol: "005930", Market: "kr", State: "CLOSED_CANCELED",
		Terminal: true, Quantity: "0", FilledQuantity: "0",
		ObservedAt: h.clk.Now().Format(time.RFC3339),
	}); err != nil {
		h.t.Fatalf("settling the cancelled order: %v", err)
	}
}

func decimalText(v float64) string { return strconv.FormatFloat(v, 'f', -1, 64) }

// exitPolicy is the default limit set. The allowlist is an AccountState field
// and an entry control, so a reduction never sees it — which is the point: a
// symbol removed from the allowlist must still be closeable.
func exitPolicy() risk.Policy { return risk.DefaultPolicy() }

// entry places a filled buy and returns the projected position.
//
// It goes through the journal's public API rather than inserting rows, so the
// position under test is one the projection actually produced: the entry
// decision, its intent and attempt, the acknowledgement and the fill.
func (h *exitHarness) entry(symbol, quantity, limit, stop, avg string) journal.Position {
	h.t.Helper()
	ctx := context.Background()

	limits, err := execgw.EncodeLimits(execgw.Limits{
		MaxQuantity: execgw.Bound(1000), MaxNotional: execgw.Bound(1e9),
		MaxTotalExposure: execgw.Bound(1e9), MaxDailyLossAmount: execgw.Bound(1e6),
		MaxDailyLossRatio: execgw.Bound(0.02), Currency: "KRW",
	})
	if err != nil {
		h.t.Fatalf("EncodeLimits: %v", err)
	}
	decisionID := "d-entry-" + symbol
	if _, err := h.journal.RecordDecision(ctx, journal.DecisionRequest{
		ID: decisionID, AccountRef: exitAccount, SafetyClass: journal.SafetyClassExposureRaising,
		Kind: journal.KindPlace,
		Preimage: journal.RiskIntent{
			AccountRef: exitAccount, Market: "kr", Symbol: symbol, Side: "BUY",
			Quantity: quantity, EntryPrice: limit, StopPrice: stop,
			TargetPrice: "999999", PolicyVersion: "test/v1",
		},
		LimitsJSON: limits, Nonce: "nonce-" + decisionID,
		IssuedAt: h.clk.Now(), ExpiresAt: h.clk.Now().Add(time.Hour),
	}); err != nil {
		h.t.Fatalf("RecordDecision: %v", err)
	}

	orderID := h.place(decisionID, journal.SafetyClassExposureRaising, symbol, "BUY", quantity, limit)
	if _, err := h.journal.RecordFill(ctx, journal.FillObservation{
		OrderID: orderID, Symbol: symbol, Market: "kr", State: "CLOSED_FILLED", Terminal: true,
		Quantity: quantity, FilledQuantity: quantity, AveragePrice: avg,
		ObservedAt: h.clk.Now().Format(time.RFC3339),
	}); err != nil {
		h.t.Fatalf("RecordFill(entry): %v", err)
	}
	p, err := h.journal.CurrentPosition(ctx, exitAccount, "kr", symbol)
	if err != nil {
		h.t.Fatalf("CurrentPosition: %v", err)
	}
	return p
}

// place records an acknowledged order and returns its broker order id.
func (h *exitHarness) place(decisionID, class, symbol, side, quantity, price string) string {
	h.t.Helper()
	ctx := context.Background()
	h.ids++
	intentID := fmt.Sprintf("i-%d", h.ids)
	attemptID := fmt.Sprintf("a-%d", h.ids)
	orderID := fmt.Sprintf("O-%d", h.ids)

	attempt, err := h.journal.Prepare(ctx, journal.PrepareRequest{
		Intent: journal.Intent{
			ID: intentID, Market: "kr", TradingDay: "2026-03-30", AccountRef: exitAccount,
			Symbol: symbol, Side: side, OrderType: "LIMIT", TimeInForce: "DAY",
			Quantity: quantity, Price: price, Currency: "KRW", Source: "engine/test",
			Fingerprint: "fp-" + intentID,
		},
		Kind: journal.KindPlace, AttemptID: attemptID,
		AccountRef: exitAccount, DecisionID: decisionID, SafetyClass: class,
		ClientOrderID: journal.DeriveClientOrderID(decisionID, 0),
	})
	if err != nil {
		h.t.Fatalf("Prepare: %v", err)
	}
	if err := attempt.MarkDispatchStarted(ctx); err != nil {
		h.t.Fatalf("MarkDispatchStarted: %v", err)
	}
	if err := attempt.MarkAcked(ctx, orderID); err != nil {
		h.t.Fatalf("MarkAcked: %v", err)
	}
	if err := attempt.Settle(ctx, journal.StateConfirmed, "broker_accepted", ""); err != nil {
		h.t.Fatalf("Settle: %v", err)
	}
	return orderID
}

// workingEntry leaves a live buy on the book without filling it.
func (h *exitHarness) workingEntry(symbol, quantity, limit string) string {
	h.t.Helper()
	limits, err := execgw.EncodeLimits(execgw.Limits{
		MaxQuantity: execgw.Bound(1000), MaxNotional: execgw.Bound(1e9),
		MaxTotalExposure: execgw.Bound(1e9), MaxDailyLossAmount: execgw.Bound(1e6),
		MaxDailyLossRatio: execgw.Bound(0.02), Currency: "KRW",
	})
	if err != nil {
		h.t.Fatalf("EncodeLimits: %v", err)
	}
	h.ids++
	decisionID := fmt.Sprintf("d-working-%d", h.ids)
	if _, err := h.journal.RecordDecision(context.Background(), journal.DecisionRequest{
		ID: decisionID, AccountRef: exitAccount, SafetyClass: journal.SafetyClassExposureRaising,
		Kind: journal.KindPlace,
		Preimage: journal.RiskIntent{
			AccountRef: exitAccount, Market: "kr", Symbol: symbol, Side: "BUY",
			Quantity: quantity, EntryPrice: limit, StopPrice: "1", TargetPrice: "999999",
			PolicyVersion: "test/v1",
		},
		LimitsJSON: limits, Nonce: "nonce-" + decisionID,
		IssuedAt: h.clk.Now(), ExpiresAt: h.clk.Now().Add(time.Hour),
	}); err != nil {
		h.t.Fatalf("RecordDecision(working): %v", err)
	}
	return h.place(decisionID, journal.SafetyClassExposureRaising, symbol, "BUY", quantity, limit)
}

func (h *exitHarness) quote(symbol string, last float64) {
	h.prices.last[symbol] = last
}

func (h *exitHarness) state(positionID string) journal.ExitState {
	h.t.Helper()
	s, err := h.journal.ExitState(context.Background(), positionID)
	if err != nil {
		h.t.Fatalf("ExitState: %v", err)
	}
	return s
}

func (h *exitHarness) observe() engine.ExitCycle {
	h.t.Helper()
	return h.observer.ObserveOnce(context.Background())
}

func (h *exitHarness) mode() string {
	h.t.Helper()
	rec, err := h.journal.CurrentOperatingMode(context.Background(), exitAccount)
	if err != nil {
		h.t.Fatalf("CurrentOperatingMode: %v", err)
	}
	return rec.Mode
}

// --- opening the state ---------------------------------------------------------

// TestTheLoopOpensTheExitStateOfANewlyHeldPosition is D5's first correction
// arriving through the loop: a position is protected from the instant it is
// held, at the stop its own entry decision named.
func TestTheLoopOpensTheExitStateOfANewlyHeldPosition(t *testing.T) {
	h := newExitHarness(t, nil)
	p := h.entry("005930", "10", "70000", "68000", "70000")
	h.quote("005930", 70100)

	cycle := h.observe()
	if cycle.Err != nil {
		t.Fatalf("cycle error: %v", cycle.Err)
	}
	if cycle.Opened != 1 {
		t.Fatalf("opened = %d, want 1", cycle.Opened)
	}
	state := h.state(p.ID)
	if state.Baseline != "68000" {
		t.Errorf("baseline = %s, want the entry decision's stop", state.Baseline)
	}
	if state.InitialRisk != "2000" {
		t.Errorf("initial risk = %s, want the frozen entry − stop", state.InitialRisk)
	}
	if state.PolicyKind != journal.ExitPolicyRatchet {
		t.Errorf("policy = %s, want the RATCHET default", state.PolicyKind)
	}
	// The second cycle must not open it again.
	if again := h.observe(); again.Opened != 0 {
		t.Errorf("opened = %d on the second cycle, want 0", again.Opened)
	}
}

// TestAPositionWithNoEntryDecisionIsSkippedAndAlertedOnce is exit-policy's
// second scenario. The alert is once because the loop runs every five seconds
// and an operator reading the same line 720 times an hour reads none of them.
func TestAPositionWithNoEntryDecisionIsSkippedAndAlertedOnce(t *testing.T) {
	h := newExitHarness(t, nil)
	ctx := context.Background()

	watermark, err := h.journal.FillWatermark(ctx, "000660")
	if err != nil {
		t.Fatalf("FillWatermark: %v", err)
	}
	if _, err := h.journal.ApplyPositionAdjustment(ctx, journal.AdjustmentRequest{
		AccountRef: exitAccount, Market: "kr", Symbol: "000660",
		Kind: "EXTERNAL", ExpectedPrevQuantity: "0", ExpectedFillWatermark: watermark,
		NewQuantity: "7", NewAvgPrice: "120000",
		BrokerAsOf: h.clk.Now().Format(time.RFC3339), Evidence: "the account holds it and we did not buy it",
	}); err != nil {
		t.Fatalf("ApplyPositionAdjustment: %v", err)
	}

	first := h.observe()
	if first.Unmanaged != 1 {
		t.Fatalf("unmanaged = %d, want 1", first.Unmanaged)
	}
	if first.Judged != 0 || h.prices.calls != 0 {
		t.Errorf("an unmanaged position must not be judged and must cost no price read (calls=%d)",
			h.prices.calls)
	}
	h.observe()
	h.observe()
	if got := h.alerts.count(obs.EventExitPositionUnmanaged); got != 1 {
		t.Errorf("unmanaged alerts = %d, want exactly one", got)
	}
	if e, _ := h.alerts.first(obs.EventExitPositionUnmanaged); obs.SeverityOf(e.Type) != obs.SeverityNormal {
		t.Error("somebody trading their own account by hand is not a critical condition")
	}
}

// --- the observation ------------------------------------------------------------

// TestASuccessfulObservationStampsThePriceFreshness is the wiring the task calls
// out by name: without the stamp a gate-ON engine refuses every entry as
// QUERY_STALE, because nothing else in the build ever reads a price.
func TestASuccessfulObservationStampsThePriceFreshness(t *testing.T) {
	h := newExitHarness(t, nil)
	h.entry("005930", "10", "70000", "68000", "70000")
	h.quote("005930", 70100)

	if h.gate.CheckEntry() == nil {
		t.Fatal("before any observation the price query has never succeeded and entries must be blocked")
	}
	if cycle := h.observe(); cycle.Err != nil {
		t.Fatalf("cycle error: %v", cycle.Err)
	}
	if rejected := h.gate.CheckEntry(); rejected != nil {
		t.Fatalf("after a successful observation the entry gate must be open, got %v", rejected)
	}

	// And it goes stale again on its own, which is the first rung of the ladder.
	h.clk.Advance(16 * time.Second)
	if h.gate.CheckEntry() == nil {
		t.Error("a price older than the staleness threshold must block entries again")
	}
}

// TestOneCycleIsOnePriceReadForEveryHeldSymbol is the §0.4 budget claim as a
// test: the request count is a function of the interval, not of the portfolio.
func TestOneCycleIsOnePriceReadForEveryHeldSymbol(t *testing.T) {
	h2 := newExitHarness(t, nil)
	h2.entry("005930", "10", "70000", "68000", "70000")
	h2.entry("000660", "4", "120000", "118000", "120000")
	h2.quote("005930", 70100)
	h2.quote("000660", 120100)

	h2.observe()
	if h2.prices.calls != 1 {
		t.Fatalf("price reads = %d for two symbols, want one fan-out call", h2.prices.calls)
	}
	if got := strings.Join(h2.prices.asked[0], ","); got != "000660,005930" {
		t.Errorf("asked for %q, want both held symbols in one call", got)
	}
}

// TestAFailedObservationHoldsTheJudgement is exit-policy's fail-safe: the
// baseline and the watermark are kept and nothing is judged.
func TestAFailedObservationHoldsTheJudgement(t *testing.T) {
	h := newExitHarness(t, nil)
	p := h.entry("005930", "10", "70000", "68000", "70000")
	h.quote("005930", 84100) // +7R: enough to move everything, if it were read
	h.observe()
	before := h.state(p.ID)

	submitted := len(h.submit.places)
	h.prices.err = errors.New("the market data endpoint is down")
	cycle := h.observe()
	if cycle.Err == nil {
		t.Fatal("a failed read must be reported on the cycle")
	}
	after := h.state(p.ID)
	if after.Baseline != before.Baseline || after.HighWater != before.HighWater {
		t.Errorf("state moved on a failed observation: %+v → %+v", before, after)
	}
	if len(h.submit.places) != submitted {
		t.Error("a failed observation must propose nothing")
	}
}

// TestAQuoteWithNoLastTradeIsNotAnObservation guards the one shape that would be
// catastrophic to accept: a zero price reads as a total collapse and would
// liquidate every position that received it.
func TestAQuoteWithNoLastTradeIsNotAnObservation(t *testing.T) {
	h := newExitHarness(t, nil)
	p := h.entry("005930", "10", "70000", "68000", "70000")
	h.quote("005930", 70100)
	h.observe()

	h.quote("005930", 0)
	cycle := h.observe()
	if cycle.Err == nil {
		t.Fatal("a quote with no last trade must not count as an observation")
	}
	if len(h.submit.places) != 0 {
		t.Fatal("a zero price must never produce a liquidation")
	}
	if h.state(p.ID).Baseline != "68000" {
		t.Error("the baseline must be held")
	}
}

// --- the outage ladder ----------------------------------------------------------

// TestASustainedOutageBlocksEntriesAndAlertsOnce is exit-policy's "관측 장기
// 두절" scenario, and the fourth producer of an automatic tightening.
func TestASustainedOutageBlocksEntriesAndAlertsOnce(t *testing.T) {
	h := newExitHarness(t, nil)
	p := h.entry("005930", "10", "70000", "68000", "70000")
	h.quote("005930", 70100)
	h.observe()

	h.prices.err = errors.New("down")
	h.clk.Advance(30 * time.Second)
	if cycle := h.observe(); cycle.Escalated {
		t.Fatal("thirty seconds is inside the threshold; the ladder's second rung must not fire yet")
	}
	if h.mode() != journal.ModeNormal {
		t.Fatalf("mode = %s at 30s, want NORMAL", h.mode())
	}

	h.clk.Advance(31 * time.Second)
	cycle := h.observe()
	if !cycle.Escalated {
		t.Fatal("past the threshold the loop must tighten the operating mode")
	}
	if h.mode() != journal.ModeEntryBlocked {
		t.Fatalf("mode = %s, want ENTRY_BLOCKED", h.mode())
	}
	alert, ok := h.alerts.first(obs.EventExitObservationOutage)
	if !ok {
		t.Fatal("the outage must be alerted")
	}
	if obs.SeverityOf(alert.Type) != obs.SeverityCritical {
		t.Error("with no broker-resident stop an unobserved position is unprotected; the alert is critical")
	}

	h.clk.Advance(10 * time.Second)
	h.observe()
	if got := h.alerts.count(obs.EventExitObservationOutage); got != 1 {
		t.Errorf("outage alerts = %d, want one per outage", got)
	}

	// Recovery re-arms the alert for the *next* outage rather than leaving it spent.
	h.prices.err = nil
	h.observe()
	h.prices.err = errors.New("down again")
	h.clk.Advance(61 * time.Second)
	h.observe()
	if got := h.alerts.count(obs.EventExitObservationOutage); got != 2 {
		t.Errorf("outage alerts = %d after a second outage, want 2", got)
	}
	_ = p
}

// TestAnAccountHoldingNothingIsNotInAnOutage: escalating an idle account would
// block the entries that are the only thing that could clear the condition.
func TestAnAccountHoldingNothingIsNotInAnOutage(t *testing.T) {
	h := newExitHarness(t, nil)
	h.clk.Advance(10 * time.Minute)

	cycle := h.observe()
	if cycle.Escalated || h.mode() != journal.ModeNormal {
		t.Fatalf("an account with no positions must not be tightened (mode=%s)", h.mode())
	}
	if h.prices.calls != 0 {
		t.Errorf("price reads = %d with nothing held, want 0", h.prices.calls)
	}
}

// TestTheCycleYieldsToFillDetection is the SLO deference: fill detection first.
func TestTheCycleYieldsToFillDetection(t *testing.T) {
	h := newExitHarness(t, nil)
	h.entry("005930", "10", "70000", "68000", "70000")
	h.quote("005930", 60000) // a breach, which is exactly what must still wait

	h.slo.behind = true
	cycle := h.observe()
	if !cycle.Deferred {
		t.Fatal("the cycle must report that it yielded")
	}
	if h.prices.calls != 0 {
		t.Errorf("price reads = %d while deferring, want 0", h.prices.calls)
	}

	// Deference is not free: the outage clock keeps running, so a detector that
	// never catches up escalates exactly as an outage does.
	h.clk.Advance(61 * time.Second)
	if got := h.observe(); !got.Escalated {
		t.Error("sustained deference must reach the same ladder as sustained failure")
	}
}

// --- judgement and proposal ------------------------------------------------------

// TestABaselineBreachProposesTheWholePosition is the t0 scenario: below the
// entry stop, before +0.4R, the whole position is liquidated.
func TestABaselineBreachProposesTheWholePosition(t *testing.T) {
	h := newExitHarness(t, nil)
	p := h.entry("005930", "10", "70000", "68000", "70000")
	h.quote("005930", 70100)
	h.observe()

	h.quote("005930", 67900)
	cycle := h.observe()
	if cycle.Err != nil {
		t.Fatalf("cycle error: %v", cycle.Err)
	}
	if cycle.Proposed != 1 {
		t.Fatalf("proposed = %d, want the liquidation", cycle.Proposed)
	}
	if len(h.submit.places) != 1 {
		t.Fatalf("places = %d, want one sell", len(h.submit.places))
	}
	place := h.submit.places[0]
	if place.Intent.Side != "sell" || place.Intent.Quantity != 10 {
		t.Errorf("submitted %+v, want the whole position as a sell", place.Intent)
	}
	if place.Intent.OrderType != "limit" {
		t.Errorf("order type = %s; automated orders are LIMIT only", place.Intent.OrderType)
	}
	if place.IntentID == "" {
		t.Fatal("the place must carry the pre-minted intent id")
	}
	state := h.state(p.ID)
	if state.PendingAction != string(exitpolicy.ActionBaselineBreach) {
		t.Errorf("pending action = %q, want the armed breach", state.PendingAction)
	}
	if state.PendingIntentID != place.IntentID {
		t.Errorf("pending intent = %q but the order went under %q; a fill would resolve nothing",
			state.PendingIntentID, place.IntentID)
	}

	events, err := h.journal.ExitEvents(context.Background(), p.ID)
	if err != nil {
		t.Fatalf("ExitEvents: %v", err)
	}
	var proposed bool
	for _, e := range events {
		if e.Action == string(exitpolicy.ActionBaselineBreach) && e.ProposedIntentID == place.IntentID {
			proposed = true
		}
	}
	if !proposed {
		t.Errorf("the judgement history carries no proposal row: %+v", events)
	}
}

// TestAnUnchangedJudgementWritesNoHistoryRow is the de-duplication issues.md
// hands this task: exit_events is append-only and the loop runs every 5 s.
func TestAnUnchangedJudgementWritesNoHistoryRow(t *testing.T) {
	h := newExitHarness(t, nil)
	p := h.entry("005930", "10", "70000", "68000", "70000")
	h.quote("005930", 70100)
	h.observe()
	before, err := h.journal.ExitEvents(context.Background(), p.ID)
	if err != nil {
		t.Fatalf("ExitEvents: %v", err)
	}

	// The same price, three more times. The watermark cannot move and no level
	// can change, so nothing is worth recording.
	h.observe()
	h.observe()
	h.observe()
	after, err := h.journal.ExitEvents(context.Background(), p.ID)
	if err != nil {
		t.Fatalf("ExitEvents: %v", err)
	}
	if len(after) != len(before) {
		t.Errorf("history grew from %d to %d rows on unchanged observations", len(before), len(after))
	}
}

// TestTheRatchetRaisesAndThenTakesFortyPercent walks the trigger table through
// the loop: +0.4R raises the baseline, +1.0R proposes the partial.
func TestTheRatchetRaisesAndThenTakesFortyPercent(t *testing.T) {
	h := newExitHarness(t, nil)
	p := h.entry("005930", "10", "70000", "68000", "70000")

	// +0.5R → HALF_RISK, baseline to −0.5R = 69000.
	h.quote("005930", 71000)
	h.observe()
	state := h.state(p.ID)
	if state.RatchetLevel != journal.RatchetHalfRisk {
		t.Fatalf("level = %s, want HALF_RISK", state.RatchetLevel)
	}
	if state.Baseline != "69000" {
		t.Errorf("baseline = %s, want −0.5R", state.Baseline)
	}
	if len(h.submit.places) != 0 {
		t.Error("a baseline raise proposes no order")
	}

	// +1.0R → the 40 % partial, sized off the remaining quantity.
	h.quote("005930", 72000)
	cycle := h.observe()
	if cycle.Proposed != 1 {
		t.Fatalf("proposed = %d at +1.0R, want the partial", cycle.Proposed)
	}
	if len(h.submit.places) != 1 || h.submit.places[0].Intent.Quantity != 4 {
		t.Fatalf("submitted %+v, want 40%% of 10", h.submit.places[0].Intent)
	}
	if got := h.state(p.ID).PendingAction; got != string(exitpolicy.ActionRatchetPartial) {
		t.Errorf("pending action = %q, want the partial", got)
	}
}

// TestAnUnresolvedProposalSuppressesTheNextOne is the pending lifecycle seen
// from the loop: exit-policy's "미체결 중 재관측".
func TestAnUnresolvedProposalSuppressesTheNextOne(t *testing.T) {
	h := newExitHarness(t, nil)
	h.entry("005930", "10", "70000", "68000", "70000")
	h.quote("005930", 72000)
	h.observe()
	if len(h.submit.places) != 1 {
		t.Fatalf("places = %d, want the first proposal", len(h.submit.places))
	}

	h.quote("005930", 72500)
	h.observe()
	if len(h.submit.places) != 1 {
		t.Errorf("places = %d; a second proposal must be suppressed while the first is outstanding",
			len(h.submit.places))
	}
}

// TestABreachDisplacesAnOutstandingTakeProfit is §0-3 through the loop: a
// pending partial must not delay a stop. The take-profit's order comes off the
// book first, because two working sells on one position can oversell it.
func TestABreachDisplacesAnOutstandingTakeProfit(t *testing.T) {
	h := newExitHarness(t, nil)
	p := h.entry("005930", "10", "70000", "68000", "70000")
	h.quote("005930", 72000)
	h.observe() // the 40 % partial is armed and submitted

	h.quote("005930", 68500) // below the raised baseline of 69000
	cycle := h.observe()
	if cycle.Err != nil {
		t.Fatalf("cycle error: %v", cycle.Err)
	}
	if len(h.submit.cancels) == 0 {
		t.Fatal("the outstanding take-profit must be cancelled before the liquidation is submitted")
	}
	if len(h.submit.places) != 2 {
		t.Fatalf("places = %d, want the partial and then the liquidation", len(h.submit.places))
	}
	state := h.state(p.ID)
	if state.PendingAction != string(exitpolicy.ActionBaselineBreach) {
		t.Errorf("pending action = %q, want the breach to have displaced the partial", state.PendingAction)
	}
	if state.PendingIntentID != h.submit.places[1].IntentID {
		t.Error("the armed proposal must name the liquidation's intent")
	}
}

// TestAWorkingEntryIsCancelledBeforeTheLiquidation is the E31 succession: a sell
// submitted over a working buy can be answered by an entry fill during the
// close, which the projection refuses and turns into a RECONCILE.
func TestAWorkingEntryIsCancelledBeforeTheLiquidation(t *testing.T) {
	h := newExitHarness(t, nil)
	h.entry("005930", "10", "70000", "68000", "70000")
	orderID := h.workingEntry("005930", "5", "69500")
	h.quote("005930", 67900)

	h.observe()
	if len(h.submit.cancels) != 1 {
		t.Fatalf("cancels = %d, want the working entry", len(h.submit.cancels))
	}
	if h.submit.cancels[0].Intent.OrderID != orderID {
		t.Errorf("cancelled %q, want the working entry %q", h.submit.cancels[0].Intent.OrderID, orderID)
	}
	if len(h.submit.places) != 1 {
		t.Fatalf("places = %d, want the liquidation after the cancel", len(h.submit.places))
	}
}

// TestAnUncancellableEntryWithholdsTheLiquidationAndAlertsPastTheBound is the
// §0.3 delay window: the liquidation is not stacked on top of a live buy, the
// judgement still advances, and the delay becomes a critical alert once it
// passes the resolution bound.
func TestAnUncancellableEntryWithholdsTheLiquidationAndAlertsPastTheBound(t *testing.T) {
	h := newExitHarness(t, nil)
	p := h.entry("005930", "10", "70000", "68000", "70000")
	h.workingEntry("005930", "5", "69500")
	h.submit.cancelFails = true
	h.quote("005930", 67900)

	h.observe()
	if len(h.submit.places) != 0 {
		t.Fatal("no liquidation may be submitted over a buy that is still on the book")
	}
	if h.state(p.ID).Pending() {
		t.Error("nothing was submitted, so nothing may be armed")
	}
	if h.state(p.ID).HighWater == "" {
		t.Error("the judgement itself must still have been recorded")
	}
	if got := h.alerts.count(obs.EventExitLiquidationDelayed); got != 0 {
		t.Errorf("delay alerts = %d inside the bound, want 0", got)
	}

	h.clk.Advance(31 * time.Second)
	h.observe()
	alert, ok := h.alerts.first(obs.EventExitLiquidationDelayed)
	if !ok {
		t.Fatal("a delay past the bound is a critical alert")
	}
	if obs.SeverityOf(alert.Type) != obs.SeverityCritical {
		t.Error("the delay alert is critical: it is unprotected exposure")
	}

	// Once the cancel succeeds the liquidation goes and the alert re-arms.
	h.submit.cancelFails = false
	h.observe()
	if len(h.submit.places) != 1 {
		t.Fatalf("places = %d once the entry could be cancelled, want the liquidation", len(h.submit.places))
	}
}

// --- the confirmed floor (task 7.5) ----------------------------------------------

// TestTheConfirmedFloorCapsTheLiquidation is exit-policy's "확정 하한 캡"
// scenario: what the floor authorises is submitted, the remainder is not, and
// the cap is alerted.
func TestTheConfirmedFloorCapsTheLiquidation(t *testing.T) {
	h := newExitHarness(t, nil)
	h.entry("005930", "10", "70000", "68000", "70000")
	h.floor.applies, h.floor.quantity, h.floor.bound = true, "3", riskcalc.FloorBoundHoldings
	h.quote("005930", 67900)

	h.observe()
	if len(h.submit.places) != 1 {
		t.Fatalf("places = %d, want the capped liquidation", len(h.submit.places))
	}
	if got := h.submit.places[0].Intent.Quantity; got != 3 {
		t.Errorf("submitted %v, want the floor of 3", got)
	}
	alert, ok := h.alerts.first(obs.EventExitProposalCapped)
	if !ok {
		t.Fatal("a cap is alerted")
	}
	if alert.Fields["remainder"] != "7" {
		t.Errorf("alert remainder = %v, want 7", alert.Fields["remainder"])
	}
}

// TestAZeroFloorSubmitsNothingAndLeavesTheLevelProposable is the other half of
// the same requirement. The proposal is *released* rather than left armed,
// because an armed proposal suppresses exactly the re-proposal exit-policy asks
// for once the disagreement is resolved.
func TestAZeroFloorSubmitsNothingAndLeavesTheLevelProposable(t *testing.T) {
	h := newExitHarness(t, nil)
	p := h.entry("005930", "10", "70000", "68000", "70000")
	h.floor.applies, h.floor.quantity, h.floor.bound = true, "0", riskcalc.FloorBoundStaleSnapshot
	h.quote("005930", 67900)

	h.observe()
	if len(h.submit.places) != 0 {
		t.Fatal("a floor of zero authorises no sale")
	}
	if h.state(p.ID).Pending() {
		t.Fatal("the level must stay proposable, so the armed proposal is released")
	}

	// The disagreement clears: the same level identity is proposed again, and it
	// is re-derived from the ledger rather than from anything this process
	// remembered.
	h.floor.applies = false
	h.observe()
	if len(h.submit.places) != 1 {
		t.Fatalf("places = %d after the floor lifted, want the re-proposal", len(h.submit.places))
	}
	if got := h.submit.places[0].Intent.Quantity; got != 10 {
		t.Errorf("re-proposed %v, want the whole remainder", got)
	}
}

// TestAFloorThatCannotBeComputedSellsNothing keeps the direction honest: under
// RECONCILE a floor the engine failed to establish is zero, not absent.
func TestAFloorThatCannotBeComputedSellsNothing(t *testing.T) {
	h := newExitHarness(t, nil)
	h.entry("005930", "10", "70000", "68000", "70000")
	h.floor.err = errors.New("the holdings snapshot could not be read")
	h.quote("005930", 67900)

	h.observe()
	if len(h.submit.places) != 0 {
		t.Fatal("a floor that could not be computed must authorise nothing")
	}
}

// TestNoFloorSourceCapsNothing is the §0.3 regression: an engine with no
// reconciliation loop must not have its liquidations bounded by a source that
// does not exist.
func TestNoFloorSourceCapsNothing(t *testing.T) {
	h := newExitHarness(t, func(o *engine.ExitObserverOptions) { o.Floor = nil })
	h.entry("005930", "10", "70000", "68000", "70000")
	h.quote("005930", 67900)

	h.observe()
	if len(h.submit.places) != 1 || h.submit.places[0].Intent.Quantity != 10 {
		t.Fatalf("places = %+v, want the whole position", h.submit.places)
	}
}

// --- refusals ---------------------------------------------------------------------

// TestARefusedProposalReleasesTheLevelAndAlerts: a liquidation the gateway would
// not take is protection that did not happen, so it is critical, and the level
// has to stay proposable.
func TestARefusedProposalReleasesTheLevelAndAlerts(t *testing.T) {
	h := newExitHarness(t, nil)
	p := h.entry("005930", "10", "70000", "68000", "70000")
	h.submit.placeOutcome = &execgw.Outcome{
		State: journal.StateFailedConfirmed, Reason: execgw.ReasonBrokerRejected, Detail: "no",
	}
	h.quote("005930", 67900)

	h.observe()
	if h.state(p.ID).Pending() {
		t.Fatal("a refused proposal must be released so its level can be proposed again")
	}
	alert, ok := h.alerts.first(obs.EventExitProposalRefused)
	if !ok {
		t.Fatal("a refused proposal is alerted")
	}
	if obs.SeverityOf(alert.Type) != obs.SeverityCritical {
		t.Error("the protection the policy asked for did not happen; that is critical")
	}
}

// TestAnInDoubtSubmissionKeepsTheProposalArmed: the order may exist, and
// releasing would let the next observation stack a second sell on top of it.
func TestAnInDoubtSubmissionKeepsTheProposalArmed(t *testing.T) {
	h := newExitHarness(t, nil)
	p := h.entry("005930", "10", "70000", "68000", "70000")
	h.submit.placeOutcome = &execgw.Outcome{State: journal.StateInDoubt, Detail: "transport gave up"}
	h.quote("005930", 67900)

	h.observe()
	if !h.state(p.ID).Pending() {
		t.Fatal("an IN_DOUBT submission may have reached the broker; the proposal stays armed")
	}
	h.quote("005930", 67800)
	h.observe()
	if len(h.submit.places) != 1 {
		t.Errorf("places = %d; nothing may be stacked on an IN_DOUBT sell", len(h.submit.places))
	}
}

// --- the ladder policy gap (issues.md, task 7.2) ------------------------------------

// TestARungTableSwappedUnderALivePositionIsRefused is the storage-free half of
// the policy_id gap. A table whose rung locks above the baseline it should have
// set is not the table that rung was activated from.
func TestARungTableSwappedUnderALivePositionIsRefused(t *testing.T) {
	h := newExitHarness(t, func(o *engine.ExitObserverOptions) {
		// A table whose first rung locks at +5 %: a position standing on rung 0
		// with a baseline of 68000 (below 70000 × 1.05) cannot have come from it.
		policy := exitpolicy.DefaultLadderPolicy()
		policy.Rungs = []exitpolicy.Rung{{TargetPct: "6.0", StopPct: "5.0", PartialRatio: "1"}}
		o.Ladder = &policy
	})
	p := h.entry("005930", "10", "70000", "68000", "70000")
	ctx := context.Background()
	if _, err := h.journal.OpenExitState(ctx, journal.ExitStateSeed{
		PositionID: p.ID, PolicyKind: journal.ExitPolicyLadder,
		EntryPrice: "70000", InitialStop: "68000",
	}); err != nil {
		t.Fatalf("OpenExitState: %v", err)
	}
	if err := h.journal.RecordExitJudgement(ctx, journal.ExitJudgement{
		PositionID: p.ID, HighWater: "71000", Baseline: "68000", ActiveRung: 0,
	}); err != nil {
		t.Fatalf("RecordExitJudgement: %v", err)
	}

	h.quote("005930", 71000)
	h.observe()
	alert, ok := h.alerts.first(obs.EventExitJudgementRefused)
	if !ok {
		t.Fatal("a rung table replaced under a live position is a refusal to judge, not a quiet no-op")
	}
	if !strings.Contains(alert.Body, "rung table was replaced") {
		t.Errorf("alert body = %q, want it to name the cause", alert.Body)
	}
	if len(h.submit.places) != 0 {
		t.Error("a position the policy cannot judge must not have orders proposed for it")
	}
	// The alert is once, not once per cycle.
	h.observe()
	if got := h.alerts.count(obs.EventExitJudgementRefused); got != 1 {
		t.Errorf("refusal alerts = %d, want one", got)
	}
}

// TestARungTableShorterThanTheStoredIndexIsRefused is the other detectable
// shape: the stored index does not name a rung at all.
func TestARungTableShorterThanTheStoredIndexIsRefused(t *testing.T) {
	h := newExitHarness(t, func(o *engine.ExitObserverOptions) {
		policy := exitpolicy.DefaultLadderPolicy()
		policy.Rungs = policy.Rungs[:1]
		o.Ladder = &policy
	})
	p := h.entry("005930", "10", "70000", "68000", "70000")
	ctx := context.Background()
	if _, err := h.journal.OpenExitState(ctx, journal.ExitStateSeed{
		PositionID: p.ID, PolicyKind: journal.ExitPolicyLadder,
		EntryPrice: "70000", InitialStop: "68000",
	}); err != nil {
		t.Fatalf("OpenExitState: %v", err)
	}
	if err := h.journal.RecordExitJudgement(ctx, journal.ExitJudgement{
		PositionID: p.ID, HighWater: "74000", Baseline: "70700", ActiveRung: 2,
	}); err != nil {
		t.Fatalf("RecordExitJudgement: %v", err)
	}

	h.quote("005930", 74000)
	h.observe()
	if _, ok := h.alerts.first(obs.EventExitJudgementRefused); !ok {
		t.Fatal("a stored rung the table cannot name is a refusal to judge")
	}
}

// --- crash restore ------------------------------------------------------------------

// TestAnArmedProposalIsNotProposedTwiceAfterARestart is exit-policy's "제출 전
// 크래시" scenario at the loop level: a fresh observer over the same journal
// picks the armed proposal up rather than making it again.
func TestAnArmedProposalIsNotProposedTwiceAfterARestart(t *testing.T) {
	h := newExitHarness(t, nil)
	p := h.entry("005930", "10", "70000", "68000", "70000")
	h.quote("005930", 67900)
	h.observe()

	armed := h.state(p.ID)
	if !armed.Pending() {
		t.Fatal("the proposal must be armed")
	}
	attempted, err := h.journal.IntentAttempted(context.Background(), armed.PendingIntentID)
	if err != nil {
		t.Fatalf("IntentAttempted: %v", err)
	}
	if attempted {
		t.Skip("the fake submitter records no attempt; the identity check is covered in execgw")
	}

	// "Restart": a second observer over the same journal, with its own memory.
	restarted, err := engine.NewExitObserver(engine.ExitObserverOptions{
		Journal: h.journal, Prices: h.prices,
		Retrier: &execgw.Retrier{
			Clock: h.clk, Gate: h.gate,
			Policy: execgw.RetryPolicy{MaxAttempts: 1, Budget: time.Second},
		},
		Issuer: stubIssuer{}, Submit: h.submit, Alerts: h.alerts,
		Costs: costs.DefaultModel(), AccountRef: exitAccount, Clock: h.clk,
	})
	if err != nil {
		t.Fatalf("NewExitObserver: %v", err)
	}
	before := len(h.submit.places)
	restarted.ObserveOnce(context.Background())
	if len(h.submit.places) != before {
		t.Errorf("places = %d after the restart, want %d — the armed proposal must not be made again",
			len(h.submit.places), before)
	}
	if h.state(p.ID).PendingIntentID != armed.PendingIntentID {
		t.Error("the restored proposal must keep its intent id")
	}
}

type stubIssuer struct{}

func (stubIssuer) IssueReduction(context.Context, execgw.ReductionIssuance) (execgw.Issued, error) {
	return execgw.Issued{}, errors.New("the restarted observer must not need to issue anything")
}

// --- lifecycle ------------------------------------------------------------------------

// TestRunStopsWithItsContextAndLeavesNoGoroutine is the loop-lifecycle contract.
func TestRunStopsWithItsContextAndLeavesNoGoroutine(t *testing.T) {
	h := newExitHarness(t, nil)
	h.entry("005930", "10", "70000", "68000", "70000")
	h.quote("005930", 70100)

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	done := make(chan error, 1)
	go func() { done <- h.observer.Run(ctx) }()

	if !h.clk.WaitForSleepers(1, 2*time.Second) {
		t.Fatal("the loop never reached its interval sleep")
	}
	cancel()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Errorf("Run returned %v, want context.Canceled", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Run did not return")
	}
}

// --- construction -----------------------------------------------------------------------

func TestTheObserverRefusesToExistWithoutItsRetrier(t *testing.T) {
	_, err := engine.NewExitObserver(engine.ExitObserverOptions{
		Journal: nil,
	})
	if err == nil {
		t.Fatal("an observer with no journal must not be constructible")
	}
}

func TestTheObserverRefusesAnUnconfiguredCostModel(t *testing.T) {
	h := newExitHarness(t, nil)
	_, err := engine.NewExitObserver(engine.ExitObserverOptions{
		Journal: h.journal, Prices: h.prices,
		Retrier:    &execgw.Retrier{Clock: h.clk},
		Issuer:     stubIssuer{},
		Submit:     h.submit,
		AccountRef: exitAccount,
	})
	if err == nil || !strings.Contains(err.Error(), "cost model") {
		t.Fatalf("err = %v, want a refusal naming the cost model", err)
	}
}

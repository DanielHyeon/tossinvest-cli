package engine_test

// tracer_test.go is task 8.2's verification: the tracer slice driven end to end
// against the httptest broker, plus the parameter refusals that are the only
// thing standing between "a bounded experiment" and "an automated trader with
// no strategy".
//
// The live run is the verify track's (D8). Every order the tracer places needs a
// GuardianDecision, and while ProtectionReady is UNWIRED the gateway refuses
// the exposure-raising entry even though the verified runtime itself may start.

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/app/engine"
	"github.com/JungHoonGhae/tossinvest-cli/internal/costs"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/risk"
)

// tracerParams is the smallest set that validates: one symbol, LIMIT, one unit.
func tracerParams() engine.TracerParams {
	return engine.TracerParams{
		Symbol:          "005930",
		Market:          "kr",
		Quantity:        "1",
		LimitPrice:      "70000",
		StopPrice:       "68000",
		TargetPrice:     "76000",
		NotionalCeiling: "100000",
		Freshness:       15 * time.Second,
		MaxCycles:       20,
		MaxDuration:     5 * time.Minute,
	}
}

// newTracer builds the tracer over the same real stack the end-to-end test uses.
func newTracer(t *testing.T, s *e2eStack, params engine.TracerParams) *engine.Tracer {
	t.Helper()
	guardian, err := execgw.NewRiskGuardian(execgw.RiskGuardianOptions{
		Journal: s.engine.Journal, Clock: s.clk, AccountRef: s.engine.AccountRef,
		Policy: risk.DefaultPolicy(), Costs: costs.DefaultModel(),
		PolicyVersion: "add-core-domain/8.2",
	})
	if err != nil {
		t.Fatalf("NewRiskGuardian: %v", err)
	}
	tracer, err := engine.NewTracer(engine.TracerOptions{
		Journal: s.engine.Journal, Issuer: guardian, Submit: s.engine.Gateway,
		Observer: s.observer, Retrier: s.engine.Retrier, Reads: s.engine.Official,
		Clock: s.clk, AccountRef: s.engine.AccountRef, Params: params,
	})
	if err != nil {
		t.Fatalf("NewTracer: %v", err)
	}
	return tracer
}

// --- the parameter surface --------------------------------------------------------

func TestTheTracerParametersAreAllRequired(t *testing.T) {
	cases := []struct {
		name   string
		break_ func(*engine.TracerParams)
		want   string
	}{
		{"no symbol", func(p *engine.TracerParams) { p.Symbol = "" }, "one named symbol"},
		{"two symbols", func(p *engine.TracerParams) { p.Symbol = "005930,000660" }, "exactly one"},
		{"unknown market", func(p *engine.TracerParams) { p.Market = "jp" }, "market"},
		{"fractional quantity", func(p *engine.TracerParams) { p.Quantity = "0.5" }, "whole number"},
		{"zero quantity", func(p *engine.TracerParams) { p.Quantity = "0" }, "whole number"},
		{"no limit price", func(p *engine.TracerParams) { p.LimitPrice = "" }, "LIMIT only"},
		{"no stop", func(p *engine.TracerParams) { p.StopPrice = "" }, "No Stop = No Trade"},
		{"stop above entry", func(p *engine.TracerParams) { p.StopPrice = "71000" }, "not below the entry"},
		{"target below entry", func(p *engine.TracerParams) { p.TargetPrice = "69000" }, "not above the entry"},
		{"no ceiling", func(p *engine.TracerParams) { p.NotionalCeiling = "" }, "notional ceiling"},
		{"over the ceiling", func(p *engine.TracerParams) { p.NotionalCeiling = "1000" }, "exceeds the notional ceiling"},
		{"no freshness", func(p *engine.TracerParams) { p.Freshness = 0 }, "freshness bound"},
		{"no cycle budget", func(p *engine.TracerParams) { p.MaxCycles = 0 }, "abort criteria"},
		{"no time budget", func(p *engine.TracerParams) { p.MaxDuration = 0 }, "abort criteria"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p := tracerParams()
			tc.break_(&p)
			err := p.Validate()
			if !errors.Is(err, engine.ErrTracerRefused) {
				t.Fatalf("err = %v, want a tracer refusal", err)
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Errorf("err = %q, want it to mention %q", err, tc.want)
			}
		})
	}
}

func TestTheSmallestValidParameterSetPasses(t *testing.T) {
	if err := tracerParams().Validate(); err != nil {
		t.Fatalf("the documented minimum must validate: %v", err)
	}
	if got := tracerParams().Notional(); got != "70000" {
		t.Errorf("notional = %s, want 70000", got)
	}
}

// --- the run ------------------------------------------------------------------------

// TestScalarStartupReadinessCannotAuthorizeTheTracer proves that the legacy
// engine status seam is reporting-only. Even when that test seam reports entry
// permitted, the shipped Gateway still requires a sealed market snapshot.
func TestScalarStartupReadinessCannotAuthorizeTheTracer(t *testing.T) {
	s := newEntryCapableE2EStack(t)
	s.broker.quote("005930", "70000")
	tracer := newTracer(t, s, tracerParams())

	report, err := tracer.Run(context.Background())
	if !errors.Is(err, engine.ErrTracerRefused) || !strings.Contains(strings.ToLower(err.Error()), "protection") {
		t.Fatalf("tracer err = %v, want sealed protection-readiness refusal", err)
	}
	if report.EntryOrderID != "" || report.Closed || report.Outcome != nil {
		t.Fatalf("refused tracer changed trade state: %+v", report)
	}
}

// runTracerWithFills drives the tracer while feeding the broker's fills in.
//
// Two things have to happen that are not the tracer's: the fake clock has to
// advance so the run's interval sleeps return, and the broker has to answer the
// two fills the run needs — the entry, and the liquidation the exit policy
// submits once the price falls through the baseline. In production the first is
// wall time and the second is the fill detector.
//
// The clock is advanced *only* while this function is waiting for something,
// which is what keeps the run deterministic: between the steps below the tracer
// is parked in Sleep and cannot burn its cycle budget on a world the test has
// not finished setting up.
func runTracerWithFills(t *testing.T, s *e2eStack, tracer *engine.Tracer,
	ctx context.Context) (engine.TracerReport, error) {
	t.Helper()

	out := make(chan result, 1)
	go func() {
		report, err := tracer.Run(ctx)
		out <- result{report: report, err: err}
	}()

	// The entry order reaches the broker, then fills. The wait is on the
	// *journal* and not on the broker handler: the create body arrives before
	// the gateway has finished settling the attempt, and a fill recorded in that
	// window resolves to no order and produces no projection.
	driveUntilOrders(t, s, out, 1)
	entry := waitForJournaledOrder(t, s, out)
	s.fill(entry, "005930", "BUY", "1", "1", "70000", true)

	// The price falls through the entry stop: the exit policy liquidates.
	s.broker.quote("005930", "67900")
	driveUntilOrders(t, s, out, 2)
	liquidation := waitForJournaledOrder(t, s, out)
	s.fill(liquidation, "005930", "SELL", "1", "1", "67900", true)

	// One more cycle for the run to notice the close.
	r, ok := drive(t, s, out, func() bool { return false })
	if !ok {
		t.Fatal("the tracer run did not finish")
	}
	return r.report, r.err
}

// result is one finished tracer run.
type result struct {
	report engine.TracerReport
	err    error
}

// drive advances the fake clock until the run ends or done() reports the test
// has what it was waiting for.
func drive(t *testing.T, s *e2eStack, out <-chan result, done func() bool) (result, bool) {
	t.Helper()
	for i := 0; i < 4000; i++ {
		if done() {
			return result{}, false
		}
		select {
		case r := <-out:
			return r, true
		default:
		}
		if s.clk.Sleepers() > 0 {
			s.clk.Advance(s.observer.Interval())
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal("the tracer made no progress")
	return result{}, false
}

// waitForJournaledOrder blocks until the most recent broker order is one the
// journal can resolve, and returns its id.
func waitForJournaledOrder(t *testing.T, s *e2eStack, out <-chan result) string {
	t.Helper()
	want := s.lastBrokerOrder()
	drive(t, s, out, func() bool {
		live, err := s.engine.Journal.LiveOrdersForSymbol(
			context.Background(), s.engine.AccountRef, "kr", "005930")
		if err != nil {
			return false
		}
		for _, o := range live {
			if o.OrderID == want {
				return true
			}
		}
		return false
	})
	return want
}

func driveUntilOrders(t *testing.T, s *e2eStack, out <-chan result, want int) {
	t.Helper()
	r, ended := drive(t, s, out, func() bool { return len(s.broker.placed()) >= want })
	if ended {
		t.Fatalf("the run ended before placing %d order(s): report=%+v err=%v", want, r.report, r.err)
	}
}

// TestTheTracerRefusesANonFlatAccount: a tracer measures one trade, and on an
// account that already holds something it can neither size the entry nor tell
// its own position from the others.
func TestTheTracerRefusesANonFlatAccount(t *testing.T) {
	s := newE2EStack(t)
	s.entry("005930", "10", "70000", "68000")
	s.broker.quote("005930", "70000")

	tracer := newTracer(t, s, tracerParams())
	_, err := tracer.Run(context.Background())
	if !errors.Is(err, engine.ErrTracerRefused) {
		t.Fatalf("err = %v, want a refusal", err)
	}
	if !strings.Contains(err.Error(), "already holds") {
		t.Errorf("err = %q, want it to name the holding", err)
	}
	if len(s.broker.placed()) != 0 {
		t.Error("a refused run must place nothing")
	}
}

// TestTheTracerRefusesAPriceItCannotRead: an entry submitted without a fresh
// price would be refused by the gateway as QUERY_STALE a moment later, and the
// tracer says so at the point the operator can act on it.
func TestTheTracerRefusesAPriceItCannotRead(t *testing.T) {
	s := newE2EStack(t)
	// No quote is registered, so the broker answers with an empty list.
	tracer := newTracer(t, s, tracerParams())

	_, err := tracer.Run(context.Background())
	if !errors.Is(err, engine.ErrTracerRefused) {
		t.Fatalf("err = %v, want a refusal", err)
	}
	if !strings.Contains(err.Error(), "no last trade") {
		t.Errorf("err = %q, want it to name the missing price", err)
	}
}

// TestTheTracerNeedsAnObserver is the invariant that makes the slice a slice:
// an entry with nothing watching it is an unprotected position, which is the one
// thing a tracer must never produce.
func TestTheTracerNeedsAnObserver(t *testing.T) {
	s := newE2EStack(t)
	_, err := engine.NewTracer(engine.TracerOptions{
		Journal: s.engine.Journal, Issuer: stubEntryIssuer{}, Submit: s.engine.Gateway,
		Retrier: s.engine.Retrier, Reads: s.engine.Official,
		AccountRef: s.engine.AccountRef, Params: tracerParams(),
	})
	if !errors.Is(err, engine.ErrTracerRefused) {
		t.Fatalf("err = %v, want a refusal", err)
	}
	if !strings.Contains(err.Error(), "nothing watching it") {
		t.Errorf("err = %q, want it to say why", err)
	}
}

type stubEntryIssuer struct{}

func (stubEntryIssuer) IssueEntry(context.Context, execgw.EntryIssuance) (execgw.Issued, error) {
	return execgw.Issued{}, errors.New("not reached")
}

// TestALiveTracerRunIsBlockedByTheProtectionRefusal is the D8 claim, asserted
// rather than assumed — with the reason it is true now stated correctly.
//
// It used to be true by construction: the tracer's orders need a Guardian, the
// engine only holds one behind a verified gate, and clause 6 made a verified gate
// unreachable, so there was no Context to run a tracer with at all. Since
// interlock-gates-entry-not-exit the Context exists, and the claim rests on two
// narrower facts instead:
//
//   - EntryPermitted is false, asserted here. Everything the tracer submits is a
//     buy, and this is the runtime saying so before one is attempted.
//   - The submission itself is refused at the gateway, which is asserted where
//     the refusal lives (internal/execgw's protection_test.go) rather than
//     re-driven through the tracer here. That is deliberate: a second assertion
//     of the same guard in a second place is a second thing to get wrong, which
//     is the reasoning tracer.go already gives for not adding a second gate.
//
// The third fact, that nothing in the runtime can reach the tracer at all, is
// entryreach_test.go.
func TestALiveTracerRunIsBlockedByTheProtectionRefusal(t *testing.T) {
	dir := isolate(t)
	writeGateConfig(t, dir, smallLiveGate())
	writeCredentials(t, dir, "test-api-key-000000", "test-secret")
	writeAttestation(t, dir, nil)
	srv, _ := interlockServer(t, "123-45")

	eng, err := openGateEngine(t, dir, srv, realGuardian(t, risk.DefaultPolicy()))
	if err != nil {
		t.Fatalf("the gate must verify on the small_live set: %v", err)
	}
	if eng.Automation.EntryPermitted {
		t.Fatal("EntryPermitted = true: the tracer's buys would be authorised on a build " +
			"that leaves no protective order at the broker")
	}
	if eng.Automation.Protection != engine.ProtectionUnwired {
		t.Errorf("Protection = %q, want UNWIRED", eng.Automation.Protection)
	}
}

// TestTheExitObserverIsUnavailableWithoutAVerifiedGate is the same refusal from
// the other side: even holding a Context, the loop that would place unattended
// orders cannot be assembled with the master switch off.
func TestTheExitObserverIsUnavailableWithoutAVerifiedGate(t *testing.T) {
	s := newE2EStack(t)
	_, err := s.engine.ExitObserver(engine.ExitObserverOptions{Costs: costs.DefaultModel()})
	if !errors.Is(err, engine.ErrExitObserverUnavailable) {
		t.Fatalf("err = %v, want ErrExitObserverUnavailable", err)
	}
	if !strings.Contains(err.Error(), "automation gate") {
		t.Errorf("err = %q, want it to name the gate", err)
	}
}

var _ = journal.PositionOpen // the fixtures above assert against the projection's states

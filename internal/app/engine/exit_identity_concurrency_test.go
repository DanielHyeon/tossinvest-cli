package engine_test

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/app/engine"
	"github.com/JungHoonGhae/tossinvest-cli/internal/costs"
	"github.com/JungHoonGhae/tossinvest-cli/internal/domain"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/exitpolicy"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
)

type concurrentQuoteReader struct {
	mu      sync.Mutex
	waiters int
	ready   chan struct{}
	quote   domain.Quote
}

func (r *concurrentQuoteReader) Prices(_ context.Context, _ []string) ([]domain.Quote, error) {
	r.mu.Lock()
	r.waiters++
	if r.waiters == 2 {
		close(r.ready)
	}
	r.mu.Unlock()
	<-r.ready
	return []domain.Quote{r.quote}, nil
}

type journalMutationSubmitter struct {
	journal *journal.Journal
	calls   atomic.Int32
	mu      sync.Mutex
	places  []execgw.PlaceRequest
}

func (s *journalMutationSubmitter) Place(ctx context.Context, req execgw.PlaceRequest) (execgw.Outcome, error) {
	n := s.calls.Add(1)
	s.mu.Lock()
	s.places = append(s.places, req)
	s.mu.Unlock()
	orderID := fmt.Sprintf("O-concurrent-%d", n)
	attempt, err := s.journal.Prepare(ctx, journal.PrepareRequest{
		Intent: journal.Intent{
			ID: req.IntentID, Market: "kr", TradingDay: "2026-03-30", AccountRef: exitAccount,
			Symbol: req.Intent.Symbol, Side: "SELL", OrderType: "LIMIT", TimeInForce: "DAY",
			Quantity: decimalText(req.Intent.Quantity), Price: decimalText(req.Intent.Price),
			Currency: "KRW", Source: "engine/concurrency-test", Fingerprint: "fp-" + req.IntentID,
		},
		Kind: journal.KindPlace, AttemptID: "a-" + orderID, AccountRef: exitAccount,
		DecisionID: req.Decision.ID, SafetyClass: journal.SafetyClassRiskReducing,
		ClientOrderID: journal.DeriveClientOrderID(req.Decision.ID, 0),
	})
	if err != nil {
		return execgw.Outcome{}, err
	}
	if err := attempt.MarkDispatchStarted(ctx); err != nil {
		return execgw.Outcome{}, err
	}
	if err := attempt.MarkAcked(ctx, orderID); err != nil {
		return execgw.Outcome{}, err
	}
	if err := attempt.Settle(ctx, journal.StateConfirmed, "broker_accepted", ""); err != nil {
		return execgw.Outcome{}, err
	}
	return execgw.Outcome{IntentID: req.IntentID, State: journal.StateConfirmed, BrokerOrderID: orderID}, nil
}

func (s *journalMutationSubmitter) Cancel(context.Context, execgw.CancelRequest) (execgw.Outcome, error) {
	return execgw.Outcome{State: journal.StateConfirmed}, nil
}

func (s *journalMutationSubmitter) requests() []execgw.PlaceRequest {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]execgw.PlaceRequest(nil), s.places...)
}

func TestExitObserverConcurrentSameQuoteArmsAndMutatesOnce(t *testing.T) {
	h := newExitHarness(t, nil)
	p := h.entry("005930", "10", "10000", "9800", "10000")
	identity, err := exitpolicy.LegacyRatchetPolicyIdentity()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := h.journal.OpenExitState(context.Background(), journal.ExitStateSeed{
		PositionID: p.ID, PolicyKind: journal.ExitPolicyRatchet, PolicyIdentity: identity,
		EntryPrice: "10000", InitialStop: "9800",
	}); err != nil {
		t.Fatal(err)
	}

	prices := &concurrentQuoteReader{
		ready: make(chan struct{}),
		quote: domain.Quote{Symbol: "005930", Last: 9700, FetchedAt: exitNow.Add(17 * time.Second)},
	}
	submit := &journalMutationSubmitter{journal: h.journal}
	observers := []*engine.ExitObserver{
		newConcurrentObserver(t, h, prices, submit),
		newConcurrentObserver(t, h, prices, submit),
	}

	cycles := make(chan engine.ExitCycle, len(observers))
	for _, observer := range observers {
		go func(o *engine.ExitObserver) { cycles <- o.ObserveOnce(context.Background()) }(observer)
	}
	first, second := <-cycles, <-cycles
	if first.Err != nil || second.Err != nil {
		t.Fatalf("cycles: first=%+v second=%+v", first, second)
	}
	if got := first.Proposed + second.Proposed; got != 1 {
		t.Fatalf("proposals = %d, want exactly one (cycles %+v %+v)", got, first, second)
	}
	if got := submit.calls.Load(); got != 1 {
		t.Fatalf("mutation calls = %d, want exactly one", got)
	}
	reqs := submit.requests()
	if len(reqs) != 1 || reqs[0].ExitProvenance == nil {
		t.Fatalf("submitted provenance = %+v", reqs)
	}
	provenance := reqs[0].ExitProvenance
	if !strings.HasPrefix(provenance.ObservationID, "obs_") || len(provenance.ObservationID) != len("obs_")+64 {
		t.Fatalf("observation id = %q, want opaque sha256 id", provenance.ObservationID)
	}
	if strings.Contains(provenance.ObservationID, exitAccount) {
		t.Fatalf("observation id leaked account reference: %q", provenance.ObservationID)
	}
	if reqs[0].IntentID != "exit_"+strings.TrimPrefix(provenance.DecisionID, "eld_") {
		t.Fatalf("intent id = %q, decision = %q", reqs[0].IntentID, provenance.DecisionID)
	}
	state := h.state(p.ID)
	if !state.Pending() || state.PendingIntentID != reqs[0].IntentID {
		t.Fatalf("pending state = %+v, want winning deterministic intent", state)
	}
	attempted, err := h.journal.IntentAttempted(context.Background(), reqs[0].IntentID)
	if err != nil || !attempted {
		t.Fatalf("winning mutation was not recorded: attempted=%v err=%v", attempted, err)
	}
}

func TestLegacyLadderStateRefusesAReinterpretedSameIDAndVersion(t *testing.T) {
	mutated := exitpolicy.DefaultLadderPolicy()
	mutated.Rungs[0] = exitpolicy.Rung{TargetPct: "0.5", StopPct: "0", PartialRatio: "1"}
	mutated.PolicyDigest = ""
	h := newExitHarness(t, func(opts *engine.ExitObserverOptions) { opts.Ladder = &mutated })
	p := h.entry("005930", "10", "10000", "9800", "10000")
	if _, err := h.journal.OpenExitState(context.Background(), journal.ExitStateSeed{
		PositionID: p.ID, PolicyKind: journal.ExitPolicyLadder, PolicyID: "default_v1",
		EntryPrice: "10000", InitialStop: "9800",
	}); err != nil {
		t.Fatal(err)
	}
	h.quote("005930", 10100)
	cycle := h.observe()
	if cycle.Err != nil || cycle.Proposed != 0 || len(h.submit.places) != 0 {
		t.Fatalf("reinterpreted state reached order path: cycle=%+v places=%d", cycle, len(h.submit.places))
	}
	if got := h.alerts.count(obs.EventExitJudgementRefused); got != 1 {
		t.Fatalf("judgement refusal alerts = %d, want 1", got)
	}
}

func newConcurrentObserver(t *testing.T, h *exitHarness, prices engine.PriceReader, submit engine.ExitSubmitter) *engine.ExitObserver {
	t.Helper()
	gate := execgw.NewEntryGate(h.clk, map[execgw.RequiredQuery]time.Duration{execgw.QueryPrice: 15 * time.Second})
	guardian, err := execgw.NewRiskGuardian(execgw.RiskGuardianOptions{
		Journal: h.journal, Clock: h.clk, AccountRef: exitAccount,
		Policy: exitPolicy(), Costs: costs.DefaultModel(), PolicyVersion: "test/v1",
	})
	if err != nil {
		t.Fatal(err)
	}
	observer, err := engine.NewExitObserver(engine.ExitObserverOptions{
		Journal: h.journal, Prices: prices,
		Retrier: &execgw.Retrier{Clock: h.clk, Gate: gate, Policy: execgw.RetryPolicy{MaxAttempts: 1, Budget: time.Second}},
		Issuer:  guardian, Submit: submit, Costs: costs.DefaultModel(), AccountRef: exitAccount, Clock: h.clk,
	})
	if err != nil {
		t.Fatal(err)
	}
	return observer
}

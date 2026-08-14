package engine_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/app/engine"
	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/costs"
	"github.com/JungHoonGhae/tossinvest-cli/internal/domain"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/exitpolicy"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
)

// These tests deliberately use the real exit observer, journal, policy
// evaluators and Guardian.  Only the broker read and broker submission surfaces
// are substituted.  That keeps a flat-quote optimization from passing by
// bypassing the durable state and order-safety contracts it is meant to retain.

type a111BatchPrices struct {
	batches      [][]domain.Quote
	err          error
	calls        int
	asked        [][]string
	beforeReturn func(int)
}

func (p *a111BatchPrices) Prices(_ context.Context, symbols []string) ([]domain.Quote, error) {
	p.calls++
	p.asked = append(p.asked, append([]string(nil), symbols...))
	if p.beforeReturn != nil {
		p.beforeReturn(p.calls)
	}
	if p.err != nil {
		return nil, p.err
	}
	if len(p.batches) == 0 {
		return nil, nil
	}
	idx := p.calls - 1
	if idx >= len(p.batches) {
		idx = len(p.batches) - 1
	}
	return append([]domain.Quote(nil), p.batches[idx]...), nil
}

type a111CountingIssuer struct {
	delegate engine.ReductionIssuer
	calls    int
	symbols  []string
}

func (i *a111CountingIssuer) IssueReduction(ctx context.Context,
	req execgw.ReductionIssuance) (execgw.Issued, error) {
	i.calls++
	i.symbols = append(i.symbols, req.Intent.Symbol)
	return i.delegate.IssueReduction(ctx, req)
}

type a111SubmitSpy struct {
	delegate     engine.ExitSubmitter
	places       int
	cancels      int
	beforePlace  func(execgw.PlaceRequest)
	beforeCancel func(execgw.CancelRequest)
}

// a111RecordBoundaryClock advances only when the observer has entered a pure
// policy judgement/record stack.  It does not depend on a brittle global Now
// call number: a check only before judge sees the base time, while the required
// immediate pre-record/refresh check sees an expired lease.
type a111RecordBoundaryClock struct {
	base     *clock.Fake
	advanced bool
}

func (c *a111RecordBoundaryClock) Now() time.Time {
	pcs := make([]uintptr, 24)
	n := runtime.Callers(2, pcs)
	frames := runtime.CallersFrames(pcs[:n])
	for {
		frame, more := frames.Next()
		if strings.Contains(frame.Function, ".(*ExitObserver).judgeRatchet") ||
			strings.Contains(frame.Function, ".(*ExitObserver).judgeLadder") ||
			strings.Contains(frame.Function, ".(*ExitObserver).record") {
			if !c.advanced {
				c.base.Advance(16 * time.Second)
				c.advanced = true
			}
			break
		}
		if !more {
			break
		}
	}
	return c.base.Now()
}

func (c *a111RecordBoundaryClock) Since(at time.Time) time.Duration { return c.Now().Sub(at) }

func (c *a111RecordBoundaryClock) Sleep(ctx context.Context, d time.Duration) error {
	return c.base.Sleep(ctx, d)
}

// a111WallElapsedClock makes wall movement and elapsed movement independently
// controllable. A lease must use both: elapsed time prevents a wall rollback
// from extending evidence, while the wall bound prevents using evidence before
// its official source timestamp.
type a111WallElapsedClock struct {
	base *clock.Fake

	mu            sync.Mutex
	wallOffset    time.Duration
	elapsedOffset time.Duration
}

func (c *a111WallElapsedClock) Now() time.Time {
	c.mu.Lock()
	offset := c.wallOffset
	c.mu.Unlock()
	return c.base.Now().Add(offset)
}

func (c *a111WallElapsedClock) Since(at time.Time) time.Duration {
	c.mu.Lock()
	elapsed := c.elapsedOffset
	c.mu.Unlock()
	return c.base.Now().Sub(at) + elapsed
}

func (c *a111WallElapsedClock) Sleep(ctx context.Context, d time.Duration) error {
	return c.base.Sleep(ctx, d)
}

func (c *a111WallElapsedClock) advanceElapsed(d time.Duration) {
	c.mu.Lock()
	c.elapsedOffset += d
	c.mu.Unlock()
}

func (c *a111WallElapsedClock) rollbackWall(d time.Duration) {
	c.mu.Lock()
	c.wallOffset -= d
	c.mu.Unlock()
}

func (s *a111SubmitSpy) Place(ctx context.Context, req execgw.PlaceRequest) (execgw.Outcome, error) {
	s.places++
	if s.beforePlace != nil {
		s.beforePlace(req)
	}
	return s.delegate.Place(ctx, req)
}

func (s *a111SubmitSpy) Cancel(ctx context.Context, req execgw.CancelRequest) (execgw.Outcome, error) {
	s.cancels++
	if s.beforeCancel != nil {
		s.beforeCancel(req)
	}
	return s.delegate.Cancel(ctx, req)
}

func a111SeedRatchet(t *testing.T, h *exitHarness, p journal.Position) {
	t.Helper()
	if _, err := h.journal.OpenExitState(context.Background(), journal.ExitStateSeed{
		PositionID: p.ID, PolicyKind: journal.ExitPolicyRatchet,
		EntryPrice: "70000", InitialStop: "68000",
	}); err != nil {
		t.Fatalf("OpenExitState: %v", err)
	}
}

func a111Events(t *testing.T, j *journal.Journal, positionID string) []journal.ExitEvent {
	t.Helper()
	events, err := j.ExitEvents(context.Background(), positionID)
	if err != nil {
		t.Fatalf("ExitEvents: %v", err)
	}
	return events
}

func a111Snapshot(t *testing.T, state journal.ExitState) *journal.StoredExitSnapshot {
	t.Helper()
	if state.Snapshot.Snapshot == nil {
		t.Fatalf("state has no effective snapshot: %+v", state)
	}
	return state.Snapshot.Snapshot
}

func a111ObserverForJournal(t *testing.T, j *journal.Journal, account string,
	clk clock.Clock, prices engine.PriceReader, gate *execgw.EntryGate,
	submit engine.ExitSubmitter) (*engine.ExitObserver, *a111CountingIssuer) {
	t.Helper()
	guardian, err := execgw.NewRiskGuardian(execgw.RiskGuardianOptions{
		Journal: j, Clock: clk, AccountRef: account,
		Policy: exitPolicy(), Costs: costs.DefaultModel(), PolicyVersion: "a111-test/v1",
	})
	if err != nil {
		t.Fatalf("NewRiskGuardian: %v", err)
	}
	issuer := &a111CountingIssuer{delegate: guardian}
	observer, err := engine.NewExitObserver(engine.ExitObserverOptions{
		Journal: j, Prices: prices,
		Retrier: &execgw.Retrier{Clock: clk, Gate: gate,
			Policy: execgw.RetryPolicy{MaxAttempts: 1, Budget: time.Second}},
		Issuer: issuer, Submit: submit, Alerts: &fakeAlerts{}, Costs: costs.DefaultModel(),
		Floor: &fakeFloor{}, SLO: &fakeSLO{}, Escalate: j,
		AccountRef: account, Clock: clk,
		NewID: func() string { return "a111-exit-intent" },
	})
	if err != nil {
		t.Fatalf("NewExitObserver: %v", err)
	}
	return observer, issuer
}

func TestA111UnchangedFirstRatchetQuoteCreatesOneEvaluatedSnapshotWithoutSideEffects(t *testing.T) {
	prices := &a111BatchPrices{}
	var issuer *a111CountingIssuer
	var submit *a111SubmitSpy
	h := newExitHarness(t, func(opts *engine.ExitObserverOptions) {
		issuer = &a111CountingIssuer{delegate: opts.Issuer}
		submit = &a111SubmitSpy{delegate: opts.Submit}
		opts.Prices, opts.Issuer, opts.Submit = prices, issuer, submit
	})
	p := h.entry("005930", "10", "70000", "68000", "70000")
	a111SeedRatchet(t, h, p)
	before := len(a111Events(t, h.journal, p.ID))
	prices.batches = [][]domain.Quote{{{
		Symbol: "005930", Last: 70000, FetchedAt: h.clk.Now(),
	}}}

	cycle := h.observe()
	if cycle.Err != nil {
		t.Fatalf("ObserveOnce: %v", cycle.Err)
	}
	state := h.state(p.ID)
	if state.SnapshotStatus != journal.SnapshotStatusEvaluated {
		t.Fatalf("snapshot status = %q, want EVALUATED", state.SnapshotStatus)
	}
	snapshot := a111Snapshot(t, state)
	if snapshot.ObservationSource != "quote_fetched_at" || snapshot.ObservedAt == "" {
		t.Fatalf("official first quote lost its source time: %+v", snapshot)
	}
	if snapshot.Line.CurrentProtection != "68000" || snapshot.Line.NextTarget != "70800" ||
		snapshot.Line.NextProtection != "69000" {
		t.Fatalf("first flat ratchet line = %+v", snapshot.Line)
	}
	events := a111Events(t, h.journal, p.ID)
	if len(events) != before+1 {
		t.Fatalf("event count = %d, want %d", len(events), before+1)
	}
	last := events[len(events)-1]
	if last.Action != "" || last.ProposedIntentID != "" {
		t.Fatalf("flat seed evaluation carried order evidence: %+v", last)
	}
	if last.Evaluation.Effective.Snapshot == nil ||
		last.Evaluation.Effective.Snapshot.Line.SnapshotID != snapshot.Line.SnapshotID {
		t.Fatalf("semantic event and effective state do not carry the same snapshot: event=%+v state=%+v",
			last.Evaluation.Effective, snapshot)
	}
	if cycle.Proposed != 0 || state.Pending() || issuer.calls != 0 || submit.places != 0 || submit.cancels != 0 {
		t.Fatalf("flat seed evaluation caused side effects: cycle=%+v issuer=%d place=%d cancel=%d state=%+v",
			cycle, issuer.calls, submit.places, submit.cancels, state)
	}
}

func TestA111AutomaticallyAdoptedCommonLadderEvaluatesItsUnchangedFirstQuote(t *testing.T) {
	ctx := context.Background()
	h := newDriverHarness(t, func(opts *engine.ReconcileDriverOptions) {
		opts.CommonPolicy = exitpolicy.CommonLadderBalanced
	})
	h.holds("005930", "10", "55000", 70000)
	adopted := h.cycle()
	if adopted.Err != nil || adopted.Adopted != 1 {
		t.Fatalf("automatic adoption: %+v", adopted)
	}
	p := h.position("005930")
	seed, err := h.journal.ExitState(ctx, p.ID)
	if err != nil {
		t.Fatal(err)
	}
	if seed.SnapshotStatus != journal.SnapshotStatusSeed || seed.PolicyID != exitpolicy.CommonLadderBalanced {
		t.Fatalf("adoption did not leave the selected common ladder at SEED: %+v", seed)
	}
	before := len(a111Events(t, h.journal, p.ID))
	prices := &a111BatchPrices{batches: [][]domain.Quote{{
		{Symbol: "005930", Last: 70000, Currency: "KRW", FetchedAt: h.clk.Now()},
	}}}
	gate := execgw.NewEntryGate(h.clk, map[execgw.RequiredQuery]time.Duration{
		execgw.QueryPrice: execgw.QueryPriceEvidenceDuration,
	})
	broker := &fakeSubmitter{}
	submit := &a111SubmitSpy{delegate: broker}
	observer, issuer := a111ObserverForJournal(t, h.journal, reconcileAccount, h.clk, prices, gate, submit)

	cycle := observer.ObserveOnce(ctx)
	if cycle.Err != nil {
		t.Fatalf("ObserveOnce: %v", cycle.Err)
	}
	state, err := h.journal.ExitState(ctx, p.ID)
	if err != nil {
		t.Fatal(err)
	}
	snapshot := a111Snapshot(t, state)
	if state.SnapshotStatus != journal.SnapshotStatusEvaluated ||
		snapshot.Line.CurrentProtection != "66500" || snapshot.Line.NextTarget != "71050" ||
		snapshot.Line.NextProtection != "70000" {
		t.Fatalf("first flat common-ladder line = %+v, state=%+v", snapshot.Line, state)
	}
	if len(a111Events(t, h.journal, p.ID)) != before+1 || cycle.Proposed != 0 ||
		issuer.calls != 0 || submit.places != 0 || submit.cancels != 0 || state.Pending() {
		t.Fatalf("common-ladder flat evaluation was not one event with zero side effects: cycle=%+v state=%+v",
			cycle, state)
	}
	first := *snapshot
	eventCount := len(a111Events(t, h.journal, p.ID))
	for pass := 0; pass < 2; pass++ {
		h.clk.Advance(20 * time.Second)
		prices.batches = append(prices.batches, []domain.Quote{{
			Symbol: "005930", Last: 70000, Currency: "KRW", FetchedAt: h.clk.Now(),
		}})
		if cycle := observer.ObserveOnce(ctx); cycle.Err != nil || cycle.Proposed != 0 {
			t.Fatalf("repeated common-ladder flat quote %d: %+v", pass+1, cycle)
		}
	}
	latest := a111Snapshot(t, stateFromJournal(t, h.journal, p.ID))
	if latest.ObservedAt == first.ObservedAt || latest.Line.ObservationID == first.Line.ObservationID {
		t.Fatalf("common ladder used a SEED-only exception instead of durable refresh: first=%+v latest=%+v",
			first, latest)
	}
	if fresh := stateFromJournal(t, h.journal, p.ID).Snapshot.WithFreshness(
		h.clk.Now().Add(30*time.Second), 30*time.Second); fresh.Stale {
		t.Fatalf("repeated common-ladder quote did not keep the line fresh: %+v", fresh)
	}
	if len(a111Events(t, h.journal, p.ID)) != eventCount || prices.calls != 3 ||
		issuer.calls != 0 || submit.places != 0 || submit.cancels != 0 {
		t.Fatalf("repeated common-ladder refresh grew history or side effects: events=%d/%d calls=%d issuer=%d place=%d cancel=%d",
			len(a111Events(t, h.journal, p.ID)), eventCount, prices.calls,
			issuer.calls, submit.places, submit.cancels)
	}
}

func stateFromJournal(t *testing.T, j *journal.Journal, positionID string) journal.ExitState {
	t.Helper()
	state, err := j.ExitState(context.Background(), positionID)
	if err != nil {
		t.Fatalf("ExitState: %v", err)
	}
	return state
}

func TestA111FlatAndInBandQuotesRefreshProvenanceWithoutHistoryOrOrders(t *testing.T) {
	prices := &a111BatchPrices{}
	var issuer *a111CountingIssuer
	var submit *a111SubmitSpy
	h := newExitHarness(t, func(opts *engine.ExitObserverOptions) {
		issuer = &a111CountingIssuer{delegate: opts.Issuer}
		submit = &a111SubmitSpy{delegate: opts.Submit}
		opts.Prices, opts.Issuer, opts.Submit = prices, issuer, submit
	})
	p := h.entry("005930", "10", "70000", "68000", "70000")
	a111SeedRatchet(t, h, p)
	prices.batches = [][]domain.Quote{{{Symbol: "005930", Last: 70100, FetchedAt: h.clk.Now()}}}
	if cycle := h.observe(); cycle.Err != nil {
		t.Fatalf("initial changed evaluation: %v", cycle.Err)
	}
	first := *a111Snapshot(t, h.state(p.ID))
	eventCount := len(a111Events(t, h.journal, p.ID))

	h.clk.Advance(20 * time.Second)
	prices.batches = append(prices.batches,
		[]domain.Quote{{Symbol: "005930", Last: 70100, FetchedAt: h.clk.Now()}})
	if cycle := h.observe(); cycle.Err != nil {
		t.Fatalf("identical refresh: %v", cycle.Err)
	}
	identical := *a111Snapshot(t, h.state(p.ID))
	if identical.ObservedAt == first.ObservedAt ||
		identical.Line.ObservationID == first.Line.ObservationID {
		t.Fatalf("the second identical quote did not refresh provenance: first=%+v second=%+v",
			first, identical)
	}
	if got := len(a111Events(t, h.journal, p.ID)); got != eventCount {
		t.Fatalf("the second identical quote appended an event: got %d want %d", got, eventCount)
	}
	h.clk.Advance(20 * time.Second)
	prices.batches = append(prices.batches,
		[]domain.Quote{{Symbol: "005930", Last: 70050, FetchedAt: h.clk.Now()}})
	if cycle := h.observe(); cycle.Err != nil {
		t.Fatalf("in-band refresh: %v", cycle.Err)
	}
	state := h.state(p.ID)
	latest := a111Snapshot(t, state)
	if latest.ObservedAt == first.ObservedAt || latest.Line.ObservationID == first.Line.ObservationID ||
		latest.Line.ObservedPrice != "70050" {
		t.Fatalf("flat refresh did not replace complete observation provenance: first=%+v latest=%+v", first, latest)
	}
	if fresh := state.Snapshot.WithFreshness(h.clk.Now().Add(30*time.Second), 30*time.Second); fresh.Stale {
		t.Fatalf("latest flat observation is already stale at the inclusive boundary: %+v", fresh)
	}
	if got := len(a111Events(t, h.journal, p.ID)); got != eventCount {
		t.Fatalf("flat refresh appended events: got %d want %d", got, eventCount)
	}
	if issuer.calls != 0 || submit.places != 0 || submit.cancels != 0 || state.Pending() {
		t.Fatalf("flat refresh reached an order surface: issuer=%d place=%d cancel=%d state=%+v",
			issuer.calls, submit.places, submit.cancels, state)
	}
	if prices.calls != 3 || len(prices.asked) != 3 {
		t.Fatalf("three observation cycles used %d calls / requests %+v", prices.calls, prices.asked)
	}
	for cycle, asked := range prices.asked {
		if len(asked) != 1 || asked[0] != "005930" {
			t.Fatalf("flat cycle %d was not one complete batch: %+v", cycle+1, asked)
		}
	}
}

// A quantity-only fixture cannot reach the legacy classifier hole:
// ProjectWholeShares either stays at zero (no operational projection change),
// or becomes positive, which makes Orderable and therefore legacy Changed true.
// Pending suppression is the evaluator-generated boundary D1 actually needs:
// the scalar state stays fixed while action/ratio/projection/orderability and
// suppression change, so only complete operational equality may call it flat.
func TestA111PendingSuppressionSemanticChangeUsesFullJudgementPath(t *testing.T) {
	prices := &a111BatchPrices{}
	var issuer *a111CountingIssuer
	var submit *a111SubmitSpy
	h := newExitHarness(t, func(opts *engine.ExitObserverOptions) {
		issuer = &a111CountingIssuer{delegate: opts.Issuer}
		submit = &a111SubmitSpy{delegate: opts.Submit}
		opts.Prices, opts.Issuer, opts.Submit = prices, issuer, submit
	})
	p := h.entry("005930", "10", "70000", "68000", "70000")
	a111SeedRatchet(t, h, p)
	prices.batches = [][]domain.Quote{{{Symbol: "005930", Last: 72000, FetchedAt: h.clk.Now()}}}
	if cycle := h.observe(); cycle.Err != nil || cycle.Proposed != 1 {
		t.Fatalf("initial breach proposal: %+v", cycle)
	}
	beforeState := h.state(p.ID)
	before := a111Snapshot(t, beforeState).Line
	if before.Action != exitpolicy.ActionRatchetPartial || !before.Orderable ||
		before.ProjectedQuantity != "4" || !beforeState.Pending() ||
		issuer.calls != 1 || submit.places != 1 || submit.cancels != 0 {
		t.Fatalf("fixture did not arm exactly one durable breach: line=%+v state=%+v issuer=%d place=%d cancel=%d",
			before, beforeState, issuer.calls, submit.places, submit.cancels)
	}
	eventsBefore := len(a111Events(t, h.journal, p.ID))
	pendingIntent := beforeState.PendingIntentID
	h.clk.Advance(time.Second)
	prices.batches = append(prices.batches,
		[]domain.Quote{{Symbol: "005930", Last: 72000, FetchedAt: h.clk.Now()}})
	if cycle := h.observe(); cycle.Err != nil {
		t.Fatalf("pending suppression re-evaluation: %v", cycle.Err)
	}
	afterState := h.state(p.ID)
	after := a111Snapshot(t, afterState).Line
	if after.HighWater != before.HighWater || after.CurrentProtection != before.CurrentProtection ||
		after.RatchetLevel != before.RatchetLevel {
		t.Fatalf("fixture changed a legacy scalar instead of only operational semantics: before=%+v after=%+v",
			before, after)
	}
	if after.Action != exitpolicy.ActionNone || after.Orderable ||
		after.ProjectedQuantity != "0" || after.Suppressed != exitpolicy.SuppressedPending {
		t.Fatalf("pending proposal did not produce the evaluator's non-scalar suppression line: %+v", after)
	}
	if got := len(a111Events(t, h.journal, p.ID)); got != eventsBefore+1 {
		t.Fatalf("operational difference used the scalar flat early return: events=%d want %d", got, eventsBefore+1)
	}
	last := a111Events(t, h.journal, p.ID)[eventsBefore]
	if last.Evaluation.Effective.Snapshot == nil ||
		last.Evaluation.Effective.Snapshot.Line.SnapshotID != after.SnapshotID ||
		last.Evaluation.Effective.Snapshot.RecoveryPolicy.Ratchet == nil {
		t.Fatalf("projection change lost full-record recovery evidence: %+v", last.Evaluation.Effective)
	}
	if !afterState.Pending() || afterState.PendingIntentID != pendingIntent ||
		last.ProposedIntentID != "" || issuer.calls != 1 || submit.places != 1 || submit.cancels != 0 {
		t.Fatalf("full semantic record disturbed the existing pending proposal: state=%+v event=%+v issuer=%d place=%d cancel=%d",
			afterState, last, issuer.calls, submit.places, submit.cancels)
	}
}

func TestA111BreachStillRecordsBeforeSubmitAndPendingStillDeduplicates(t *testing.T) {
	var h *exitHarness
	var positionID string
	var issuer *a111CountingIssuer
	var submit *a111SubmitSpy
	h = newExitHarness(t, func(opts *engine.ExitObserverOptions) {
		issuer = &a111CountingIssuer{delegate: opts.Issuer}
		submit = &a111SubmitSpy{delegate: opts.Submit}
		submit.beforePlace = func(req execgw.PlaceRequest) {
			state := h.state(positionID)
			if !state.Pending() || state.PendingIntentID != req.IntentID {
				t.Fatalf("submit ran before the proposal was durable: state=%+v req=%+v", state, req)
			}
		}
		opts.Issuer, opts.Submit = issuer, submit
	})
	p := h.entry("005930", "10", "70000", "68000", "70000")
	positionID = p.ID
	a111SeedRatchet(t, h, p)
	h.quote("005930", 67000)
	first := h.observe()
	if first.Err != nil || first.Proposed != 1 || issuer.calls != 1 || submit.places != 1 {
		t.Fatalf("breach did not preserve the full protective path: cycle=%+v issuer=%d place=%d",
			first, issuer.calls, submit.places)
	}
	if state := h.state(p.ID); !state.Pending() {
		t.Fatalf("breach proposal is not durable after submit: %+v", state)
	}
	h.observe()
	if issuer.calls != 1 || submit.places != 1 || submit.cancels != 0 {
		t.Fatalf("pending breach was proposed twice: issuer=%d place=%d cancel=%d",
			issuer.calls, submit.places, submit.cancels)
	}
}

func TestA111SupersededRejudgementStillReleasesWithoutNonprotectiveOrderSideEffects(t *testing.T) {
	var issuer *a111CountingIssuer
	var submit *a111SubmitSpy
	h := newExitHarness(t, func(opts *engine.ExitObserverOptions) {
		policy := exitpolicy.DefaultLadderPolicy()
		opts.Ladder = &policy
		issuer = &a111CountingIssuer{delegate: opts.Issuer}
		submit = &a111SubmitSpy{delegate: opts.Submit}
		opts.Issuer, opts.Submit = issuer, submit
	})
	p := h.entry("005930", "10", "70000", "68000", "70000")
	if _, err := h.journal.OpenExitState(context.Background(), journal.ExitStateSeed{
		PositionID: p.ID, PolicyKind: journal.ExitPolicyLadder,
		EntryPrice: "70000", InitialStop: "68000",
	}); err != nil {
		t.Fatalf("OpenExitState: %v", err)
	}
	h.quote("005930", 70500)
	if cycle := h.observe(); cycle.Err != nil {
		t.Fatalf("initial ladder evaluation: %v", cycle.Err)
	}
	supersededQuarantine(t, h, p)
	eventsBefore := len(a111Events(t, h.journal, p.ID))
	beforeID := a111Snapshot(t, h.state(p.ID)).Line.SnapshotID
	if cycle := h.observe(); cycle.Err != nil || cycle.Judged != 1 || cycle.Proposed != 0 {
		t.Fatalf("re-judgement: %v", cycle.Err)
	}
	if q, active, err := h.journal.ActiveExitSnapshotQuarantine(context.Background(), p.ID, p.InstanceSeq); err != nil {
		t.Fatal(err)
	} else if active {
		t.Fatalf("successful re-judgement left the quarantine active: %+v", q)
	}
	after := a111Snapshot(t, h.state(p.ID))
	if len(a111Events(t, h.journal, p.ID)) != eventsBefore+1 ||
		after.Line.SnapshotID == "" || after.RecoveryPolicy.Ladder == nil ||
		after.Line.SnapshotID == beforeID {
		t.Fatalf("re-judgement did not durably select/re-record recovery evidence: before=%s after=%+v events=%d/%d",
			beforeID, after, len(a111Events(t, h.journal, p.ID)), eventsBefore+1)
	}
	if issuer.calls != 0 || submit.places != 0 || submit.cancels != 0 {
		t.Fatalf("flat re-judgement reached order adapters: issuer=%d place=%d cancel=%d",
			issuer.calls, submit.places, submit.cancels)
	}
}

func TestA111QuoteEvidenceUsesOnePostBatchClockAndNeverFallsBackFromBadOfficialTime(t *testing.T) {
	tests := []struct {
		name       string
		fetchedAt  func(time.Time) time.Time
		last       float64
		wantJudged bool
	}{
		{name: "exactly_fifteen_seconds_old", fetchedAt: func(now time.Time) time.Time { return now.Add(-execgw.QueryPriceEvidenceDuration) }, last: 70100, wantJudged: true},
		{name: "older_than_fifteen_seconds", fetchedAt: func(now time.Time) time.Time { return now.Add(-execgw.QueryPriceEvidenceDuration - time.Nanosecond) }, last: 70100},
		{name: "future_official_time", fetchedAt: func(now time.Time) time.Time { return now.Add(time.Nanosecond) }, last: 70100},
		{name: "nan", fetchedAt: func(now time.Time) time.Time { return now }, last: math.NaN()},
		{name: "positive_infinity", fetchedAt: func(now time.Time) time.Time { return now }, last: math.Inf(1)},
		{name: "missing_zero", fetchedAt: func(now time.Time) time.Time { return now }, last: 0},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			prices := &a111BatchPrices{}
			h := newExitHarness(t, func(opts *engine.ExitObserverOptions) { opts.Prices = prices })
			p := h.entry("005930", "10", "70000", "68000", "70000")
			a111SeedRatchet(t, h, p)
			before := h.state(p.ID)
			eventsBefore := len(a111Events(t, h.journal, p.ID))
			prices.batches = [][]domain.Quote{{{
				Symbol: "005930", Last: tc.last, FetchedAt: tc.fetchedAt(h.clk.Now()),
			}}}
			cycle := h.observe()
			after := h.state(p.ID)
			if tc.wantJudged {
				if cycle.Err != nil || after.SnapshotStatus != journal.SnapshotStatusEvaluated {
					t.Fatalf("inclusive boundary was not accepted: cycle=%+v state=%+v", cycle, after)
				}
				return
			}
			if after.SnapshotStatus != before.SnapshotStatus ||
				len(a111Events(t, h.journal, p.ID)) != eventsBefore {
				t.Fatalf("invalid quote advanced durable evidence: cycle=%+v before=%+v after=%+v", cycle, before, after)
			}
		})
	}

	t.Run("post_batch_clock_is_shared", func(t *testing.T) {
		prices := &a111BatchPrices{}
		h := newExitHarness(t, func(opts *engine.ExitObserverOptions) { opts.Prices = prices })
		first := h.entry("000001", "10", "70000", "68000", "70000")
		second := h.entry("999999", "10", "70000", "68000", "70000")
		a111SeedRatchet(t, h, first)
		a111SeedRatchet(t, h, second)
		fetched := h.clk.Now().Add(-14 * time.Second)
		prices.beforeReturn = func(int) { h.clk.Advance(2 * time.Second) }
		prices.batches = [][]domain.Quote{{
			{Symbol: "000001", Last: 70100, FetchedAt: fetched},
			{Symbol: "999999", Last: 70100, FetchedAt: fetched},
		}}
		cycle := h.observe()
		if cycle.Judged != 0 || h.state(first.ID).SnapshotStatus != journal.SnapshotStatusSeed ||
			h.state(second.ID).SnapshotStatus != journal.SnapshotStatusSeed {
			t.Fatalf("quotes stale at the one post-batch clock were judged: %+v", cycle)
		}
	})
}

func TestA111InvalidQuoteNeverRefreshesAnEvaluatedSnapshot(t *testing.T) {
	tests := []struct {
		name    string
		invalid func(time.Time) []domain.Quote
	}{
		{name: "source_stale", invalid: func(now time.Time) []domain.Quote {
			return []domain.Quote{{Symbol: "005930", Last: 70200, FetchedAt: now.Add(-execgw.QueryPriceEvidenceDuration - time.Nanosecond)}}
		}},
		{name: "future", invalid: func(now time.Time) []domain.Quote {
			return []domain.Quote{{Symbol: "005930", Last: 70200, FetchedAt: now.Add(time.Nanosecond)}}
		}},
		{name: "nan", invalid: func(now time.Time) []domain.Quote {
			return []domain.Quote{{Symbol: "005930", Last: math.NaN(), FetchedAt: now}}
		}},
		{name: "infinity", invalid: func(now time.Time) []domain.Quote {
			return []domain.Quote{{Symbol: "005930", Last: math.Inf(1), FetchedAt: now}}
		}},
		{name: "zero", invalid: func(now time.Time) []domain.Quote {
			return []domain.Quote{{Symbol: "005930", Last: 0, FetchedAt: now}}
		}},
		{name: "managed_symbol_missing", invalid: func(now time.Time) []domain.Quote {
			return []domain.Quote{{Symbol: "UNRELATED", Last: 100, FetchedAt: now}}
		}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			prices := &a111BatchPrices{}
			h := newExitHarness(t, func(opts *engine.ExitObserverOptions) { opts.Prices = prices })
			p := h.entry("005930", "10", "70000", "68000", "70000")
			a111SeedRatchet(t, h, p)
			prices.batches = [][]domain.Quote{{{
				Symbol: "005930", Last: 70100, FetchedAt: h.clk.Now(),
			}}}
			if cycle := h.observe(); cycle.Err != nil {
				t.Fatalf("initial valid evaluation: %v", cycle.Err)
			}
			before := *a111Snapshot(t, h.state(p.ID))
			eventsBefore := len(a111Events(t, h.journal, p.ID))
			h.clk.Advance(time.Second)
			prices.batches = append(prices.batches, tc.invalid(h.clk.Now()))
			_ = h.observe()
			after := a111Snapshot(t, h.state(p.ID))
			if after.ObservedAt != before.ObservedAt ||
				after.Line.ObservationID != before.Line.ObservationID ||
				after.OutputDigest != before.OutputDigest ||
				len(a111Events(t, h.journal, p.ID)) != eventsBefore {
				t.Fatalf("invalid evidence refreshed/mutated the evaluated snapshot: before=%+v after=%+v",
					before, after)
			}
		})
	}
}

func TestA111AllInvalidHTTPBatchIsPermanentNonRetryableAndDoesNotOpenTheGate(t *testing.T) {
	prices := &a111BatchPrices{batches: [][]domain.Quote{{
		{Symbol: "005930", Last: math.NaN()}, {Symbol: "000660", Last: math.Inf(1)},
	}}}
	var gate *execgw.EntryGate
	h := newExitHarness(t, func(opts *engine.ExitObserverOptions) {
		gate = execgw.NewEntryGate(opts.Clock, map[execgw.RequiredQuery]time.Duration{
			execgw.QueryPrice: execgw.QueryPriceEvidenceDuration,
		})
		opts.Prices = prices
		opts.Retrier = &execgw.Retrier{Clock: opts.Clock, Gate: gate,
			Policy: execgw.RetryPolicy{MaxAttempts: 3, Budget: time.Minute}}
		opts.OutageAfter = 10 * time.Second
	})
	p := h.entry("005930", "10", "70000", "68000", "70000")
	a111SeedRatchet(t, h, p)
	cycle := h.observe()
	if cycle.Err == nil || execgw.ClassifyQueryError(cycle.Err) != execgw.ClassPermanent {
		t.Fatalf("all-invalid evidence error = %v class=%s, want typed permanent",
			cycle.Err, execgw.ClassifyQueryError(cycle.Err))
	}
	if prices.calls != 1 {
		t.Fatalf("all-invalid evidence was retried: price calls=%d", prices.calls)
	}
	if blocked := gate.CheckEntry(); blocked == nil || blocked.Reason != execgw.ReasonQueryStale {
		t.Fatalf("all-invalid evidence falsely stamped QueryPrice success: %v", blocked)
	}
	if state := h.state(p.ID); state.SnapshotStatus != journal.SnapshotStatusSeed {
		t.Fatalf("all-invalid evidence mutated the exit state: %+v", state)
	}
	h.clk.Advance(11 * time.Second)
	second := h.observe()
	if second.Err == nil || execgw.ClassifyQueryError(second.Err) != execgw.ClassPermanent ||
		!second.Escalated {
		t.Fatalf("all-invalid evidence reset the outage clock: %+v", second)
	}
	if prices.calls != 2 {
		t.Fatalf("two invalid cycles made %d calls, want one non-retried call each", prices.calls)
	}
}

func TestA111MissingManagedSymbolIsInvalidEvidenceAndDoesNotResetTheOutage(t *testing.T) {
	prices := &a111BatchPrices{batches: [][]domain.Quote{{
		{Symbol: "UNRELATED", Last: 100, FetchedAt: exitNow},
	}}}
	h := newExitHarness(t, func(opts *engine.ExitObserverOptions) {
		opts.Prices = prices
		opts.OutageAfter = 10 * time.Second
	})
	p := h.entry("005930", "10", "70000", "68000", "70000")
	a111SeedRatchet(t, h, p)
	first := h.observe()
	if first.Err == nil || execgw.ClassifyQueryError(first.Err) != execgw.ClassPermanent {
		t.Fatalf("missing managed symbol = %+v, want permanent evidence error", first)
	}
	h.clk.Advance(11 * time.Second)
	second := h.observe()
	if second.Err == nil || !second.Escalated {
		t.Fatalf("invalid evidence reset the outage clock instead of escalating: %+v", second)
	}
	if h.state(p.ID).SnapshotStatus != journal.SnapshotStatusSeed {
		t.Fatalf("missing quote mutated the state: %+v", h.state(p.ID))
	}
}

func TestA111ValidSiblingIsJudgedWithoutLendingFreshnessToInvalidSymbol(t *testing.T) {
	prices := &a111BatchPrices{}
	h := newExitHarness(t, func(opts *engine.ExitObserverOptions) { opts.Prices = prices })
	valid := h.entry("000001", "10", "70000", "68000", "70000")
	invalid := h.entry("999999", "10", "70000", "68000", "70000")
	a111SeedRatchet(t, h, valid)
	a111SeedRatchet(t, h, invalid)
	prices.batches = [][]domain.Quote{{
		{Symbol: "000001", Last: 70100, FetchedAt: h.clk.Now()},
		{Symbol: "999999", Last: math.NaN(), FetchedAt: h.clk.Now()},
	}}
	cycle := h.observe()
	if cycle.Err != nil || cycle.Judged != 1 || prices.calls != 1 {
		t.Fatalf("valid sibling cycle = %+v, calls=%d", cycle, prices.calls)
	}
	if h.state(valid.ID).SnapshotStatus != journal.SnapshotStatusEvaluated ||
		h.state(invalid.ID).SnapshotStatus != journal.SnapshotStatusSeed {
		t.Fatalf("valid/invalid sibling isolation failed: valid=%+v invalid=%+v",
			h.state(valid.ID), h.state(invalid.ID))
	}
	if blocked := h.gate.CheckEntry(); blocked != nil {
		t.Fatalf("one valid sibling did not establish the successful account batch: %v", blocked)
	}
}

func TestA111FallbackObservationSourceIsDurableAndRestartMonotone(t *testing.T) {
	prices := &a111BatchPrices{batches: [][]domain.Quote{{
		{Symbol: "005930", Last: 70100}, {Symbol: "999999", Last: 70100},
	}}}
	h := newExitHarness(t, func(opts *engine.ExitObserverOptions) { opts.Prices = prices })
	p := h.entry("005930", "10", "70000", "68000", "70000")
	other := h.entry("999999", "10", "70000", "68000", "70000")
	a111SeedRatchet(t, h, p)
	a111SeedRatchet(t, h, other)
	if cycle := h.observe(); cycle.Err != nil {
		t.Fatalf("first fallback observation: %v", cycle.Err)
	}
	if prices.calls != 1 {
		t.Fatalf("first fallback cycle added a broker read: calls=%d", prices.calls)
	}
	first := a111Snapshot(t, h.state(p.ID))
	if first.ObservationSource != "cycle:1" {
		t.Fatalf("fallback source = %q, want cycle:1", first.ObservationSource)
	}
	seeded := *a111Snapshot(t, h.state(other.ID))
	seeded.ObservationSource = "cycle:41"
	encoded, err := json.Marshal(seeded)
	if err != nil {
		t.Fatalf("marshal seeded maximum: %v", err)
	}
	if _, err := openRaw(t, h).ExecContext(context.Background(), `
		UPDATE exit_states
		   SET last_observation_source=?, effective_snapshot_json=?
		 WHERE position_id=?`, seeded.ObservationSource, string(encoded), other.ID); err != nil {
		t.Fatalf("seeding persisted maximum fallback sequence: %v", err)
	}

	prices2 := &a111BatchPrices{batches: [][]domain.Quote{{{Symbol: "005930", Last: 70050}}}}
	gate := execgw.NewEntryGate(h.clk, map[execgw.RequiredQuery]time.Duration{
		execgw.QueryPrice: execgw.QueryPriceEvidenceDuration,
	})
	observer, _ := a111ObserverForJournal(t, h.journal, exitAccount, h.clk, prices2, gate, &fakeSubmitter{})
	if cycle := observer.ObserveOnce(context.Background()); cycle.Err != nil {
		t.Fatalf("restart fallback observation: %v", cycle.Err)
	}
	if prices2.calls != 1 {
		t.Fatalf("restart max recovery added a broker read: calls=%d", prices2.calls)
	}
	after := a111Snapshot(t, h.state(p.ID))
	if after.ObservationSource != "cycle:42" || after.Line.ObservationID == first.Line.ObservationID {
		t.Fatalf("restart did not recover/advance fallback sequence: before=%+v after=%+v", first, after)
	}
}

func TestA111SlowFirstPositionExpiresLaterQuoteWithoutAbandoningStartedProtection(t *testing.T) {
	prices := &a111BatchPrices{}
	var issuer *a111CountingIssuer
	var submit *a111SubmitSpy
	h := newExitHarness(t, func(opts *engine.ExitObserverOptions) {
		base := opts.Clock
		issuer = &a111CountingIssuer{delegate: opts.Issuer}
		submit = &a111SubmitSpy{delegate: opts.Submit}
		submit.beforePlace = func(execgw.PlaceRequest) {
			if fake, ok := base.(interface{ Advance(time.Duration) }); ok {
				fake.Advance(16 * time.Second)
			}
		}
		opts.Prices, opts.Issuer, opts.Submit = prices, issuer, submit
	})
	first := h.entry("000001", "10", "70000", "68000", "70000")
	later := h.entry("005930", "10", "70000", "68000", "70000")
	a111SeedRatchet(t, h, first)
	a111SeedRatchet(t, h, later)
	h.workingEntry("005930", "1", "69000")
	stamp := h.clk.Now()
	prices.batches = [][]domain.Quote{{
		{Symbol: "000001", Last: 67000, FetchedAt: stamp},
		{Symbol: "005930", Last: 67000, FetchedAt: stamp},
	}}
	laterEvents := len(a111Events(t, h.journal, later.ID))
	cycle := h.observe()
	if cycle.Err != nil || cycle.Proposed != 1 || submit.places != 1 || issuer.calls != 1 {
		t.Fatalf("started protective decision did not finish: cycle=%+v issuer=%d/%v places=%d",
			cycle, issuer.calls, issuer.symbols, submit.places)
	}
	if got := len(a111Events(t, h.journal, later.ID)); got != laterEvents ||
		h.state(later.ID).SnapshotStatus != journal.SnapshotStatusSeed {
		t.Fatalf("later expired candidate caused durable effects: events=%d state=%+v",
			got, h.state(later.ID))
	}
	if prices.calls != 1 {
		t.Fatalf("use-time validation added a broker call: %d", prices.calls)
	}
	if submit.cancels != 0 || len(issuer.symbols) != 1 || issuer.symbols[0] != "000001" {
		t.Fatalf("expired later breach reached clear/issuer: cancels=%d issuerSymbols=%v",
			submit.cancels, issuer.symbols)
	}
}

func TestA111QuoteUseLeaseCannotBeExtendedByWallClockRollback(t *testing.T) {
	t.Run("injected_clock_uses_deterministic_lease_fallback", func(t *testing.T) {
		base := clock.NewFake(time.Date(2026, 8, 14, 1, 2, 3, 0, time.UTC))
		split := &a111WallElapsedClock{base: base}
		anchor := clock.LeaseAnchor(split)
		split.advanceElapsed(7 * time.Second)
		if got := clock.LeaseElapsed(split, anchor); got != 7*time.Second {
			t.Fatalf("injected lease fallback elapsed=%s, want 7s", got)
		}
	})

	newFixture := func(t *testing.T) (*exitHarness, *a111BatchPrices, *a111WallElapsedClock,
		*a111CountingIssuer, *a111SubmitSpy, journal.Position,
	) {
		t.Helper()
		prices := &a111BatchPrices{}
		var split *a111WallElapsedClock
		var issuer *a111CountingIssuer
		var submit *a111SubmitSpy
		h := newExitHarness(t, func(opts *engine.ExitObserverOptions) {
			split = &a111WallElapsedClock{base: opts.Clock.(*clock.Fake)}
			issuer = &a111CountingIssuer{delegate: opts.Issuer}
			submit = &a111SubmitSpy{delegate: opts.Submit}
			opts.Clock, opts.Prices, opts.Issuer, opts.Submit = split, prices, issuer, submit
		})
		first := h.entry("000001", "10", "70000", "68000", "70000")
		later := h.entry("005930", "10", "70000", "68000", "70000")
		a111SeedRatchet(t, h, first)
		a111SeedRatchet(t, h, later)
		stamp := split.Now()
		prices.batches = [][]domain.Quote{{
			{Symbol: "000001", Last: 70000, FetchedAt: stamp},
			{Symbol: "005930", Last: 70000, FetchedAt: stamp},
		}}
		if cycle := h.observe(); cycle.Err != nil || cycle.Judged != 2 || cycle.Proposed != 0 {
			t.Fatalf("initial official evaluation: %+v", cycle)
		}
		h.workingEntry("005930", "1", "69000")
		return h, prices, split, issuer, submit, later
	}

	assertLaterUntouched := func(t *testing.T, h *exitHarness, prices *a111BatchPrices,
		issuer *a111CountingIssuer, submit *a111SubmitSpy, later journal.Position,
		beforeState []byte, beforeEvents int, cycle engine.ExitCycle,
	) {
		t.Helper()
		afterState, err := json.Marshal(h.state(later.ID))
		if err != nil {
			t.Fatalf("marshal later state: %v", err)
		}
		if cycle.Err != nil || cycle.Judged != 1 || cycle.Proposed != 1 {
			t.Errorf("expired later quote changed cycle accounting: %+v", cycle)
		}
		if !bytes.Equal(afterState, beforeState) || len(a111Events(t, h.journal, later.ID)) != beforeEvents {
			t.Errorf("expired later quote judged/recorded/refreshed/cleared state: before=%s after=%s events=%d/%d",
				beforeState, afterState, beforeEvents, len(a111Events(t, h.journal, later.ID)))
		}
		if issuer.calls != 1 || len(issuer.symbols) != 1 || issuer.symbols[0] != "000001" ||
			submit.places != 1 || submit.cancels != 0 {
			t.Errorf("expired later quote reached order side effects: issuer=%d/%v place=%d cancel=%d",
				issuer.calls, issuer.symbols, submit.places, submit.cancels)
		}
		if prices.calls != 2 {
			t.Errorf("use-time lease check added a broker read: calls=%d", prices.calls)
		}
	}

	t.Run("monotone_elapsed_expiry_wins_while_wall_is_inside_absolute_lease", func(t *testing.T) {
		h, prices, split, issuer, submit, later := newFixture(t)
		stamp := split.Now()
		beforeState, err := json.Marshal(h.state(later.ID))
		if err != nil {
			t.Fatalf("marshal later state: %v", err)
		}
		beforeEvents := len(a111Events(t, h.journal, later.ID))
		submit.beforePlace = func(req execgw.PlaceRequest) {
			if req.Intent.Symbol == "000001" {
				split.advanceElapsed(execgw.QueryPriceEvidenceDuration + time.Nanosecond)
			}
		}
		prices.batches = append(prices.batches, []domain.Quote{
			{Symbol: "000001", Last: 67000, FetchedAt: stamp},
			{Symbol: "005930", Last: 67000, FetchedAt: stamp},
		})
		cycle := h.observe()
		if wall := split.Now(); wall.Before(stamp) || !wall.Before(stamp.Add(execgw.QueryPriceEvidenceDuration)) {
			t.Fatalf("fixture wall escaped the source/absolute lease bounds: source=%s wall=%s", stamp, wall)
		}
		if elapsed := split.Since(stamp); elapsed <= execgw.QueryPriceEvidenceDuration {
			t.Fatalf("fixture monotone lease did not expire: elapsed=%s", elapsed)
		}
		assertLaterUntouched(t, h, prices, issuer, submit, later, beforeState, beforeEvents, cycle)
	})

	t.Run("wall_before_official_source_is_rejected_before_elapsed_expiry", func(t *testing.T) {
		h, prices, split, issuer, submit, later := newFixture(t)
		stamp := split.Now()
		beforeState, err := json.Marshal(h.state(later.ID))
		if err != nil {
			t.Fatalf("marshal later state: %v", err)
		}
		beforeEvents := len(a111Events(t, h.journal, later.ID))
		submit.beforePlace = func(req execgw.PlaceRequest) {
			if req.Intent.Symbol == "000001" {
				split.advanceElapsed(time.Second)
				split.rollbackWall(2 * time.Second)
			}
		}
		prices.batches = append(prices.batches, []domain.Quote{
			{Symbol: "000001", Last: 67000, FetchedAt: stamp},
			{Symbol: "005930", Last: 67000, FetchedAt: stamp},
		})
		cycle := h.observe()
		if wall := split.Now(); !wall.Before(stamp) {
			t.Fatalf("fixture wall did not roll behind official source: source=%s wall=%s", stamp, wall)
		}
		if elapsed := split.Since(stamp); elapsed >= execgw.QueryPriceEvidenceDuration {
			t.Fatalf("fixture elapsed lease unexpectedly expired: elapsed=%s", elapsed)
		}
		assertLaterUntouched(t, h, prices, issuer, submit, later, beforeState, beforeEvents, cycle)
	})
}

func TestA111ObserverUsesClockLeaseHelpersForTheUseLease(t *testing.T) {
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller did not locate the owned test")
	}
	production := filepath.Join(filepath.Dir(thisFile), "exitloop.go")
	fset := token.NewFileSet()
	parsed, err := parser.ParseFile(fset, production, nil, 0)
	if err != nil {
		t.Fatalf("parse %s: %v", production, err)
	}
	render := func(node ast.Node) string {
		var out bytes.Buffer
		if err := format.Node(&out, fset, node); err != nil {
			t.Fatalf("format engine AST: %v", err)
		}
		return out.String()
	}
	var observe, quoteUsable *ast.FuncDecl
	for _, declaration := range parsed.Decls {
		fn, ok := declaration.(*ast.FuncDecl)
		if !ok || fn.Recv == nil {
			continue
		}
		switch fn.Name.Name {
		case "observe":
			observe = fn
		case "quoteUsable":
			quoteUsable = fn
		}
	}
	if observe == nil || quoteUsable == nil {
		t.Fatalf("lease functions not found: observe=%v quoteUsable=%v", observe != nil, quoteUsable != nil)
	}

	anchorAssignments, anchorFields := 0, 0
	ast.Inspect(observe.Body, func(node ast.Node) bool {
		switch typed := node.(type) {
		case *ast.AssignStmt:
			if render(typed) == "leaseAnchor := clock.LeaseAnchor(o.clk)" {
				anchorAssignments++
			}
		case *ast.KeyValueExpr:
			if render(typed.Key) == "leaseAnchor" && render(typed.Value) == "leaseAnchor" {
				anchorFields++
			}
		}
		return true
	})
	if anchorAssignments != 1 || anchorFields != 1 {
		t.Fatalf("observe lease anchor wiring=assignment:%d observedQuote-field:%d, want 1/1",
			anchorAssignments, anchorFields)
	}

	elapsedBounds, rawSinceCalls := 0, 0
	ast.Inspect(quoteUsable.Body, func(node ast.Node) bool {
		switch typed := node.(type) {
		case *ast.BinaryExpr:
			if typed.Op == token.LEQ &&
				render(typed.X) == "clock.LeaseElapsed(o.clk, quote.leaseAnchor)" &&
				render(typed.Y) == "execgw.QueryPriceEvidenceDuration" {
				elapsedBounds++
			}
		case *ast.CallExpr:
			if render(typed.Fun) == "o.clk.Since" {
				rawSinceCalls++
			}
		}
		return true
	})
	if elapsedBounds != 1 || rawSinceCalls != 0 {
		t.Fatalf("use lease wiring=LeaseElapsed-bound:%d raw-Clock.Since:%d, want 1/0",
			elapsedBounds, rawSinceCalls)
	}
}

func TestA111LeaseIsRecheckedAtTheRecordOrRefreshBoundary(t *testing.T) {
	t.Run("semantic_record", func(t *testing.T) {
		prices := &a111BatchPrices{}
		var boundary *a111RecordBoundaryClock
		h := newExitHarness(t, func(opts *engine.ExitObserverOptions) {
			boundary = &a111RecordBoundaryClock{base: opts.Clock.(*clock.Fake)}
			opts.Clock = boundary
			opts.Prices = prices
		})
		p := h.entry("005930", "10", "70000", "68000", "70000")
		a111SeedRatchet(t, h, p)
		prices.batches = [][]domain.Quote{{{
			Symbol: "005930", Last: 70100, FetchedAt: h.clk.Now(),
		}}}
		eventsBefore := len(a111Events(t, h.journal, p.ID))
		_ = h.observe()
		if !boundary.advanced {
			t.Fatal("observer never re-read its clock at the judgement/record boundary")
		}
		if len(a111Events(t, h.journal, p.ID)) != eventsBefore ||
			h.state(p.ID).SnapshotStatus != journal.SnapshotStatusSeed {
			t.Fatalf("lease expired at the record boundary but the judgement committed: events=%d/%d state=%+v",
				len(a111Events(t, h.journal, p.ID)), eventsBefore, h.state(p.ID))
		}
	})

	t.Run("flat_refresh", func(t *testing.T) {
		h := newExitHarness(t, nil)
		p := h.entry("005930", "10", "70000", "68000", "70000")
		a111SeedRatchet(t, h, p)
		h.quote("005930", 70100)
		if cycle := h.observe(); cycle.Err != nil {
			t.Fatalf("initial evaluation: %v", cycle.Err)
		}
		before := *a111Snapshot(t, h.state(p.ID))
		boundary := &a111RecordBoundaryClock{base: h.clk}
		prices := &a111BatchPrices{batches: [][]domain.Quote{{{
			Symbol: "005930", Last: 70100, FetchedAt: h.clk.Now(),
		}}}}
		gate := execgw.NewEntryGate(boundary,
			map[execgw.RequiredQuery]time.Duration{execgw.QueryPrice: execgw.QueryPriceEvidenceDuration})
		observer, _ := a111ObserverForJournal(t, h.journal, exitAccount, boundary, prices, gate, &fakeSubmitter{})
		_ = observer.ObserveOnce(context.Background())
		after := a111Snapshot(t, h.state(p.ID))
		if !boundary.advanced {
			t.Fatal("observer never re-read its clock at the judgement/refresh boundary")
		}
		if after.ObservedAt != before.ObservedAt ||
			after.Line.ObservationID != before.Line.ObservationID {
			t.Fatalf("lease expired at the refresh boundary but provenance advanced: before=%+v after=%+v",
				before, after)
		}
	})
}

func TestA111OneBatchAndFillDetectionPriorityRemainUnchanged(t *testing.T) {
	t.Run("one_complete_batch", func(t *testing.T) {
		prices := &a111BatchPrices{}
		h := newExitHarness(t, func(opts *engine.ExitObserverOptions) { opts.Prices = prices })
		for _, symbol := range []string{"999999", "000001"} {
			p := h.entry(symbol, "10", "70000", "68000", "70000")
			a111SeedRatchet(t, h, p)
		}
		prices.batches = [][]domain.Quote{{
			{Symbol: "000001", Last: 70100, FetchedAt: h.clk.Now()},
			{Symbol: "999999", Last: 70100, FetchedAt: h.clk.Now()},
		}}
		if cycle := h.observe(); cycle.Err != nil {
			t.Fatalf("ObserveOnce: %v", cycle.Err)
		}
		if prices.calls != 1 || len(prices.asked) != 1 {
			t.Fatalf("price batches = calls:%d asked:%+v", prices.calls, prices.asked)
		}
		asked := append([]string(nil), prices.asked[0]...)
		sort.Strings(asked)
		if strings.Join(asked, ",") != "000001,999999" {
			t.Fatalf("batched symbols = %v", asked)
		}
	})

	t.Run("fill_detection_defers_before_price_read", func(t *testing.T) {
		prices := &a111BatchPrices{}
		h := newExitHarness(t, func(opts *engine.ExitObserverOptions) { opts.Prices = prices })
		p := h.entry("005930", "10", "70000", "68000", "70000")
		a111SeedRatchet(t, h, p)
		h.slo.behind = true
		cycle := h.observe()
		if !cycle.Deferred || prices.calls != 0 || cycle.Judged != 0 {
			t.Fatalf("fill priority changed: cycle=%+v calls=%d", cycle, prices.calls)
		}
	})
}

func TestA111FallbackSequenceRecoveryIsLazyAndPriceEvidenceUsesTheGateDuration(t *testing.T) {
	if got := execgw.DefaultStaleness()[execgw.QueryPrice]; got != execgw.QueryPriceEvidenceDuration {
		t.Fatalf("price evidence duration = %s, gate staleness = %s", execgw.QueryPriceEvidenceDuration, got)
	}

	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller did not locate the owned test")
	}
	production := filepath.Join(filepath.Dir(thisFile), "exitloop.go")
	source, err := os.ReadFile(production)
	if err != nil {
		t.Fatalf("read %s: %v", production, err)
	}

	validateLazyRecovery := func(source []byte) error {
		fset := token.NewFileSet()
		parsed, err := parser.ParseFile(fset, production, source, 0)
		if err != nil {
			return fmt.Errorf("parse: %w", err)
		}
		render := func(node ast.Node) string {
			var out bytes.Buffer
			if err := format.Node(&out, fset, node); err != nil {
				return "<format-error: " + err.Error() + ">"
			}
			return out.String()
		}
		var observeOnce, observe *ast.FuncDecl
		for _, decl := range parsed.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Recv == nil {
				continue
			}
			switch fn.Name.Name {
			case "ObserveOnce":
				observeOnce = fn
			case "observe":
				observe = fn
			}
		}
		if observeOnce == nil || observe == nil {
			return fmt.Errorf("observer functions not found: ObserveOnce=%v observe=%v",
				observeOnce != nil, observe != nil)
		}

		countMaxCalls := func(node ast.Node) int {
			calls := 0
			ast.Inspect(node, func(node ast.Node) bool {
				selector, ok := node.(*ast.SelectorExpr)
				if ok && selector.Sel.Name == "MaxExitObservationCycle" {
					calls++
				}
				return true
			})
			return calls
		}
		if calls := countMaxCalls(observeOnce); calls != 0 {
			return fmt.Errorf("official cycles eagerly recover fallback sequence in ObserveOnce: calls=%d", calls)
		}
		if total := countMaxCalls(observe); total != 1 {
			return fmt.Errorf("MaxExitObservationCycle calls in observe=%d, want exactly 1", total)
		}

		var recoveryGuards []*ast.IfStmt
		ast.Inspect(observe.Body, func(node ast.Node) bool {
			branch, ok := node.(*ast.IfStmt)
			if ok && countMaxCalls(branch.Body) > 0 {
				recoveryGuards = append(recoveryGuards, branch)
			}
			return true
		})
		if len(recoveryGuards) != 1 {
			return fmt.Errorf("if bodies enclosing MaxExitObservationCycle=%d, want exactly 1", len(recoveryGuards))
		}
		guard, ok := recoveryGuards[0].Cond.(*ast.Ident)
		if !ok || guard.Name != "needsFallback" {
			return fmt.Errorf("fallback recovery guard=%q, want exact positive needsFallback",
				render(recoveryGuards[0].Cond))
		}

		var trueAssignments []*ast.AssignStmt
		ast.Inspect(observe.Body, func(node ast.Node) bool {
			assignment, ok := node.(*ast.AssignStmt)
			if !ok || assignment.Tok != token.ASSIGN || len(assignment.Lhs) != 1 || len(assignment.Rhs) != 1 {
				return true
			}
			left, leftOK := assignment.Lhs[0].(*ast.Ident)
			right, rightOK := assignment.Rhs[0].(*ast.Ident)
			if leftOK && rightOK && left.Name == "needsFallback" && right.Name == "true" {
				trueAssignments = append(trueAssignments, assignment)
			}
			return true
		})
		if len(trueAssignments) != 1 {
			return fmt.Errorf("needsFallback=true assignments=%d, want exactly 1", len(trueAssignments))
		}
		assignment := trueAssignments[0]

		retainedBranchFound := false
		ast.Inspect(observe.Body, func(node ast.Node) bool {
			loop, ok := node.(*ast.RangeStmt)
			if !ok || render(loop.X) != "quotes" || render(loop.Value) != "q" {
				return true
			}
			var retained, zeroFetched *ast.IfStmt
			for _, statement := range loop.Body.List {
				branch, ok := statement.(*ast.IfStmt)
				if !ok {
					continue
				}
				if branch.Init != nil && render(branch.Init) == "_, needed := wanted[symbol]" &&
					render(branch.Cond) == "!needed || q.Last <= 0 || math.IsNaN(q.Last) || math.IsInf(q.Last, 0)" &&
					len(branch.Body.List) == 1 {
					if jump, ok := branch.Body.List[0].(*ast.BranchStmt); ok && jump.Tok == token.CONTINUE {
						retained = branch
					}
				}
				if render(branch.Cond) == "q.FetchedAt.IsZero()" {
					zeroFetched = branch
				}
			}
			if retained != nil && zeroFetched != nil && retained.End() < zeroFetched.Pos() &&
				zeroFetched.Body.Pos() < assignment.Pos() && assignment.End() < zeroFetched.Body.End() {
				retainedBranchFound = true
			}
			return true
		})
		if !retainedBranchFound {
			return errors.New("needsFallback=true is not confined to the retained managed, finite/positive, zero-FetchedAt quote branch")
		}
		return nil
	}

	if err := validateLazyRecovery(source); err != nil {
		t.Fatalf("production lazy-recovery structure: %v", err)
	}

	for _, mutation := range []struct {
		name        string
		replacement string
	}{
		{name: "negated_guard", replacement: "if !needsFallback {"},
		{name: "always_true_disjunction", replacement: "if needsFallback || true {"},
		{name: "unconditional_true", replacement: "if true {"},
	} {
		t.Run("kills_"+mutation.name, func(t *testing.T) {
			mutated := bytes.Replace(source, []byte("if needsFallback {"), []byte(mutation.replacement), 1)
			if bytes.Equal(mutated, source) {
				t.Fatal("mutation target if needsFallback was not found")
			}
			if err := validateLazyRecovery(mutated); err == nil {
				t.Fatalf("mutation %q survived the lazy-recovery AST contract", mutation.replacement)
			} else {
				t.Logf("mutation rejected: %v", err)
			}
		})
	}
}

func TestA111MalformedTransportEvidenceDoesNotMutateOrRetryPermanentlyInvalidPayload(t *testing.T) {
	// domain.Quote is already parsed and therefore cannot carry malformed text.
	// The production-like seam for malformed payload evidence is the official
	// reader's typed 4xx decode/refusal error.
	prices := &a111BatchPrices{err: &official.APIError{Code: 422, Body: "malformed quote payload"}}
	h := newExitHarness(t, func(opts *engine.ExitObserverOptions) {
		opts.Prices = prices
		opts.Retrier = &execgw.Retrier{Clock: opts.Clock, Gate: opts.Retrier.Gate,
			Policy: execgw.RetryPolicy{MaxAttempts: 3, Budget: time.Minute}}
	})
	p := h.entry("005930", "10", "70000", "68000", "70000")
	a111SeedRatchet(t, h, p)
	cycle := h.observe()
	if cycle.Err == nil || !errors.As(cycle.Err, new(*official.APIError)) || prices.calls != 1 {
		t.Fatalf("malformed payload handling = cycle:%+v calls:%d", cycle, prices.calls)
	}
	if h.state(p.ID).SnapshotStatus != journal.SnapshotStatusSeed {
		t.Fatalf("malformed payload mutated the exit state: %+v", h.state(p.ID))
	}
}

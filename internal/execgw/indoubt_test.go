package execgw_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/domain"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
	"github.com/JungHoonGhae/tossinvest-cli/internal/orderintent"
	"github.com/JungHoonGhae/tossinvest-cli/internal/trading"
)

// IN_DOUBT resolution tests (harden-execution-base task 2.7).
//
// The rule that shapes every case below: automatic re-submission is forbidden,
// unconditionally, because the broker has no idempotency key. Resolution can only
// observe — find the order, prove it is absent, or admit it cannot tell.

// --- fake broker reads ------------------------------------------------------

// pagedOrders is a fake OrderPager. Pages are served in order, and every cursor
// handed out is recorded so a test can prove the whole list was walked.
type pagedOrders struct {
	mu sync.Mutex
	// pages maps a status ("OPEN"/"CLOSED") to its list of pages.
	pages map[string][][]string
	// served counts page fetches per status.
	served map[string]int
	err    error
}

func newPagedOrders() *pagedOrders {
	return &pagedOrders{
		pages:  map[string][][]string{},
		served: map[string]int{},
	}
}

func (p *pagedOrders) set(status string, pages ...[]string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.pages[strings.ToUpper(status)] = pages
}

func (p *pagedOrders) OrdersPageRaw(_ context.Context, q execgw.OrderQuery, cursor string) (execgw.OrderPage, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.err != nil {
		return execgw.OrderPage{}, p.err
	}
	status := strings.ToUpper(q.Status)
	pages := p.pages[status]
	index := 0
	if cursor != "" {
		if _, err := fmt.Sscanf(cursor, "p%d", &index); err != nil {
			return execgw.OrderPage{}, fmt.Errorf("bad cursor %q", cursor)
		}
	}
	p.served[status]++
	if index >= len(pages) {
		return execgw.OrderPage{}, nil
	}
	page := execgw.OrderPage{}
	for _, body := range pages[index] {
		page.Orders = append(page.Orders, json.RawMessage(body))
	}
	if index+1 < len(pages) {
		page.HasNext = true
		page.NextCursor = fmt.Sprintf("p%d", index+1)
	}
	return page, nil
}

func (p *pagedOrders) fetches(status string) int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.served[strings.ToUpper(status)]
}

// orderJSON builds an official-shaped order payload.
func orderJSON(id, symbol, side, status, qty, price, orderedAt string) string {
	return fmt.Sprintf(`{"orderId":%q,"symbol":%q,"side":%q,"status":%q,"quantity":%q,`+
		`"price":%q,"currency":"KRW","orderedAt":%q,"execution":{"filledQuantity":"0"}}`,
		id, symbol, side, status, qty, price, orderedAt)
}

// fakeAccount answers the balance/holding cross-check.
type fakeAccount struct {
	mu          sync.Mutex
	buyingPower float64
	holding     float64
	err         error
}

func (a *fakeAccount) BuyingPower(context.Context, string) (float64, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.buyingPower, a.err
}

func (a *fakeAccount) HoldingQuantity(context.Context, string) (float64, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.holding, a.err
}

// --- fixture ----------------------------------------------------------------

type doubtFixture struct {
	gw      *execgw.Gateway
	journal *journal.Journal
	clk     *clock.Fake
	orders  *pagedOrders
	account *fakeAccount
	gate    *execgw.EntryGate
	broker  *fakeBroker
}

// newDoubtFixture wires a gateway whose broker always answers "we cannot tell",
// which is how an attempt lands in IN_DOUBT in the first place.
func newDoubtFixture(t *testing.T) *doubtFixture {
	t.Helper()
	clk := clock.NewFake(fixedNow)
	j := openJournal(t, clk)
	broker := &fakeBroker{err: official.ErrServer}
	gate := execgw.NewEntryGate(clk, map[execgw.RequiredQuery]time.Duration{})
	gw, err := execgw.New(execgw.Options{
		Journal: j, Trading: trading.NewService(openPolicy(), broker), Clock: clk,
		AccountRef: "acct-7", Source: "test", Entry: gate,
	})
	if err != nil {
		t.Fatalf("execgw.New: %v", err)
	}
	return &doubtFixture{
		gw: gw, journal: j, clk: clk,
		orders:  newPagedOrders(),
		account: &fakeAccount{buyingPower: 1_000_000, holding: 0},
		gate:    gate, broker: broker,
	}
}

func (f *doubtFixture) resolver() *execgw.Resolver {
	cfg := execgw.DefaultResolveConfig()
	return &execgw.Resolver{
		Journal: f.journal,
		Orders:  f.orders,
		Account: f.account,
		Clock:   f.clk,
		Gate:    f.gate,
		Config:  cfg,
	}
}

// placeInDoubt drives one place into IN_DOUBT and returns its attempt id.
func (f *doubtFixture) placeInDoubt(t *testing.T, baseline *execgw.Baseline) string {
	t.Helper()
	req := placeRequest(t, f.journal, f.clk)
	req.Baseline = baseline
	out, err := f.gw.Place(context.Background(), req)
	if err == nil {
		t.Fatal("the fixture broker must fail so the attempt lands IN_DOUBT")
	}
	if out.State != journal.StateInDoubt {
		t.Fatalf("state: got %s, want IN_DOUBT (%s)", out.State, out.Detail)
	}
	return out.AttemptID
}

// resolveAsync runs the resolver while advancing the fake clock past each
// stabilisation sleep.
func resolveAsync(t *testing.T, r *execgw.Resolver, clk *clock.Fake, attemptID string) (execgw.Resolution, error) {
	t.Helper()
	type result struct {
		res execgw.Resolution
		err error
	}
	done := make(chan result, 1)
	go func() {
		res, err := r.Resolve(context.Background(), attemptID)
		done <- result{res, err}
	}()
	deadline := time.Now().Add(10 * time.Second)
	for {
		select {
		case out := <-done:
			return out.res, out.err
		case <-time.After(5 * time.Millisecond):
			if clk.Sleepers() > 0 {
				clk.Advance(r.Config.PollInterval)
			}
			if time.Now().After(deadline) {
				t.Fatal("resolution never finished")
			}
		}
	}
}

// --- tests ------------------------------------------------------------------

// TestResolveFindsOrderOnSecondPage is the spec's "제출 응답 유실 후 주문이 2페이지에
// 존재" scenario. A resolver that only looked at page one would declare the order
// absent and leave a live position untracked.
func TestResolveFindsOrderOnSecondPage(t *testing.T) {
	f := newDoubtFixture(t)
	attemptID := f.placeInDoubt(t, &execgw.Baseline{BuyingPower: 1_000_000, Holding: 0, Currency: "KRW"})

	f.orders.set("OPEN",
		[]string{
			orderJSON("O-other", "000660", "BUY", "OPEN", "2", "70000", "2026-03-30T10:30:00+09:00"),
		},
		[]string{
			orderJSON("O-target", "005930", "BUY", "OPEN", "2", "70000", "2026-03-30T10:30:00+09:00"),
		},
	)

	res, err := resolveAsync(t, f.resolver(), f.clk, attemptID)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if res.State != journal.StateConfirmed {
		t.Fatalf("state: got %s, want CONFIRMED (%s)", res.State, res.Detail)
	}
	if res.BrokerOrderID != "O-target" {
		t.Errorf("broker order id: got %q, want O-target", res.BrokerOrderID)
	}
	if f.orders.fetches("OPEN") < 2 {
		t.Errorf("OPEN pages fetched: got %d, want the pagination walked to completion", f.orders.fetches("OPEN"))
	}
	// CLOSED must be walked too: a duplicate hiding there is exactly what the
	// both-lists rule is for.
	if f.orders.fetches("CLOSED") == 0 {
		t.Error("the CLOSED list was never scanned")
	}

	rec, err := f.journal.LookupAttempt(context.Background(), attemptID)
	if err != nil {
		t.Fatalf("LookupAttempt: %v", err)
	}
	if rec.State != journal.StateConfirmed || rec.BrokerOrderID != "O-target" {
		t.Errorf("journal: state=%s brokerOrderID=%q", rec.State, rec.BrokerOrderID)
	}
}

// TestResolveFindsOrderInClosedList: a submitted order that already filled or was
// cancelled is gone from OPEN. Scanning only OPEN would call it absent.
func TestResolveFindsOrderInClosedList(t *testing.T) {
	f := newDoubtFixture(t)
	attemptID := f.placeInDoubt(t, &execgw.Baseline{BuyingPower: 1_000_000, Currency: "KRW"})
	f.orders.set("CLOSED", []string{
		orderJSON("O-filled", "005930", "BUY", "CLOSED", "2", "70000", "2026-03-30T10:30:00+09:00"),
	})

	res, err := resolveAsync(t, f.resolver(), f.clk, attemptID)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if res.State != journal.StateConfirmed || res.BrokerOrderID != "O-filled" {
		t.Fatalf("got %s/%s, want CONFIRMED/O-filled (%s)", res.State, res.BrokerOrderID, res.Detail)
	}
}

// TestSingleAbsentScanIsNotAFailure is the spec's "단발 부재 조회" scenario: one
// empty list is not proof. The resolver must keep observing.
func TestSingleAbsentScanIsNotAFailure(t *testing.T) {
	f := newDoubtFixture(t)
	attemptID := f.placeInDoubt(t, &execgw.Baseline{BuyingPower: 1_000_000, Currency: "KRW"})

	r := f.resolver()
	// The order appears only after the first scan came back empty.
	go func() {
		time.Sleep(20 * time.Millisecond)
		f.orders.set("OPEN", []string{
			orderJSON("O-late", "005930", "BUY", "OPEN", "2", "70000", "2026-03-30T10:30:00+09:00"),
		})
	}()

	res, err := resolveAsync(t, r, f.clk, attemptID)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if res.State != journal.StateConfirmed {
		t.Errorf("a late-appearing order must be found, got %s (%s)", res.State, res.Detail)
	}
}

// TestAbsenceNeedsStableObservationsAndDelta covers the positive path of proving
// absence: N consecutive empty scans, a minimum observation span, and account
// deltas that show nothing happened.
func TestAbsenceNeedsStableObservationsAndDelta(t *testing.T) {
	f := newDoubtFixture(t)
	attemptID := f.placeInDoubt(t, &execgw.Baseline{BuyingPower: 1_000_000, Holding: 0, Currency: "KRW"})

	r := f.resolver()
	res, err := resolveAsync(t, r, f.clk, attemptID)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if res.State != journal.StateFailedConfirmed {
		t.Fatalf("state: got %s, want FAILED_CONFIRMED (%s)", res.State, res.Detail)
	}
	if res.Observations < r.Config.StableObservations {
		t.Errorf("observations: got %d, want at least %d", res.Observations, r.Config.StableObservations)
	}
	if elapsed := f.clk.Now().Sub(fixedNow); elapsed < r.Config.MinObservation {
		t.Errorf("absence was declared after %s, before the %s minimum observation window",
			elapsed, r.Config.MinObservation)
	}
}

// TestAbsenceRefusedWhenHoldingsMoved: the lists say the order is not there, but
// the account says something happened. That contradiction is not a failure proof,
// it is a stop condition.
func TestAbsenceRefusedWhenHoldingsMoved(t *testing.T) {
	f := newDoubtFixture(t)
	attemptID := f.placeInDoubt(t, &execgw.Baseline{BuyingPower: 1_000_000, Holding: 0, Currency: "KRW"})
	f.account.holding = 2 // the shares arrived from somewhere

	res, err := resolveAsync(t, f.resolver(), f.clk, attemptID)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if res.State != journal.StateUnresolvedInDoubt {
		t.Fatalf("state: got %s, want UNRESOLVED_IN_DOUBT (%s)", res.State, res.Detail)
	}
	if res.Reason != execgw.ReasonUnresolvedInDoubt {
		t.Errorf("reason: got %q", res.Reason)
	}
}

// TestAbsenceRefusedWithoutBaseline: without a pre-submission snapshot there is no
// cross-check, so absence cannot be proven and the attempt must not be closed as
// failed.
func TestAbsenceRefusedWithoutBaseline(t *testing.T) {
	f := newDoubtFixture(t)
	attemptID := f.placeInDoubt(t, nil)

	res, err := resolveAsync(t, f.resolver(), f.clk, attemptID)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if res.State != journal.StateUnresolvedInDoubt {
		t.Errorf("state: got %s, want UNRESOLVED_IN_DOUBT (%s)", res.State, res.Detail)
	}
}

// TestAmbiguousMatchIsUnresolved: two orders match the fingerprint. Picking either
// would be a guess about a live account.
func TestAmbiguousMatchIsUnresolved(t *testing.T) {
	f := newDoubtFixture(t)
	attemptID := f.placeInDoubt(t, &execgw.Baseline{BuyingPower: 1_000_000, Currency: "KRW"})
	f.orders.set("OPEN", []string{
		orderJSON("O-a", "005930", "BUY", "OPEN", "2", "70000", "2026-03-30T10:30:00+09:00"),
		orderJSON("O-b", "005930", "BUY", "OPEN", "2", "70000", "2026-03-30T10:30:05+09:00"),
	})

	res, err := resolveAsync(t, f.resolver(), f.clk, attemptID)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if res.State != journal.StateUnresolvedInDoubt {
		t.Errorf("state: got %s, want UNRESOLVED_IN_DOUBT (%s)", res.State, res.Detail)
	}
}

// TestUnresolvedBlocksTheSymbolUntilAnOperatorResolves is the spec's "해소 불능"
// scenario end to end: the block is permanent, new entries are refused, and only
// an explicit operator resolution reopens the symbol.
func TestUnresolvedBlocksTheSymbolUntilAnOperatorResolves(t *testing.T) {
	f := newDoubtFixture(t)
	attemptID := f.placeInDoubt(t, nil) // no baseline → unprovable
	ctx := context.Background()

	if _, err := resolveAsync(t, f.resolver(), f.clk, attemptID); err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if blocked := f.gate.CheckEntry(); blocked == nil || blocked.Reason != execgw.ReasonUnresolvedInDoubt {
		t.Fatalf("an unresolved attempt must latch the entry gate, got %v", blocked)
	}

	// The journal is the authority: the attempt is UNRESOLVED_IN_DOUBT and Settle
	// refuses to touch it.
	rec, err := f.journal.LookupAttempt(ctx, attemptID)
	if err != nil {
		t.Fatalf("LookupAttempt: %v", err)
	}
	if rec.State != journal.StateUnresolvedInDoubt {
		t.Fatalf("journal state: got %s", rec.State)
	}

	if err := f.journal.OperatorResolve(ctx, attemptID, journal.StateFailedConfirmed,
		"operator@example", "", "checked the broker UI: no such order"); err != nil {
		t.Fatalf("OperatorResolve: %v", err)
	}
	rec, err = f.journal.LookupAttempt(ctx, attemptID)
	if err != nil {
		t.Fatalf("LookupAttempt: %v", err)
	}
	if rec.State != journal.StateFailedConfirmed {
		t.Errorf("after operator resolution: got %s, want FAILED_CONFIRMED", rec.State)
	}
}

// TestOperatorResolveRejectsNonTerminalTargets keeps the operator escape hatch
// narrow: it exists to close an attempt, not to reopen its lifecycle.
func TestOperatorResolveRejectsNonTerminalTargets(t *testing.T) {
	f := newDoubtFixture(t)
	attemptID := f.placeInDoubt(t, nil)
	ctx := context.Background()
	if _, err := resolveAsync(t, f.resolver(), f.clk, attemptID); err != nil {
		t.Fatalf("Resolve: %v", err)
	}

	for _, state := range []journal.AttemptState{
		journal.StateRecorded, journal.StateDispatchStarted, journal.StateAcked,
		journal.StateInDoubt, journal.StateNotDispatched,
	} {
		if err := f.journal.OperatorResolve(ctx, attemptID, state, "op", "", "note"); err == nil {
			t.Errorf("OperatorResolve to %s must be refused", state)
		}
	}
}

// TestResolverNeverResubmits is the absolute rule. The resolver has no trading
// service, no broker mutator and no submit path — this test states it as a
// property of the type, so adding one later fails here first.
func TestResolverNeverResubmits(t *testing.T) {
	f := newDoubtFixture(t)
	attemptID := f.placeInDoubt(t, &execgw.Baseline{BuyingPower: 1_000_000, Currency: "KRW"})
	before, _, _ := f.broker.totals()

	if _, err := resolveAsync(t, f.resolver(), f.clk, attemptID); err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	after, cancels, amends := f.broker.totals()
	if after != before || cancels != 0 || amends != 0 {
		t.Errorf("resolution performed a mutation: places %d→%d, cancels %d, amends %d",
			before, after, cancels, amends)
	}
}

// TestResolveIsIdempotent: a second pass over a settled attempt must not
// re-transition it (the recovery loop runs this on every restart).
func TestResolveIsIdempotent(t *testing.T) {
	f := newDoubtFixture(t)
	attemptID := f.placeInDoubt(t, &execgw.Baseline{BuyingPower: 1_000_000, Currency: "KRW"})
	f.orders.set("OPEN", []string{
		orderJSON("O-1", "005930", "BUY", "OPEN", "2", "70000", "2026-03-30T10:30:00+09:00"),
	})

	first, err := resolveAsync(t, f.resolver(), f.clk, attemptID)
	if err != nil {
		t.Fatalf("first Resolve: %v", err)
	}
	second, err := f.resolver().Resolve(context.Background(), attemptID)
	if err != nil {
		t.Fatalf("second Resolve: %v", err)
	}
	if second.State != first.State || second.BrokerOrderID != first.BrokerOrderID {
		t.Errorf("second pass changed the outcome: %+v → %+v", first, second)
	}
}

// --- one in-flight mutation per symbol --------------------------------------

// TestOneInFlightMutationPerSymbol is what makes fingerprint matching sound: with
// at most one outstanding mutation per symbol, a fingerprint match is unique by
// construction rather than by luck.
func TestOneInFlightMutationPerSymbol(t *testing.T) {
	f := newDoubtFixture(t)
	if _, err := f.gw.Place(context.Background(), placeRequest(t, f.journal, f.clk)); err == nil {
		t.Fatal("the fixture broker must fail")
	}
	// The first attempt is now IN_DOUBT and unresolved, so the symbol is busy.

	// A second, independently issued decision for the same order: the refusal
	// under test is the symbol latch, not the one-shot nonce.
	intent := placeIntent()
	second := execgw.PlaceRequest{
		Intent:   intent,
		Decision: entryDecision(t, f.journal, f.clk, intent, testLimits()),
	}
	_, err := f.gw.Place(context.Background(), second)
	var rejected *execgw.RejectedError
	if !errors.As(err, &rejected) || rejected.Reason != execgw.ReasonSymbolInFlight {
		t.Fatalf("want symbol_mutation_in_flight, got %v", err)
	}

	// A different symbol is unaffected.
	other := placeIntent()
	other.Symbol = "000660"
	otherReq := execgw.PlaceRequest{Intent: other,
		Decision: entryDecision(t, f.journal, f.clk, other, testLimits())}
	if _, err := f.gw.Place(context.Background(), otherReq); err == nil {
		t.Fatal("expected the fixture broker failure")
	} else if errors.As(err, &rejected) && rejected.Reason == execgw.ReasonSymbolInFlight {
		t.Error("a different symbol must not be blocked by the first symbol's in-flight mutation")
	}
}

// TestConcurrentMutationsOnOneSymbolSerialise closes the check-then-act race: two
// goroutines submitting the same symbol at the same instant must not both get
// through the in-flight check.
func TestConcurrentMutationsOnOneSymbolSerialise(t *testing.T) {
	release := make(chan struct{})
	entered := make(chan struct{}, 1)
	broker := &blockingBroker{enter: entered, release: release,
		result: domain.MutationResult{OrderID: "O-1"}}

	clk := clock.NewFake(fixedNow)
	j := openJournal(t, clk)
	gw, err := execgw.New(execgw.Options{
		Journal: j, Trading: trading.NewService(openPolicy(), broker), Clock: clk,
		AccountRef: "acct-7", Source: "test",
	})
	if err != nil {
		t.Fatalf("execgw.New: %v", err)
	}

	first := placeRequest(t, j, clk)
	go func() { _, _ = gw.Place(context.Background(), first) }()
	select {
	case <-entered:
	case <-time.After(2 * time.Second):
		t.Fatal("the first mutation never reached the broker")
	}

	intent := placeIntent()
	second := execgw.PlaceRequest{Intent: intent,
		Decision: entryDecision(t, j, clk, intent, testLimits())}
	_, err = gw.Place(context.Background(), second)
	close(release)

	var rejected *execgw.RejectedError
	if !errors.As(err, &rejected) || rejected.Reason != execgw.ReasonSymbolInFlight {
		t.Fatalf("a concurrent mutation on the same symbol must be refused, got %v", err)
	}
}

// blockingBroker parks inside the broker call until released, so a test can hold a
// mutation in flight.
type blockingBroker struct {
	enter   chan struct{}
	release chan struct{}
	result  domain.MutationResult
}

func (b *blockingBroker) PlacePendingOrder(context.Context, orderintent.PlaceIntent) (domain.MutationResult, error) {
	select {
	case b.enter <- struct{}{}:
	default:
	}
	<-b.release
	return b.result, nil
}

func (b *blockingBroker) CancelPendingOrder(context.Context, orderintent.CancelIntent) (domain.MutationResult, error) {
	return b.result, nil
}

func (b *blockingBroker) AmendPendingOrder(context.Context, orderintent.AmendIntent) (domain.MutationResult, error) {
	return b.result, nil
}

func (b *blockingBroker) GetOrderAvailableActions(context.Context, string) (map[string]any, error) {
	return map[string]any{}, nil
}

// TestScanOrdersWalksEveryPage covers the completion loop on its own, including
// its two defences: a repeated cursor and a runaway page count both fail closed
// instead of looping forever.
func TestScanOrdersWalksEveryPage(t *testing.T) {
	t.Run("walks to the end", func(t *testing.T) {
		p := newPagedOrders()
		p.set("OPEN",
			[]string{orderJSON("O-1", "A", "BUY", "OPEN", "1", "10", "")},
			[]string{orderJSON("O-2", "A", "BUY", "OPEN", "1", "10", "")},
			[]string{orderJSON("O-3", "A", "BUY", "OPEN", "1", "10", "")},
		)
		got, err := execgw.ScanOrders(context.Background(), p, execgw.OrderQuery{Status: "OPEN"}, 10)
		if err != nil {
			t.Fatalf("ScanOrders: %v", err)
		}
		if len(got) != 3 {
			t.Errorf("orders: got %d, want 3", len(got))
		}
	})

	t.Run("repeated cursor", func(t *testing.T) {
		_, err := execgw.ScanOrders(context.Background(), stuckPager{}, execgw.OrderQuery{Status: "OPEN"}, 10)
		if !errors.Is(err, execgw.ErrCursorLoop) {
			t.Errorf("want ErrCursorLoop, got %v", err)
		}
	})

	t.Run("page limit", func(t *testing.T) {
		_, err := execgw.ScanOrders(context.Background(), endlessPager{}, execgw.OrderQuery{Status: "OPEN"}, 3)
		if !errors.Is(err, execgw.ErrTooManyPages) {
			t.Errorf("want ErrTooManyPages, got %v", err)
		}
	})
}

// stuckPager always hands back the same cursor.
type stuckPager struct{}

func (stuckPager) OrdersPageRaw(context.Context, execgw.OrderQuery, string) (execgw.OrderPage, error) {
	return execgw.OrderPage{HasNext: true, NextCursor: "same"}, nil
}

// endlessPager hands back a fresh cursor forever.
type endlessPager struct{ n int }

func (p endlessPager) OrdersPageRaw(_ context.Context, _ execgw.OrderQuery, cursor string) (execgw.OrderPage, error) {
	return execgw.OrderPage{HasNext: true, NextCursor: cursor + "x"}, nil
}

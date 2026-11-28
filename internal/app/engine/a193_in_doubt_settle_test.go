package engine

// a193 — 엔진이 도는 동안 IN_DOUBT 발주를 해소하는 보조 실행자(in-doubt-settle)의 시험.
//
// 실 원장(임시 경로) + 실 execgw.Resolver(가짜 목록·계좌 판독기) + 가짜 시계로 해소 갈래를 돌리고, 실행자의 알림은 가짜 critical
// 기록자로 셈. 값은 전부 합성(실제 종목·수량·가격·주문번호·계좌 아님). 브로커 호출 0 — 목록·계좌는 메모리 가짜.

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
)

const (
	a193Account = "acct-a193-synthetic"
	a193Symbol  = "Z00001"
)

var a193Now = time.Date(2026, 3, 30, 14, 0, 0, 0, time.UTC)

// --- 고정물 --------------------------------------------------------------------

// a193Criticals 는 critical 기록 단일 입구의 가짜임 — 기록 순간의 ctx 상태까지 남김.
type a193Criticals struct {
	mu        sync.Mutex
	events    []obs.Event
	remind    []time.Duration
	cancelled []bool
	err       error
}

func (c *a193Criticals) RecordCritical(ctx context.Context, e obs.Event, remindAfter time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cancelled = append(c.cancelled, ctx.Err() != nil)
	if c.err != nil {
		return c.err
	}
	c.events = append(c.events, e)
	c.remind = append(c.remind, remindAfter)
	return nil
}

func (c *a193Criticals) keys() []string {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := make([]string, 0, len(c.events))
	for _, e := range c.events {
		out = append(out, e.Key)
	}
	return out
}

func (c *a193Criticals) withPrefix(prefix string) []obs.Event {
	c.mu.Lock()
	defer c.mu.Unlock()
	var out []obs.Event
	for _, e := range c.events {
		if strings.HasPrefix(e.Key, prefix) {
			out = append(out, e)
		}
	}
	return out
}

func (c *a193Criticals) calls() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.cancelled)
}

// a193Pager 는 공식 주문 목록의 가짜임 — 요청 수를 세고, 상태별로 돌려줄 주문을 고름.
type a193Pager struct {
	mu      sync.Mutex
	calls   int
	open    []json.RawMessage
	err     error
	pages   int           // 0 이면 한 쪽, n 이면 n 쪽(빈 쪽들)
	advance time.Duration // 쪽마다 가짜 시계를 이만큼 밂(느린 응답 흉내)
	clk     *clock.Fake
	block   chan struct{} // 닫히기 전까지 ctx 가 끝날 때를 기다림
}

func (p *a193Pager) OrdersPageRaw(ctx context.Context, q execgw.OrderQuery, cursor string) (execgw.OrderPage, error) {
	p.mu.Lock()
	p.calls++
	block := p.block
	p.mu.Unlock()
	if block != nil {
		select {
		case <-ctx.Done():
			return execgw.OrderPage{}, ctx.Err()
		case <-block:
		}
	}
	if p.advance > 0 && p.clk != nil {
		p.clk.Advance(p.advance)
	}
	if p.err != nil {
		return execgw.OrderPage{}, p.err
	}
	if p.pages > 1 {
		n := 0
		if cursor != "" {
			fmt.Sscanf(cursor, "c%d", &n)
		}
		if n+1 < p.pages {
			return execgw.OrderPage{HasNext: true, NextCursor: fmt.Sprintf("c%d", n+1)}, nil
		}
		return execgw.OrderPage{}, nil
	}
	if q.Status == "OPEN" {
		return execgw.OrderPage{Orders: p.open}, nil
	}
	return execgw.OrderPage{}, nil
}

func (p *a193Pager) count() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.calls
}

// a193Account 는 부재 교차 확인의 계좌 판독기 가짜임.
type a193AccountReader struct {
	holding, buyingPower float64
	block                bool
	calls                atomic.Int32
}

func (a *a193AccountReader) BuyingPower(ctx context.Context, _ string) (float64, error) {
	a.calls.Add(1)
	if a.block {
		<-ctx.Done()
		return 0, ctx.Err()
	}
	return a.buyingPower, nil
}

func (a *a193AccountReader) HoldingQuantity(ctx context.Context, _ string) (float64, error) {
	a.calls.Add(1)
	if a.block {
		<-ctx.Done()
		return 0, ctx.Err()
	}
	return a.holding, nil
}

type a193Fixture struct {
	t        *testing.T
	clk      *clock.Fake
	journal  *journal.Journal
	gate     *execgw.EntryGate
	pager    *a193Pager
	account  *a193AccountReader
	resolver *execgw.Resolver
	crit     *a193Criticals
	ids      int
}

func newA193Fixture(t *testing.T) *a193Fixture {
	t.Helper()
	clk := clock.NewFake(a193Now)
	j, err := journal.Open(context.Background(), journal.Options{
		Path:     filepath.Join(t.TempDir(), "journal.db"),
		Clock:    clk,
		FSProber: journal.FixedFSProber(journal.FSInfo{Name: "ext4", Magic: journal.MagicExt}),
	})
	if err != nil {
		t.Fatalf("journal.Open: %v", err)
	}
	t.Cleanup(func() { _ = j.Close() })
	gate := execgw.NewEntryGate(clk, map[execgw.RequiredQuery]time.Duration{})
	pager := &a193Pager{clk: clk}
	acct := &a193AccountReader{buyingPower: 1_000_000}
	return &a193Fixture{
		t: t, clk: clk, journal: j, gate: gate, pager: pager, account: acct, crit: &a193Criticals{},
		resolver: &execgw.Resolver{Journal: j, Orders: pager, Order: nil, Account: acct, Clock: clk, Gate: gate},
	}
}

// attempt 는 PLACE(또는 CANCEL) attempt 하나를 원하는 상태로 원장에 남김. baseline 이 참이면 제출 전 기준선을 intent 에 실음.
func (f *a193Fixture) attempt(side string, kind journal.MutationKind, to journal.AttemptState, baseline bool) string {
	f.t.Helper()
	ctx := context.Background()
	f.ids++
	id := fmt.Sprintf("att-a193-%d", f.ids)
	notes := ""
	if baseline {
		notes = execgw.EncodeBaseline(&execgw.Baseline{BuyingPower: 1_000_000, Holding: 0, Currency: "KRW"})
	}
	req := journal.PrepareRequest{
		Intent: journal.Intent{
			ID: fmt.Sprintf("int-a193-%d", f.ids), Market: "kr", TradingDay: "2026-03-30", AccountRef: a193Account,
			Symbol: a193Symbol, Side: side, OrderType: "LIMIT", TimeInForce: "DAY", Quantity: "10", Price: "1000",
			Currency: "KRW", Source: "engine/test", Fingerprint: fmt.Sprintf("fp-a193-%d", f.ids), Notes: notes,
		},
		Kind: kind, AttemptID: id, AccountRef: a193Account,
	}
	if kind != journal.KindPlace {
		req.TargetOrderID = "ord-synthetic-target"
	}
	att, err := f.journal.Prepare(ctx, req)
	if err != nil {
		f.t.Fatalf("Prepare: %v", err)
	}
	step := func(err error) {
		f.t.Helper()
		if err != nil {
			f.t.Fatalf("driving %s to %s: %v", id, to, err)
		}
	}
	if to == journal.StateRecorded {
		return id
	}
	step(att.MarkDispatchStarted(ctx))
	switch to {
	case journal.StateDispatchStarted:
	case journal.StateAcked:
		step(att.MarkAcked(ctx, "ord-synthetic-acked"))
	case journal.StateInDoubt:
		step(att.MarkInDoubt(ctx, "dispatch_outcome_unknown", "5xx"))
	default:
		f.t.Fatalf("unsupported target state %s", to)
	}
	return id
}

func (f *a193Fixture) state(id string) journal.AttemptRecord {
	f.t.Helper()
	rec, err := f.journal.LookupAttempt(context.Background(), id)
	if err != nil {
		f.t.Fatalf("LookupAttempt: %v", err)
	}
	return rec
}

// a193Ledger 는 원장을 감싸 무장된 exit 발의 목록과 읽기 실패를 시험이 정하게 함.
type a193Ledger struct {
	*journal.Journal
	armed      []journal.ArmedExitProposal
	armedErr   error
	pendingErr error
	pendingN   atomic.Int32
	panicky    bool
}

func (l *a193Ledger) PendingAttempts(ctx context.Context) ([]journal.AttemptRecord, error) {
	l.pendingN.Add(1)
	if l.panicky {
		panic("a193 synthetic ledger panic")
	}
	if l.pendingErr != nil {
		return nil, l.pendingErr
	}
	return l.Journal.PendingAttempts(ctx)
}

func (l *a193Ledger) ArmedExitProposals(context.Context, string) ([]journal.ArmedExitProposal, error) {
	return l.armed, l.armedErr
}

func (f *a193Fixture) settler(ledger inDoubtLedger, resolver inDoubtResolver) *inDoubtSettle {
	if ledger == nil {
		ledger = &a193Ledger{Journal: f.journal}
	}
	if resolver == nil {
		resolver = f.resolver
	}
	n := 0
	return newInDoubtSettle(inDoubtSettleDeps{
		Ledger: ledger, Resolver: resolver, Critical: f.crit, Clock: f.clk, Account: a193Account,
		ResolveBudget: execgw.DefaultResolveConfig().MaxDuration,
		NewID:         func() string { n++; return fmt.Sprintf("run-%d", n) },
	})
}

// sweepAsync 는 스윕을 goroutine 에서 돌리고, 가짜 시계를 해소의 대기마다 밀어 끝까지 가게 함.
func (f *a193Fixture) sweepDriven(s *inDoubtSettle) error {
	f.t.Helper()
	done := make(chan error, 1)
	go func() { done <- s.sweep(context.Background()) }()
	for i := 0; i < 400; i++ {
		select {
		case err := <-done:
			return err
		default:
		}
		if f.clk.WaitForSleepers(1, 20*time.Millisecond) {
			f.clk.Advance(5 * time.Second)
		}
	}
	f.t.Fatal("the sweep did not finish")
	return nil
}

func a193Order(id string, at time.Time) json.RawMessage {
	return json.RawMessage(fmt.Sprintf(`{"orderId":%q,"symbol":%q,"side":"SELL","status":"PENDING","quantity":"10","price":"1000","orderedAt":%q}`,
		id, a193Symbol, at.UTC().Format(time.RFC3339)))
}

// --- 3.1 ~ 3.3 결과 갈래 ---------------------------------------------------------

func TestA193AnInDoubtPlaceFoundInTheListIsConfirmedWhileTheEngineRuns(t *testing.T) {
	f := newA193Fixture(t)
	id := f.attempt("SELL", journal.KindPlace, journal.StateInDoubt, false)
	f.pager.open = []json.RawMessage{a193Order("ord-synthetic-1", a193Now)}

	if err := f.sweepDriven(f.settler(nil, nil)); err != nil {
		t.Fatalf("sweep: %v", err)
	}
	rec := f.state(id)
	if rec.State != journal.StateConfirmed || rec.BrokerOrderID != "ord-synthetic-1" {
		t.Fatalf("attempt = %s / %q, want CONFIRMED with the found order id", rec.State, rec.BrokerOrderID)
	}
	if refusal := f.gate.CheckEntry(); refusal != nil {
		t.Fatalf("a confirmed resolution latched the entry gate: %+v", refusal)
	}
	if got := f.crit.withPrefix(string(obs.EventOrderUnresolved)); len(got) != 0 {
		t.Fatalf("a confirmed resolution raised a park alert: %+v", got)
	}
}

func TestA193AnInDoubtSellWithoutBaselineParksAndTheParkCauseAlertsOnce(t *testing.T) {
	f := newA193Fixture(t)
	id := f.attempt("SELL", journal.KindPlace, journal.StateInDoubt, false)
	ledger := &a193Ledger{Journal: f.journal}
	rec := f.state(id)
	intent, _ := f.journal.LookupIntent(context.Background(), rec.IntentID)
	ledger.armed = []journal.ArmedExitProposal{{PositionID: "pos-a193", IntentID: intent.ID}}

	if err := f.sweepDriven(f.settler(ledger, nil)); err != nil {
		t.Fatalf("sweep: %v", err)
	}
	got := f.state(id)
	if got.State != journal.StateUnresolvedInDoubt || !strings.Contains(got.Detail, "without a pre-dispatch account baseline") {
		t.Fatalf("attempt = %s (%q), want parked for the missing baseline", got.State, got.Detail)
	}
	if refusal := f.gate.CheckEntry(); refusal == nil || refusal.Reason != execgw.ReasonUnresolvedInDoubt {
		t.Fatalf("park did not latch the account-wide entry gate: %+v", refusal)
	}
	// exit 발의의 attempt 는 기존 recordParkCause 가 알림 — 그 입력(Parked)에 이 attempt 가 있어야 하고 실행자는 알리지 않음.
	facts, err := f.journal.ExitIntentAttempts(context.Background(), intent.ID)
	if err != nil || len(facts.Parked) != 1 || facts.Parked[0].ID != id {
		t.Fatalf("ExitIntentAttempts.Parked = %+v (%v), want the parked attempt for recordParkCause", facts.Parked, err)
	}
	if got := f.crit.withPrefix(string(obs.EventOrderUnresolved)); len(got) != 0 {
		t.Fatalf("the settler duplicated the exit park alert: %+v", got)
	}
}

func TestA193AnInDoubtBuyProvenAbsentIsFailedConfirmed(t *testing.T) {
	f := newA193Fixture(t)
	id := f.attempt("BUY", journal.KindPlace, journal.StateInDoubt, true)
	if err := f.sweepDriven(f.settler(nil, nil)); err != nil {
		t.Fatalf("sweep: %v", err)
	}
	if got := f.state(id); got.State != journal.StateFailedConfirmed {
		t.Fatalf("attempt = %s (%q), want FAILED_CONFIRMED", got.State, got.Detail)
	}
}

// --- 3.4 · 3.5 취소 -----------------------------------------------------------

func TestA193ACancelledResolutionLeavesTheAttemptInDoubt(t *testing.T) {
	cases := []struct {
		name  string
		side  string
		base  bool
		setup func(f *a193Fixture)
		wait  func(f *a193Fixture)
	}{
		{"during the poll sleep", "SELL", false, func(*a193Fixture) {}, func(f *a193Fixture) {
			if !f.clk.WaitForSleepers(1, 5*time.Second) {
				t.Fatal("the resolver never slept")
			}
		}},
		{"during a list scan", "SELL", false, func(f *a193Fixture) { f.pager.block = make(chan struct{}) }, func(f *a193Fixture) {
			for i := 0; i < 500 && f.pager.count() == 0; i++ {
				time.Sleep(time.Millisecond)
			}
		}},
		{"during the absence cross-check", "BUY", true, func(f *a193Fixture) { f.account.block = true }, func(f *a193Fixture) {
			for i := 0; i < 2000 && f.account.calls.Load() == 0; i++ {
				if f.clk.WaitForSleepers(1, time.Millisecond) {
					f.clk.Advance(5 * time.Second)
				}
			}
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f := newA193Fixture(t)
			id := f.attempt(tc.side, journal.KindPlace, journal.StateInDoubt, tc.base)
			tc.setup(f)
			s := f.settler(nil, nil)
			ctx, cancel := context.WithCancel(context.Background())
			done := make(chan error, 1)
			go func() { done <- s.sweep(ctx) }()
			tc.wait(f)
			before := f.crit.calls()
			cancel()
			err := <-done
			if !errors.Is(err, context.Canceled) {
				t.Fatalf("sweep returned %v, want context.Canceled", err)
			}
			got := f.state(id)
			if got.State != journal.StateInDoubt || got.SettledAt != "" {
				t.Fatalf("a cancelled resolution wrote %s (settled %q); the attempt must stay IN_DOUBT", got.State, got.SettledAt)
			}
			if refusal := f.gate.CheckEntry(); refusal != nil {
				t.Fatalf("a cancelled resolution latched the entry gate: %+v", refusal)
			}
			if after := f.crit.calls(); after != before {
				t.Fatalf("a cancelled resolution recorded %d critical(s)", after-before)
			}
		})
	}
	t.Run("a stale attempt swept with a cancelled context records nothing", func(t *testing.T) {
		f := newA193Fixture(t)
		f.attempt("SELL", journal.KindPlace, journal.StateInDoubt, false)
		f.clk.Advance(time.Hour)
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		if err := f.settler(nil, nil).sweep(ctx); !errors.Is(err, context.Canceled) {
			t.Fatalf("sweep = %v, want context.Canceled", err)
		}
		if n := f.crit.calls(); n != 0 {
			t.Fatalf("a cancelled sweep tried %d critical record(s) — a failed record latches the gate", n)
		}
	})
}

func TestA193AShutdownMidResolutionIsAGracefulAuxiliaryStop(t *testing.T) {
	f := newA193Fixture(t)
	id := f.attempt("SELL", journal.KindPlace, journal.StateInDoubt, false)
	s := f.settler(nil, nil)
	stopped := atomic.Int32{}
	exec := s.executor()
	exec.OnStop = func(context.Context, error) { stopped.Add(1) }
	rt := a193Runtime(t, f.clk, nil, exec)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- rt.Run(ctx) }()
	// 감독 루프 하나 + 건강 감시 하나 + 해소의 대기 하나.
	if !f.clk.WaitForSleepers(3, 5*time.Second) || f.pager.count() == 0 {
		t.Fatalf("the resolution never reached its poll sleep (broker list requests %d)", f.pager.count())
	}
	cancel()
	if err := <-done; err != nil {
		t.Fatalf("runtime returned %v, want a graceful nil", err)
	}
	if stopped.Load() != 0 {
		t.Fatal("a graceful stop reached OnStop")
	}
	if got := f.state(id); got.State != journal.StateInDoubt {
		t.Fatalf("attempt = %s after shutdown, want IN_DOUBT", got.State)
	}
}

// a193Runtime 은 가짜 감독 루프 하나와 보조 실행자 하나로 된 런타임임.
func a193Runtime(t *testing.T, clk clock.Clock, recover func(context.Context) error, aux AuxiliaryExecutor) *Runtime {
	t.Helper()
	rt, err := NewRuntime(RuntimeOptions{
		Loops: []SupervisedLoop{{Name: "a193-loop", Run: func(ctx context.Context) error {
			for {
				if err := clk.Sleep(ctx, time.Hour); err != nil {
					return err
				}
			}
		}}},
		Auxiliary: []AuxiliaryExecutor{aux},
		Clock:     clk,
		Recover:   recover,
	})
	if err != nil {
		t.Fatalf("NewRuntime: %v", err)
	}
	return rt
}

// --- 3.6 비대기 · 직렬 · 한 번 ---------------------------------------------------

type a193BlockingResolver struct {
	mu       sync.Mutex
	inflight int
	max      int
	calls    map[string]int
	release  chan struct{}
	entered  chan string
	err      error
}

func newA193BlockingResolver() *a193BlockingResolver {
	return &a193BlockingResolver{calls: map[string]int{}, release: make(chan struct{}), entered: make(chan string, 16)}
}

func (r *a193BlockingResolver) Resolve(ctx context.Context, id string) (execgw.Resolution, error) {
	r.mu.Lock()
	r.inflight++
	if r.inflight > r.max {
		r.max = r.inflight
	}
	r.calls[id]++
	r.mu.Unlock()
	r.entered <- id
	defer func() { r.mu.Lock(); r.inflight--; r.mu.Unlock() }()
	select {
	case <-ctx.Done():
		return execgw.Resolution{}, ctx.Err()
	case <-r.release:
	}
	if r.err != nil {
		return execgw.Resolution{AttemptID: id}, r.err
	}
	return execgw.Resolution{AttemptID: id, State: journal.StateConfirmed, BrokerOrderID: "ord-synthetic-x"}, nil
}

func TestA193TheExitObserverDoesNotWaitForAResolution(t *testing.T) {
	// 범위: 런타임 수준 — 해소가 막혀 있는 동안 다른 감독 루프의 주기가 계속 도는가(실제 ExitObserver 의 제출 경로는 재지 않음 — review.md 구현 리뷰 1회차 P2).
	f := newA193Fixture(t)
	f.attempt("SELL", journal.KindPlace, journal.StateInDoubt, false)
	res := newA193BlockingResolver()
	cycles := make(chan struct{}, 64)
	rt, err := NewRuntime(RuntimeOptions{
		Loops: []SupervisedLoop{{Name: "exit", Run: func(ctx context.Context) error {
			for {
				select {
				case <-ctx.Done():
					return ctx.Err()
				case cycles <- struct{}{}:
				}
			}
		}}},
		Auxiliary: []AuxiliaryExecutor{f.settler(nil, res).executor()},
		Clock:     f.clk,
	})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- rt.Run(ctx) }()
	t.Cleanup(func() { cancel(); <-done })
	deadline := time.After(10 * time.Second) // 실패 시한일 뿐 — 성공은 주기 도착으로 판정(준비 단계부터 적용)
	select {
	case <-res.entered: // 해소가 막힘 — 아래 주기는 전부 그 뒤
	case <-deadline:
		t.Fatal("the resolution never started")
	}
	for n := len(cycles); n > 0; n-- { // 진입 시점에 쌓인 것만 비움
		<-cycles
	}
	for i := 0; i < 5; i++ {
		select {
		case <-cycles:
		case <-deadline:
			t.Fatalf("the loop completed only %d cycle(s) while a resolution was blocked", i)
		}
	}
}

func TestA193AttemptsAreResolvedOneAtATime(t *testing.T) {
	f := newA193Fixture(t)
	f.attempt("SELL", journal.KindPlace, journal.StateInDoubt, false)
	f.attempt("BUY", journal.KindPlace, journal.StateInDoubt, false)
	res := newA193BlockingResolver()
	s := f.settler(nil, res)
	done := make(chan error, 1)
	go func() { done <- s.sweep(context.Background()) }()
	<-res.entered
	time.Sleep(20 * time.Millisecond)
	res.mu.Lock()
	inflight := res.inflight
	res.mu.Unlock()
	if inflight != 1 {
		t.Fatalf("%d resolutions in flight, want 1", inflight)
	}
	close(res.release)
	if err := <-done; err != nil {
		t.Fatal(err)
	}
	if res.max != 1 || len(res.calls) != 2 {
		t.Fatalf("max in flight %d, resolved %v — want serial over both", res.max, res.calls)
	}
}

func TestA193TheSameAttemptIsNotResolvedTwice(t *testing.T) {
	f := newA193Fixture(t)
	id := f.attempt("SELL", journal.KindPlace, journal.StateInDoubt, false)
	f.pager.open = []json.RawMessage{a193Order("ord-synthetic-2", a193Now)}
	s := f.settler(nil, nil)
	for i := 0; i < 3; i++ {
		if err := f.sweepDriven(s); err != nil {
			t.Fatal(err)
		}
	}
	if got := f.state(id).State; got != journal.StateConfirmed {
		t.Fatalf("state = %s", got)
	}
	// 첫 스윕의 해소 한 번 = 목록 2(OPEN·CLOSED). 종결 뒤 스윕은 브로커를 부르지 않음.
	if n := f.pager.count(); n != 2 {
		t.Fatalf("broker list requests = %d over three sweeps, want 2 (one resolution)", n)
	}
}

// --- 3.7 10분 알림 ------------------------------------------------------------

func TestA193AnAttemptInDoubtPastTheBoundAlertsOnce(t *testing.T) {
	f := newA193Fixture(t)
	id := f.attempt("SELL", journal.KindPlace, journal.StateInDoubt, false)
	res := newA193BlockingResolver()
	close(res.release)
	res.err = errors.New("synthetic resolution failure")
	s := f.settler(nil, res)
	stale := string(obs.EventOrderInDoubt) + "|stale:" + id

	f.clk.Advance(10*time.Minute - time.Second)
	if err := s.sweep(context.Background()); err != nil {
		t.Fatal(err)
	}
	if n := len(f.crit.withPrefix(stale)); n != 0 {
		t.Fatalf("stale alert before the bound: %d", n)
	}
	f.clk.Advance(time.Second)
	for i := 0; i < 3; i++ {
		if err := s.sweep(context.Background()); err != nil {
			t.Fatal(err)
		}
	}
	got := f.crit.withPrefix(stale)
	if len(got) != 1 {
		t.Fatalf("stale alerts = %d, want exactly 1 at the bound (keys %v)", len(got), f.crit.keys())
	}
	for i, e := range f.crit.events {
		if e.Key == stale && f.crit.remind[i] != 0 {
			t.Fatalf("stale alert window = %v, want 0", f.crit.remind[i])
		}
	}
	// 재시작(새 실행자)은 같은 key — 원장 입구가 같은 행으로 받음.
	again := f.settler(nil, res)
	if err := again.sweep(context.Background()); err != nil {
		t.Fatal(err)
	}
	if n := len(f.crit.withPrefix(stale)); n != 2 {
		t.Fatalf("a restarted settler did not re-offer the same episode key (%d)", n)
	}
}

func TestA193StaleFallsBackToTheRecordedTime(t *testing.T) {
	f := newA193Fixture(t)
	s := f.settler(nil, nil)
	rec := journal.AttemptRecord{ID: "att-synthetic-norecord", Kind: journal.KindPlace, State: journal.StateInDoubt,
		RecordedAt: a193Now.Add(-10 * time.Minute).Format(time.RFC3339Nano)}
	s.noteStale(context.Background(), rec)
	if n := len(f.crit.withPrefix(string(obs.EventOrderInDoubt) + "|stale:" + rec.ID)); n != 1 {
		t.Fatalf("an attempt with no dispatch time was not judged from its recorded time (%d)", n)
	}
}

func TestA193AFailedStaleRecordIsRetried(t *testing.T) {
	f := newA193Fixture(t)
	id := f.attempt("SELL", journal.KindPlace, journal.StateInDoubt, false)
	res := newA193BlockingResolver()
	close(res.release)
	res.err = errors.New("synthetic")
	s := f.settler(nil, res)
	f.clk.Advance(11 * time.Minute)
	f.crit.err = errors.New("synthetic outbox failure")
	_ = s.sweep(context.Background())
	f.crit.err = nil
	_ = s.sweep(context.Background())
	if n := len(f.crit.withPrefix(string(obs.EventOrderInDoubt) + "|stale:" + id)); n != 1 {
		t.Fatalf("a stale alert whose record failed was not retried (%d)", n)
	}
}

func TestA193TheStaleAlertCarriesNoAccountQuantityOrPrice(t *testing.T) {
	f := newA193Fixture(t)
	f.attempt("SELL", journal.KindPlace, journal.StateInDoubt, false)
	res := newA193BlockingResolver()
	close(res.release)
	res.err = errors.New("synthetic")
	s := f.settler(nil, res)
	f.clk.Advance(11 * time.Minute)
	_ = s.sweep(context.Background())
	if len(f.crit.events) == 0 {
		t.Fatal("no alert to inspect")
	}
	for _, e := range f.crit.events {
		blob := e.Title + "|" + e.Body + "|" + fmt.Sprint(e.Fields)
		for _, forbidden := range []string{a193Account, "1000", "\"10\"", "quantity", "price"} {
			if strings.Contains(blob, forbidden) {
				t.Fatalf("alert %s carries %q: %s", e.Key, forbidden, blob)
			}
		}
	}
}

// --- 3.8 대상 · 재생 · 브로커 0 ------------------------------------------------

func TestA193OnlyInDoubtPlaceAttemptsAreTouched(t *testing.T) {
	f := newA193Fixture(t)
	ids := map[string]journal.AttemptState{
		f.attempt("SELL", journal.KindPlace, journal.StateRecorded, false):        journal.StateRecorded,
		f.attempt("SELL", journal.KindPlace, journal.StateDispatchStarted, false): journal.StateDispatchStarted,
		f.attempt("SELL", journal.KindPlace, journal.StateAcked, false):           journal.StateAcked,
		f.attempt("SELL", journal.KindCancel, journal.StateInDoubt, false):        journal.StateInDoubt,
	}
	res := newA193BlockingResolver()
	close(res.release)
	s := f.settler(nil, res)
	f.clk.Advance(time.Hour)
	if err := s.sweep(context.Background()); err != nil {
		t.Fatal(err)
	}
	if len(res.calls) != 0 {
		t.Fatalf("resolved %v — only IN_DOUBT PLACE attempts may be", res.calls)
	}
	for id, want := range ids {
		if got := f.state(id).State; got != want {
			t.Fatalf("%s moved %s → %s", id, want, got)
		}
	}
	if n := f.crit.calls(); n != 0 {
		t.Fatalf("%d alert(s) for attempts the settler does not own", n)
	}
}

func TestA193NoInDoubtMeansNoBrokerCall(t *testing.T) {
	f := newA193Fixture(t)
	f.attempt("SELL", journal.KindPlace, journal.StateAcked, false)
	s := f.settler(nil, nil)
	for i := 0; i < 5; i++ {
		if err := s.sweep(context.Background()); err != nil {
			t.Fatal(err)
		}
	}
	if n := f.pager.count(); n != 0 || f.account.calls.Load() != 0 {
		t.Fatalf("broker reads with nothing in doubt: lists %d, account %d", n, f.account.calls.Load())
	}
}

func TestA193TheSettlerNeverReplays(t *testing.T) {
	// 실행자 파일은 재생 진입점·게이트웨이·주문 경로를 이름으로도 부르지 않음(구조 고정 — 해소는 조회 대조만).
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "in_doubt_settle.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	ast.Inspect(file, func(n ast.Node) bool {
		if sel, ok := n.(*ast.SelectorExpr); ok {
			switch sel.Sel.Name {
			case "ReplayInDoubt", "Gateway", "Place", "Cancel", "Amend", "Submit", "RecoverPending":
				t.Errorf("in_doubt_settle.go reaches %s at %s", sel.Sel.Name, fset.Position(sel.Pos()))
			}
		}
		return true
	})
}

// --- 3.9 요청 수 실측 ---------------------------------------------------------

func TestA193ResolutionRequestsAreBounded(t *testing.T) {
	type result struct {
		name            string
		lists, accounts int
		state           journal.AttemptState
		elapsed         time.Duration
	}
	var results []result
	run := func(name, side string, base bool, setup func(f *a193Fixture)) result {
		f := newA193Fixture(t)
		id := f.attempt(side, journal.KindPlace, journal.StateInDoubt, base)
		setup(f)
		start := f.clk.Now()
		if err := f.sweepDriven(f.settler(nil, nil)); err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		r := result{name, f.pager.count(), int(f.account.calls.Load()), f.state(id).State, f.clk.Now().Sub(start)}
		results = append(results, r)
		return r
	}
	found := run("found", "SELL", false, func(f *a193Fixture) { f.pager.open = []json.RawMessage{a193Order("ord-synthetic-3", a193Now)} })
	if found.lists != 2 || found.state != journal.StateConfirmed {
		t.Fatalf("found path: %+v", found)
	}
	absent := run("absent proven", "BUY", true, func(*a193Fixture) {})
	if absent.state != journal.StateFailedConfirmed || absent.lists > 24 || absent.accounts != 2 {
		t.Fatalf("absent path: %+v", absent)
	}
	exhausted := run("scan always fails", "SELL", false, func(f *a193Fixture) { f.pager.err = errors.New("synthetic 5xx") })
	if exhausted.state != journal.StateUnresolvedInDoubt || exhausted.lists > 61 {
		t.Fatalf("exhausted path: %+v", exhausted)
	}
	slow := run("50 pages, 20s each", "SELL", false, func(f *a193Fixture) { f.pager.pages = 50; f.pager.advance = 20 * time.Second })
	if slow.elapsed <= execgw.DefaultResolveConfig().MaxDuration {
		t.Fatalf("slow path finished inside the budget (%v) — the measurement no longer shows the missing hard bound", slow.elapsed)
	}
	for _, r := range results {
		t.Logf("a193 실측 1: %-20s lists=%d account=%d state=%s fake-elapsed=%s", r.name, r.lists, r.accounts, r.state, r.elapsed)
	}
}

// --- 3.10 정지 · 기동 순서 ------------------------------------------------------

func TestA193ASettlerStopRecordsACritical(t *testing.T) {
	f := newA193Fixture(t)
	ledger := &a193Ledger{Journal: f.journal, panicky: true}
	exec := f.settler(ledger, nil).executor()
	rt := a193Runtime(t, f.clk, nil, exec)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- rt.Run(ctx) }()
	for i := 0; i < 2000 && len(f.crit.withPrefix(string(obs.EventOrderInDoubt)+"|settler-stopped:")) == 0; i++ {
		time.Sleep(time.Millisecond)
	}
	cancel()
	<-done
	if n := len(f.crit.withPrefix(string(obs.EventOrderInDoubt) + "|settler-stopped:")); n != 1 {
		t.Fatalf("a panicking settler recorded %d stop critical(s), want 1 (keys %v)", n, f.crit.keys())
	}
}

func TestA193AnUnwiredSettlerStopsLoudly(t *testing.T) {
	crit := &a193Criticals{}
	s := newInDoubtSettle(inDoubtSettleDeps{Critical: crit, Clock: clock.NewFake(a193Now), NewID: func() string { return "run-x" }})
	err := s.Run(context.Background())
	if err == nil {
		t.Fatal("an unwired settler returned nil")
	}
	s.executor().OnStop(context.Background(), err)
	if n := len(crit.withPrefix(string(obs.EventOrderInDoubt) + "|settler-stopped:")); n != 1 {
		t.Fatalf("stop critical = %d", n)
	}
}

func TestA193TheSettlerStartsAfterRecovery(t *testing.T) {
	f := newA193Fixture(t)
	ledger := &a193Ledger{Journal: f.journal}
	release := make(chan struct{})
	entered := make(chan struct{})
	rt := a193Runtime(t, f.clk, func(ctx context.Context) error {
		close(entered)
		<-release
		return nil
	}, f.settler(ledger, nil).executor())
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- rt.Run(ctx) }()
	<-entered
	time.Sleep(20 * time.Millisecond)
	if n := ledger.pendingN.Load(); n != 0 {
		t.Fatalf("the settler swept %d time(s) before recovery finished", n)
	}
	close(release)
	for i := 0; i < 2000 && ledger.pendingN.Load() == 0; i++ {
		time.Sleep(time.Millisecond)
	}
	if ledger.pendingN.Load() == 0 {
		t.Fatal("the settler never swept after recovery")
	}
	cancel()
	<-done
}

// --- 3.12 알림 수 -------------------------------------------------------------

func TestA193FirstSightIsAlertedOnce(t *testing.T) {
	f := newA193Fixture(t)
	id := f.attempt("SELL", journal.KindPlace, journal.StateInDoubt, false)
	res := newA193BlockingResolver()
	close(res.release)
	res.err = errors.New("synthetic")
	s := f.settler(nil, res)
	key := string(obs.EventOrderInDoubt) + "|attempt:" + id
	f.crit.err = errors.New("synthetic outbox failure")
	_ = s.sweep(context.Background())
	f.crit.err = nil
	for i := 0; i < 3; i++ {
		_ = s.sweep(context.Background())
	}
	got := f.crit.withPrefix(key)
	if len(got) != 1 {
		t.Fatalf("first-sight alerts = %d, want 1 after a failed record and three sweeps (keys %v)", len(got), f.crit.keys())
	}
	if got[0].Type != obs.EventOrderInDoubt {
		t.Fatalf("first-sight type = %s", got[0].Type)
	}
}

func TestA193AParkOutsideAnExitProposalIsAlertedOnce(t *testing.T) {
	f := newA193Fixture(t)
	id := f.attempt("SELL", journal.KindPlace, journal.StateInDoubt, false)
	ledger := &a193Ledger{Journal: f.journal, armed: []journal.ArmedExitProposal{{PositionID: "pos-other", IntentID: "int-other"}}}
	s := f.settler(ledger, nil)
	if err := f.sweepDriven(s); err != nil {
		t.Fatal(err)
	}
	_ = s.sweep(context.Background())
	got := f.crit.withPrefix(string(obs.EventOrderUnresolved) + "|attempt:" + id)
	if len(got) != 1 || got[0].Type != obs.EventOrderUnresolved {
		t.Fatalf("park alerts = %+v, want one order.unresolved_in_doubt keyed by the attempt", got)
	}
}

func TestA193AParkOfAnArmedExitProposalIsNotAlertedTwice(t *testing.T) {
	// 3.2 가 같은 것을 원장 사실까지 확인함 — 여기서는 키 수만.
	f := newA193Fixture(t)
	id := f.attempt("SELL", journal.KindPlace, journal.StateInDoubt, false)
	intentID := f.state(id).IntentID
	ledger := &a193Ledger{Journal: f.journal, armed: []journal.ArmedExitProposal{{PositionID: "pos-a193", IntentID: intentID}}}
	if err := f.sweepDriven(f.settler(ledger, nil)); err != nil {
		t.Fatal(err)
	}
	if n := len(f.crit.withPrefix(string(obs.EventOrderUnresolved))); n != 0 {
		t.Fatalf("the settler alerted the park of an armed exit proposal (%d)", n)
	}
}

func TestA193AnArmedProposalReadFailureStillAlerts(t *testing.T) {
	f := newA193Fixture(t)
	id := f.attempt("SELL", journal.KindPlace, journal.StateInDoubt, false)
	ledger := &a193Ledger{Journal: f.journal, armedErr: errors.New("synthetic read failure")}
	if err := f.sweepDriven(f.settler(ledger, nil)); err != nil {
		t.Fatal(err)
	}
	if n := len(f.crit.withPrefix(string(obs.EventOrderUnresolved) + "|attempt:" + id)); n != 1 {
		t.Fatalf("an unreadable armed-proposal list suppressed the park alert (%d)", n)
	}
}

func TestA193AnEpisodeRaisesFirstSightAndParkOnly(t *testing.T) {
	f := newA193Fixture(t)
	id := f.attempt("SELL", journal.KindPlace, journal.StateInDoubt, false)
	s := f.settler(nil, nil)
	for i := 0; i < 3; i++ {
		if err := f.sweepDriven(s); err != nil {
			t.Fatal(err)
		}
	}
	want := []string{string(obs.EventOrderInDoubt) + "|attempt:" + id, string(obs.EventOrderUnresolved) + "|attempt:" + id}
	got := f.crit.keys()
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("episode alerts = %v, want exactly %v", got, want)
	}
}

// --- 3.13 원장 못 읽음 · 평가 순서 -------------------------------------------

func TestA193AnUnreadableLedgerAlertsAfterTenMinutes(t *testing.T) {
	f := newA193Fixture(t)
	ledger := &a193Ledger{Journal: f.journal, pendingErr: errors.New("synthetic ledger failure")}
	s := f.settler(ledger, nil)
	prefix := string(obs.EventOrderInDoubt) + "|ledger-unreadable:"
	_ = s.sweep(context.Background())
	f.clk.Advance(10*time.Minute - time.Second)
	_ = s.sweep(context.Background())
	if n := len(f.crit.withPrefix(prefix)); n != 0 {
		t.Fatalf("unreadable alert before 10 minutes: %d", n)
	}
	f.clk.Advance(time.Second)
	_ = s.sweep(context.Background())
	_ = s.sweep(context.Background())
	first := f.crit.withPrefix(prefix)
	if len(first) != 1 {
		t.Fatalf("unreadable alerts = %d, want 1", len(first))
	}
	ledger.pendingErr = nil
	_ = s.sweep(context.Background())
	ledger.pendingErr = errors.New("synthetic ledger failure again")
	_ = s.sweep(context.Background())
	f.clk.Advance(10 * time.Minute)
	_ = s.sweep(context.Background())
	all := f.crit.withPrefix(prefix)
	if len(all) != 2 || all[0].Key == all[1].Key {
		t.Fatalf("a new streak after a success must be a new episode: %v", f.crit.keys())
	}
}

func TestA193StaleAndFirstSightAreCheckedBeforeResolving(t *testing.T) {
	f := newA193Fixture(t)
	first := f.attempt("SELL", journal.KindPlace, journal.StateInDoubt, false)
	second := f.attempt("BUY", journal.KindPlace, journal.StateInDoubt, false)
	f.clk.Advance(11 * time.Minute)
	res := newA193BlockingResolver()
	s := f.settler(nil, res)
	done := make(chan error, 1)
	go func() { done <- s.sweep(context.Background()) }()
	<-res.entered // 첫 해소가 막힘
	for _, id := range []string{first, second} {
		for _, key := range []string{"|attempt:", "|stale:"} {
			if n := len(f.crit.withPrefix(string(obs.EventOrderInDoubt) + key + id)); n != 1 {
				t.Fatalf("%s%s not recorded before the first resolution (%v)", key, id, f.crit.keys())
			}
		}
	}
	close(res.release)
	<-done
}

// --- 변이 원장이 요구한 못 (VERIFY 단계에 더함 — analysis/mutation) ---------------

func TestA193StaleIsMeasuredFromDispatch(t *testing.T) {
	// 기록은 오래전, 전송은 방금 — 경과는 전송 시각부터(기록 시각이 아니라).
	f := newA193Fixture(t)
	s := f.settler(nil, nil)
	rec := journal.AttemptRecord{ID: "att-synthetic-late-dispatch", Kind: journal.KindPlace, State: journal.StateInDoubt,
		RecordedAt:        a193Now.Add(-time.Hour).Format(time.RFC3339Nano),
		DispatchStartedAt: a193Now.Add(-time.Minute).Format(time.RFC3339Nano)}
	s.noteStale(context.Background(), rec)
	if n := f.crit.calls(); n != 0 {
		t.Fatalf("a one-minute-old dispatch was judged stale from its recorded time (%d)", n)
	}
}

func TestA193ACancelledRunTouchesNothing(t *testing.T) {
	f := newA193Fixture(t)
	ledger := &a193Ledger{Journal: f.journal, panicky: true}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := f.settler(ledger, nil).Run(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("Run = %v, want context.Canceled", err)
	}
	if n := ledger.pendingN.Load(); n != 0 {
		t.Fatalf("an already-cancelled run read the ledger %d time(s)", n)
	}
}

func TestA193ACancelledContextIsNeverRecorded(t *testing.T) {
	f := newA193Fixture(t)
	s := f.settler(nil, nil)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if s.record(ctx, "k", obs.Event{Type: obs.EventOrderInDoubt, Key: "k"}) {
		t.Fatal("record reported success on a cancelled context")
	}
	s.onStop(ctx, errors.New("synthetic"))
	if n := f.crit.calls(); n != 0 {
		t.Fatalf("%d critical record(s) attempted on a cancelled context — a failed record latches the entry gate", n)
	}
}

// --- 구현 리뷰 1회차 P1 의 못 -------------------------------------------------

func TestA193AFailedAlertIsRetriedAfterTheAttemptSettles(t *testing.T) {
	f := newA193Fixture(t)
	id := f.attempt("SELL", journal.KindPlace, journal.StateInDoubt, false)
	s := f.settler(nil, nil)
	f.crit.err = errors.New("synthetic outbox failure")
	if err := f.sweepDriven(s); err != nil {
		t.Fatal(err)
	}
	if got := f.state(id).State; got != journal.StateUnresolvedInDoubt {
		t.Fatalf("state = %s, want parked", got)
	}
	f.crit.err = nil
	if err := s.sweep(context.Background()); err != nil {
		t.Fatal(err)
	}
	got := f.crit.keys()
	want := []string{string(obs.EventOrderInDoubt) + "|attempt:" + id, string(obs.EventOrderUnresolved) + "|attempt:" + id}
	slices.Sort(got)
	slices.Sort(want)
	if !slices.Equal(got, want) {
		t.Fatalf("after the attempt settled, retried alerts = %v, want %v", got, want)
	}
	_ = s.sweep(context.Background())
	if n := len(f.crit.keys()); n != 2 {
		t.Fatalf("a retried alert was recorded again (%d)", n)
	}
}

func TestA193AResolutionFailureLogCarriesNoOrderNumber(t *testing.T) {
	f := newA193Fixture(t)
	f.attempt("SELL", journal.KindPlace, journal.StateInDoubt, false)
	res := newA193BlockingResolver()
	close(res.release)
	res.err = fmt.Errorf("journal: reading prior evidence for confirmed order %s: %w", "ord-synthetic-secret-77", journal.ErrUnexpectedState)
	var buf bytes.Buffer
	s := f.settler(nil, res)
	s.deps.Log = obs.NewLogger(obs.LogOptions{Writer: &buf, Clock: f.clk})
	if err := s.sweep(context.Background()); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if strings.Contains(out, "ord-synthetic-secret-77") {
		t.Fatalf("the settler logged the raw resolution error with an order number: %s", out)
	}
	if !strings.Contains(out, "ledger_state_changed") {
		t.Fatalf("the failure class is missing from the log: %s", out)
	}
}

func TestA193AFailingRecordIsTriedOncePerSweep(t *testing.T) {
	// 재시도 대기 중인 key 는 스윕 머리의 재시도가 맡음 — 같은 스윕에서 알림 함수가 또 시도하지 않음(변이 M30).
	f := newA193Fixture(t)
	f.attempt("SELL", journal.KindPlace, journal.StateInDoubt, false)
	res := newA193BlockingResolver()
	close(res.release)
	res.err = errors.New("synthetic")
	s := f.settler(nil, res)
	f.crit.err = errors.New("synthetic outbox failure")
	_ = s.sweep(context.Background()) // 처음 봄 1회 시도
	before := f.crit.calls()
	_ = s.sweep(context.Background()) // 재시도 1회만
	if n := f.crit.calls() - before; n != 1 {
		t.Fatalf("a failing alert was tried %d times in one sweep, want 1", n)
	}
}

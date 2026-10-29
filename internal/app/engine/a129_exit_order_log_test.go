package engine_test

// a129 — 엔진이 낸 청산 주문은 엔진 로그에 계좌 정보 없는 한 줄을 남김(tasks 1.1 RED).
//
// 무엇을 재나: 전용 로거(ExitOrderLog)에 청산 주문의 확정 · 결과 미상 · 거부 · 브로커 전 거절 · 체결 줄이 정확히 하나씩 남는지,
// 해제가 있는 갈래는 해제가 끝난 뒤에 줄이 쓰이는지, 줄에 계좌 모양 값이 없는지, 로거가 없으면 종전과 같은지, 로거가
// 패닉하거나 실패해도 청산이 진행되는지. 2026-10-07 010170 손절이 engine.log 에 0 줄이라 손절 미발동으로 오판된 사건이 동기.

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/app/engine"
	"github.com/JungHoonGhae/tossinvest-cli/internal/costs"
	"github.com/JungHoonGhae/tossinvest-cli/internal/domain"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/exitpolicy"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
	"github.com/JungHoonGhae/tossinvest-cli/internal/redact"
	"github.com/JungHoonGhae/tossinvest-cli/internal/riskcalc"
)

// --- 고정물 ---------------------------------------------------------------------

// a129Sink 는 전용 로거의 쓰기 대상임. onWrite 는 줄이 쓰이는 순간의 원장 상태를 재려고 둠(해제 뒤 줄 순서).
type a129Sink struct {
	mu      sync.Mutex
	buf     bytes.Buffer
	onWrite func(line []byte)
	panics  bool
	fails   bool
}

func (s *a129Sink) Write(p []byte) (int, error) {
	if s.panics {
		panic("a129: the log sink panicked")
	}
	if s.onWrite != nil {
		s.onWrite(p)
	}
	if s.fails {
		return 0, errors.New("a129: the log sink is full")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.buf.Write(p)
}

func (s *a129Sink) String() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.buf.String()
}

// a129Issuer 는 진짜 Guardian 을 감싸 거절(err)이나 발급 직전 훅(before)을 끼움 — B3(Guardian 거절) · B4(intent 부착 실패) 용.
type a129Issuer struct {
	inner  engine.ReductionIssuer
	err    error
	before func()
}

func (i *a129Issuer) IssueReduction(ctx context.Context, req execgw.ReductionIssuance) (execgw.Issued, error) {
	if i.before != nil {
		i.before()
	}
	if i.err != nil {
		return execgw.Issued{}, i.err
	}
	return i.inner.IssueReduction(ctx, req)
}

// a129Submitter 는 하네스의 가짜 제출기를 감싸, 원장에 매도를 남긴 뒤(진짜 게이트웨이처럼 intent · attempt 가 있음) 결과를
// 고치거나(decorate) Place 를 패닉시킴(panics — 관측 주기의 다른 패닉을 a129 의 recover 가 삼키지 않는지).
type a129Submitter struct {
	inner    *fakeSubmitter
	decorate func(*execgw.Outcome)
	panics   bool
}

func (s *a129Submitter) Place(ctx context.Context, req execgw.PlaceRequest) (execgw.Outcome, error) {
	if s.panics {
		panic("a129: the broker surface panicked")
	}
	out, err := s.inner.Place(ctx, req)
	if s.decorate != nil {
		s.decorate(&out)
	}
	return out, err
}

func (s *a129Submitter) Cancel(ctx context.Context, req execgw.CancelRequest) (execgw.Outcome, error) {
	return s.inner.Cancel(ctx, req)
}

func (s *a129Submitter) UnsettledOnSymbol(ctx context.Context, market, symbol string) ([]journal.AttemptRecord, error) {
	return s.inner.UnsettledOnSymbol(ctx, market, symbol)
}

// a129Floor 는 하네스의 확정 하한을 감싸 조회 직전 훅을 끼움 — B2(하한 0) 해제 실패 시험용.
type a129Floor struct {
	inner  engine.FloorSource
	before func()
}

func (f *a129Floor) ConfirmedFloor(ctx context.Context, market, symbol string) (riskcalc.ConfirmedFloor, bool, error) {
	if f.before != nil {
		f.before()
	}
	return f.inner.ConfirmedFloor(ctx, market, symbol)
}

type a129Fixture struct {
	*exitHarness
	sink      *a129Sink
	issuer    *a129Issuer
	submitter *a129Submitter
	floorHook *a129Floor
}

// a129Setup 은 exit 하네스에 전용 청산 주문 로거를 꽂음(mutate 로 더 바꿀 수 있음).
func a129Setup(t *testing.T, mutate func(*engine.ExitObserverOptions)) *a129Fixture {
	t.Helper()
	f := &a129Fixture{sink: &a129Sink{}}
	f.exitHarness = newExitHarness(t, func(o *engine.ExitObserverOptions) {
		f.issuer = &a129Issuer{inner: o.Issuer}
		o.Issuer = f.issuer
		f.submitter = &a129Submitter{inner: o.Submit.(*fakeSubmitter)}
		o.Submit = f.submitter
		f.floorHook = &a129Floor{inner: o.Floor}
		o.Floor = f.floorHook
		o.ExitOrderLog = obs.NewLogger(obs.LogOptions{Writer: f.sink, JSON: true, Clock: o.Clock})
		if mutate != nil {
			mutate(o)
		}
	})
	return f
}

// lines 는 전용 로거의 줄을 파싱해 돌려줌.
func (f *a129Fixture) lines() []map[string]any {
	f.t.Helper()
	return a129Parse(f.t, f.sink.String())
}

func a129Parse(t *testing.T, raw string) []map[string]any {
	t.Helper()
	var out []map[string]any
	for _, line := range strings.Split(strings.TrimSpace(raw), "\n") {
		if line == "" {
			continue
		}
		var m map[string]any
		if err := json.Unmarshal([]byte(line), &m); err != nil {
			t.Fatalf("an exit-order line is not JSON: %q: %v", line, err)
		}
		out = append(out, m)
	}
	return out
}

func a129Of(lines []map[string]any, event obs.EventType) []map[string]any {
	var out []map[string]any
	for _, l := range lines {
		if l["event"] == string(event) {
			out = append(out, l)
		}
	}
	return out
}

// breach 는 005930 10주(진입 70000, 손절 68000)를 열고 손절선 아래 67900 을 관측해 청산 매도를 내게 함.
func (f *a129Fixture) breach() (journal.Position, engine.ExitCycle) {
	f.t.Helper()
	p := f.entry("005930", "10", "70000", "68000", "70000")
	f.quote("005930", 70100)
	f.observe()
	f.quote("005930", 67900)
	f.clk.Advance(5 * time.Second)
	return p, f.observe()
}

// a129TradingDay 는 하네스가 청산 매도 intent 에 적는 거래일임(recordSubmittedSell).
const a129TradingDay = "2026-03-30"

func (f *a129Fixture) fill(orderID, state string, terminal bool, filled, avg string) {
	f.t.Helper()
	if _, err := f.journal.RecordFill(context.Background(), journal.FillObservation{
		OrderID: orderID, Symbol: "005930", Market: "kr", State: state, Terminal: terminal,
		Quantity: "10", FilledQuantity: filled, AveragePrice: avg,
		AccountRef: exitAccount, TradingDay: a129TradingDay, Side: "SELL",
		ObservedAt: f.clk.Now().Format(time.RFC3339),
	}); err != nil {
		f.t.Fatalf("RecordFill(%s %s): %v", orderID, state, err)
	}
}

func a129Want(t *testing.T, line map[string]any, want map[string]any) {
	t.Helper()
	for k, v := range want {
		if got, ok := line[k]; !ok || got != v {
			t.Errorf("line[%q] = %v (present %v), want %v — line %v", k, got, ok, v, line)
		}
	}
}

// --- ① 확정 · 체결 ------------------------------------------------------------

// 손절 확정은 order.confirmed 한 줄, 원장의 누적 체결이 늘 때마다 fill.observed 한 줄, 종결 뒤에는 더 없음.
func TestA129AConfirmedStopAndItsFillsEachLeaveALine(t *testing.T) {
	f := a129Setup(t, nil)
	p, cycle := f.breach()
	if cycle.Err != nil || len(f.submit.places) != 1 {
		t.Fatalf("the breach must submit one sell: err=%v places=%d", cycle.Err, len(f.submit.places))
	}
	place := f.submit.places[0]
	state := f.state(p.ID)

	confirmed := a129Of(f.lines(), obs.EventOrderConfirmed)
	if len(confirmed) != 1 {
		t.Fatalf("order.confirmed lines = %d, want exactly one: %s", len(confirmed), f.sink.String())
	}
	a129Want(t, confirmed[0], map[string]any{
		"origin": "exit", "symbol": "005930", "market": "kr", "side": "SELL", "order_type": "LIMIT",
		"quantity": "10", "price": "67900", "price_source": "observed", "observed_price": "67900", "baseline": "68000",
		"action": string(exitpolicy.ActionBaselineBreach), "exit_level": state.PendingLevel,
		"intent_id": place.IntentID, "position_id": redact.String(p.ID),
		"to_state": string(journal.StateConfirmed), "capped": false, "level": "INFO",
	})
	// 줄의 지정가는 실제로 제출된 지정가와 같아야 함(sellIntent 의 규칙을 옮겨 적은 값 — 둘이 갈리면 여기서 잡힘).
	if got, want := confirmed[0]["price"], decimalText(place.Intent.Price); got != want {
		t.Errorf("line price = %v, submitted limit = %s", got, want)
	}
	if id, _ := confirmed[0]["order_id"].(string); !strings.HasPrefix(id, "O-sell-") {
		t.Errorf("order_id = %q, want the broker order the submitter journaled", id)
	}
	if f.observer.FollowedExitOrdersForTest() != 1 {
		t.Fatalf("followed exit orders = %d, want the confirmed sell", f.observer.FollowedExitOrdersForTest())
	}

	orderID := orderOf(f)
	f.fill(orderID, "OPEN_PARTIALLY_FILLED", false, "4", "67900")
	f.clk.Advance(5 * time.Second)
	f.observe()
	fills := a129Of(f.lines(), obs.EventFillObserved)
	if len(fills) != 1 {
		t.Fatalf("fill.observed lines after the partial = %d, want one: %s", len(fills), f.sink.String())
	}
	a129Want(t, fills[0], map[string]any{
		"origin": "exit", "symbol": "005930", "side": "SELL", "order_id": orderID, "intent_id": place.IntentID,
		"position_id": redact.String(p.ID), "action": string(exitpolicy.ActionBaselineBreach),
		"filled_quantity": "4", "average_price": "67900", "order_quantity": "10", "terminal": false,
	})

	// 같은 누적이 다시 보이면 줄 없음.
	f.clk.Advance(5 * time.Second)
	f.observe()
	if got := len(a129Of(f.lines(), obs.EventFillObserved)); got != 1 {
		t.Fatalf("an unchanged cumulative fill wrote a line: fill lines = %d", got)
	}

	f.fill(orderID, "FILLED", true, "10", "67900")
	f.clk.Advance(5 * time.Second)
	f.observe()
	fills = a129Of(f.lines(), obs.EventFillObserved)
	if len(fills) != 2 {
		t.Fatalf("fill.observed lines after the full fill = %d, want two: %s", len(fills), f.sink.String())
	}
	a129Want(t, fills[1], map[string]any{"filled_quantity": "10", "terminal": true, "to_state": "FILLED"})
	if f.observer.FollowedExitOrdersForTest() != 0 {
		t.Fatal("a terminal fill must end the follow")
	}
	f.clk.Advance(5 * time.Second)
	f.observe()
	if got := len(a129Of(f.lines(), obs.EventFillObserved)); got != 2 {
		t.Fatalf("a finished order kept writing lines: %d", got)
	}
	if got := len(a129Of(f.lines(), obs.EventOrderConfirmed)); got != len(f.submit.places) {
		t.Fatalf("confirmed lines = %d for %d places", got, len(f.submit.places))
	}
}

// orderOf 는 줄에 실린 첫 청산 주문 번호임(가짜 제출기가 원장에 남긴 매도).
func orderOf(f *a129Fixture) string {
	f.t.Helper()
	for _, l := range f.lines() {
		if id, ok := l["order_id"].(string); ok && id != "" {
			return id
		}
	}
	f.t.Fatal("no confirmed exit order carries an order id")
	return ""
}

// 따라가기는 그 주기의 판정 뒤에 돔 — 같은 주기에 다른 포지션의 손절 확정 줄이 앞의 주문 체결 줄보다 먼저 나옴.
func TestA129BTheFillFollowRunsAfterTheJudgementOfTheSameCycle(t *testing.T) {
	f := a129Setup(t, nil)
	f.breach()
	first := orderOf(f)
	_ = f.entry("000660", "4", "120000", "118000", "120000")
	f.quote("000660", 120100)
	f.clk.Advance(5 * time.Second)
	f.observe()
	f.fill(first, "OPEN_PARTIALLY_FILLED", false, "3", "67900")
	f.quote("000660", 117900)
	f.clk.Advance(5 * time.Second)
	f.observe()

	lines := f.lines()
	confirmAt, fillAt := -1, -1
	for i, l := range lines {
		if l["event"] == string(obs.EventOrderConfirmed) && l["symbol"] == "000660" {
			confirmAt = i
		}
		if l["event"] == string(obs.EventFillObserved) && l["order_id"] == first {
			fillAt = i
		}
	}
	if confirmAt < 0 || fillAt < 0 {
		t.Fatalf("want the second stop's confirmation and the first order's fill: %s", f.sink.String())
	}
	if confirmAt > fillAt {
		t.Fatal("the fill follow ran before a stop judgement of the same cycle — it must run after every judgement")
	}
}

// 체결 감지 양보 주기(B1)에도 따라가기가 돔 — 손절 체결이 보이는 주기가 양보 주기일 수 있음.
func TestA129CTheFillFollowRunsOnAYieldedCycle(t *testing.T) {
	f := a129Setup(t, nil)
	f.breach()
	orderID := orderOf(f)
	f.fill(orderID, "FILLED", true, "10", "67900")
	f.slo.behind = true
	f.clk.Advance(5 * time.Second)
	if cycle := f.observe(); !cycle.Deferred {
		t.Fatal("the cycle must yield to fill detection")
	}
	if got := len(a129Of(f.lines(), obs.EventFillObserved)); got != 1 {
		t.Fatalf("fill lines on a yielded cycle = %d, want one", got)
	}
}

// --- ① 거부 · 결과 미상 · 브로커 전 거절 -----------------------------------------

// 갈래마다 정확히 한 줄. 줄은 주기 끝에 쓰이므로(D4-3) 해제가 있는 갈래는 그 줄이 쓰이는 순간 이미 해제돼 있음 — 이 시험이
// 고정하는 것은 「주기 끝 쓰기」이고, submit 안의 해제 · 대기열 순서는 메모리 append 라 관측 가능한 차이가 없음.
func TestA129DEveryOtherOutcomeLeavesOneLineAfterItsRelease(t *testing.T) {
	type outcome struct {
		name     string
		arrange  func(f *a129Fixture, p *journal.Position)
		event    obs.EventType
		reason   string
		state    string
		released bool // 해제가 있는 갈래 — 줄이 쓰이는 순간 발의가 이미 비어 있어야 함
		armed    bool // 결과 미상 — 발의가 무장된 채 남음
		followed int
	}
	cases := []outcome{
		{name: "B12 broker rejected", event: obs.EventOrderRejected, reason: string(execgw.ReasonBrokerRejected),
			state: string(journal.StateFailedConfirmed), released: true,
			arrange: func(f *a129Fixture, _ *journal.Position) {
				f.submit.placeOutcome = &execgw.Outcome{State: journal.StateFailedConfirmed, Reason: execgw.ReasonBrokerRejected, Detail: "no"}
			}},
		{name: "B12 not dispatched", event: obs.EventOrderNotDispatched, reason: string(execgw.ReasonGuardianLimitExceeded),
			state: string(journal.StateNotDispatched), released: true,
			arrange: func(f *a129Fixture, _ *journal.Position) {
				f.submit.placeOutcome = &execgw.Outcome{State: journal.StateNotDispatched, Reason: execgw.ReasonGuardianLimitExceeded}
			}},
		{name: "B9 symbol in flight", event: obs.EventOrderNotDispatched, reason: string(execgw.ReasonSymbolInFlight),
			released: true,
			arrange: func(f *a129Fixture, _ *journal.Position) {
				f.submit.placeOutcome = &execgw.Outcome{Reason: execgw.ReasonSymbolInFlight, Detail: "busy"}
			}},
		{name: "B8 in doubt", event: obs.EventOrderInDoubt, state: string(journal.StateInDoubt), armed: true, followed: 1,
			arrange: func(f *a129Fixture, _ *journal.Position) {
				f.submitter.decorate = func(o *execgw.Outcome) { o.State, o.Detail = journal.StateInDoubt, "transport gave up" }
			}},
		{name: "B8 unresolved in doubt", event: obs.EventOrderUnresolved, state: string(journal.StateUnresolvedInDoubt),
			armed: true, followed: 1,
			arrange: func(f *a129Fixture, _ *journal.Position) {
				f.submitter.decorate = func(o *execgw.Outcome) { o.State = journal.StateUnresolvedInDoubt }
			}},
		{name: "B10 unsettled attempt", event: obs.EventOrderInDoubt, armed: true,
			arrange: func(f *a129Fixture, _ *journal.Position) {
				f.submit.placeOutcome = &execgw.Outcome{AttemptID: "att-unsettled"}
			}},
		{name: "B2 zero floor", event: obs.EventOrderNotDispatched, reason: "confirmed_floor_zero", released: true,
			arrange: func(f *a129Fixture, _ *journal.Position) {
				f.floor.quantity, f.floor.bound, f.floor.applies = "0", "STALE", true
			}},
		{name: "B1 floor arithmetic failed", event: obs.EventOrderNotDispatched, reason: "floor_arithmetic_failed", armed: true,
			arrange: func(f *a129Fixture, _ *journal.Position) {
				f.floor.quantity, f.floor.bound, f.floor.applies = "not-a-number", "BROKEN", true
			}},
		{name: "B3 guardian refused", event: obs.EventOrderNotDispatched, reason: "guardian_refused", released: true,
			arrange: func(f *a129Fixture, _ *journal.Position) {
				f.issuer.err = errors.New("guardian says no")
			}},
		{name: "B4 attach failed", event: obs.EventOrderNotDispatched, reason: "attach_failed",
			arrange: func(f *a129Fixture, p *journal.Position) {
				f.issuer.before = func() {
					s := f.state(p.ID)
					if _, _, err := f.journal.ReleaseUnacceptedExitProposal(context.Background(), p.ID,
						s.PendingIntentID, journal.ProposalRefused); err != nil {
						f.t.Fatalf("releasing under the submission: %v", err)
					}
				}
			}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f := a129Setup(t, nil)
			p := f.entry("005930", "10", "70000", "68000", "70000")
			f.quote("005930", 70100)
			f.observe()
			tc.arrange(f, &p)
			pendingAtWrite := map[string]bool{}
			f.sink.onWrite = func(line []byte) {
				pendingAtWrite[string(line)] = f.state(p.ID).Pending()
			}
			f.quote("005930", 67900)
			f.clk.Advance(5 * time.Second)
			f.observe()
			f.sink.onWrite = nil

			lines := f.lines()
			if len(lines) != 1 {
				t.Fatalf("exit-order lines = %d, want exactly one: %s", len(lines), f.sink.String())
			}
			line := lines[0]
			a129Want(t, line, map[string]any{
				"event": string(tc.event), "origin": "exit", "symbol": "005930", "side": "SELL",
				"position_id": redact.String(p.ID), "level": "WARN",
			})
			if tc.reason != "" {
				a129Want(t, line, map[string]any{"reason": tc.reason})
			}
			if tc.state != "" {
				a129Want(t, line, map[string]any{"to_state": tc.state})
			}
			for raw, pending := range pendingAtWrite {
				if tc.released && pending {
					t.Errorf("the line was written while the proposal was still armed — the release must come first: %s", raw)
				}
			}
			if tc.released && f.state(p.ID).Pending() {
				t.Error("the refused proposal must be released as before")
			}
			if tc.armed && !f.state(p.ID).Pending() {
				t.Error("a submission that may have reached the broker keeps the proposal armed as before")
			}
			if got := f.observer.FollowedExitOrdersForTest(); got != tc.followed {
				t.Errorf("followed exit orders = %d, want %d", got, tc.followed)
			}
		})
	}
}

// --- ② 계좌 모양 값 ------------------------------------------------------------

// 계좌 참조 · 결과 상세 · 브로커 응답 본문 · 사유 문자열 · 오류 문구에 센티널을 심어도 어느 줄에도 나오지 않음.
func TestA129ENoLineCarriesTheAccount(t *testing.T) {
	const sentinel = "987-65-SENTINEL"
	for _, path := range []string{"confirmed", "rejected", "in_doubt"} {
		t.Run(path, func(t *testing.T) {
			f := a129Setup(t, nil)
			f.entry("005930", "10", "70000", "68000", "70000")
			f.quote("005930", 70100)
			f.observe()
			result := domain.MutationResult{
				Kind: "place", Status: "accepted " + sentinel, Symbol: "005930", Market: "kr",
				OrderDate: sentinel, Warnings: []string{"account " + exitAccount + " " + sentinel},
			}
			switch path {
			case "confirmed":
				f.submitter.decorate = func(o *execgw.Outcome) {
					o.AttemptID, o.Reason = "att-1", execgw.ReasonCode("broker_accepted "+sentinel)
					o.Detail, o.Result = exitAccount+" "+sentinel, result
				}
			case "rejected":
				f.submit.placeOutcome = &execgw.Outcome{
					State: journal.StateFailedConfirmed, AttemptID: "att-2",
					Reason: execgw.ReasonCode(sentinel), Detail: "account " + exitAccount + " refused " + sentinel, Result: result,
				}
			case "in_doubt":
				f.submitter.decorate = func(o *execgw.Outcome) {
					o.State, o.Detail, o.Result = journal.StateInDoubt, sentinel, result
				}
			}
			f.quote("005930", 67900)
			f.clk.Advance(5 * time.Second)
			f.observe()
			if path != "rejected" {
				// 체결 줄도 같은 거름을 탐 — 원장 체결 행의 계좌 필드에도 계좌를 실음.
				f.fill(orderOf(f), "FILLED", true, "10", "67900")
				f.clk.Advance(5 * time.Second)
				f.observe()
				if len(a129Of(f.lines(), obs.EventFillObserved)) != 1 {
					t.Fatalf("the fill line is missing — the fill path is not under the check: %s", f.sink.String())
				}
			}
			out := f.sink.String()
			if len(f.lines()) == 0 {
				t.Fatal("no exit-order line was written — the check would be vacuous")
			}
			for _, banned := range []string{exitAccount, sentinel, "987-65"} {
				if strings.Contains(out, banned) {
					t.Fatalf("an exit-order line carries %q:\n%s", banned, out)
				}
			}
			for _, l := range f.lines() {
				if r, ok := l["reason"].(string); ok && r != "" && r != "unrecognised" && strings.ContainsAny(r, "0123456789 ") {
					t.Errorf("reason %q passed the token filter", r)
				}
			}
		})
	}
}

// --- ③ 전용 로거 없음 ----------------------------------------------------------

// ExitOrderLog nil 이면 관측자의 다른 로그 출력은 종전과 같고(:1455 줄 하나 · 새 키 없음) 따라가기 항목도 없음.
func TestA129FWithoutTheDedicatedLoggerNothingChanges(t *testing.T) {
	logs := &a129Sink{}
	f := a129Setup(t, func(o *engine.ExitObserverOptions) {
		o.ExitOrderLog = nil
		o.Log = obs.NewLogger(obs.LogOptions{Writer: logs, JSON: true, Clock: o.Clock})
	})
	f.breach()
	if len(f.submit.places) != 1 {
		t.Fatalf("places = %d, want the liquidation", len(f.submit.places))
	}
	if f.observer.FollowedExitOrdersForTest() != 0 {
		t.Fatal("without the dedicated logger nothing is followed — and so nothing is read from the ledger")
	}
	lines := a129Parse(t, logs.String())
	var events []string
	for _, l := range lines {
		events = append(events, l["event"].(string))
		if _, ok := l["origin"]; ok {
			t.Errorf("a new exit-order line reached the observer's own Log: %v", l)
		}
	}
	if got := strings.Join(events, ","); got != "engine.started,order.confirmed" {
		t.Fatalf("observer Log events = %q, want the base pair engine.started,order.confirmed", got)
	}
	if f.sink.String() != "" {
		t.Fatal("the dedicated sink was written although no dedicated logger was wired")
	}
}

// --- ④ 로거 실패 · 패닉 --------------------------------------------------------

// 쓰기 대상이 패닉하거나 실패해도 확정 갈래는 무장된 발의로 끝나고(주기 오류 없음), 거부 갈래는 해제가 끝남.
func TestA129GAFailingOrPanickingLoggerDoesNotStopTheExit(t *testing.T) {
	for _, mode := range []string{"panics", "fails"} {
		for _, refused := range []bool{false, true} {
			name := mode + "/confirmed"
			if refused {
				name = mode + "/refused"
			}
			t.Run(name, func(t *testing.T) {
				f := a129Setup(t, nil)
				p := f.entry("005930", "10", "70000", "68000", "70000")
				f.quote("005930", 70100)
				f.observe()
				if refused {
					f.submit.placeOutcome = &execgw.Outcome{State: journal.StateFailedConfirmed, Reason: execgw.ReasonBrokerRejected}
				}
				f.sink.panics, f.sink.fails = mode == "panics", mode == "fails"
				f.quote("005930", 67900)
				f.clk.Advance(5 * time.Second)
				cycle := f.observe()
				if cycle.Err != nil {
					t.Fatalf("a broken log sink surfaced as a cycle error: %v", cycle.Err)
				}
				if len(f.submit.places) != 1 {
					t.Fatalf("places = %d, want the liquidation submitted regardless of the log", len(f.submit.places))
				}
				pending := f.state(p.ID).Pending()
				if refused && pending {
					t.Fatal("a refused proposal must still be released when the log sink breaks")
				}
				if !refused && !pending {
					t.Fatal("a confirmed sell keeps its proposal armed until the fill — unchanged by the log sink")
				}
				// 다음 주기도 그대로 돎(거부면 같은 단계가 다시 발의됨).
				f.clk.Advance(5 * time.Second)
				if again := f.observe(); again.Err != nil {
					t.Fatalf("the next cycle failed: %v", again.Err)
				}
				if refused && len(f.submit.places) != 2 {
					t.Fatalf("places = %d after the refusal, want the level proposed again", len(f.submit.places))
				}
			})
		}
	}
}

// --- ⑤ 생산 조립 통과 -----------------------------------------------------------

// Context.ExitObserver 는 ExitOrderLog 를 통과시키고 Log 는 nil 로 둠(a090 R17 과 같은 모양).
func TestA129TheEngineAssemblyPassesTheExitOrderLoggerThrough(t *testing.T) {
	dir := isolate(t)
	writeGateConfig(t, dir, smallLiveGate())
	writeCredentials(t, dir, "test-api-key-000000", "test-secret")
	writeAttestation(t, dir, nil)
	srv, _ := interlockServer(t, "123-45")
	eng, err := openProtectedGateEngine(t, dir, srv, nil)
	if err != nil {
		t.Fatalf("production assembly: %v", err)
	}
	logger := obs.NewLogger(obs.LogOptions{Writer: &a129Sink{}})
	observer, err := eng.ExitObserver(engine.ExitObserverOptions{Costs: costs.DefaultModel(), ExitOrderLog: logger})
	if err != nil {
		t.Fatalf("ExitObserver: %v", err)
	}
	opts := observer.OptionsForTest()
	if opts.ExitOrderLog != logger {
		t.Fatal("the engine assembly dropped the dedicated exit-order logger")
	}
	if opts.Log != nil {
		t.Fatal("the observer's own logger is wired — its lines carry the account")
	}
}

// --- 리뷰 보강(proposal-freeze P1 #1 · P2 #4 #5 #6) -------------------------------

// 재사용된 브로커 주문 번호 — 어제 같은 번호의 종결 체결 기록이 있어도 오늘 주문의 체결로 읽지 않음(LookupFillScoped).
func TestA129HAReusedOrderIDFromAnEarlierDayIsNotThisOrdersFill(t *testing.T) {
	f := a129Setup(t, nil)
	f.breach()
	orderID := orderOf(f)
	if _, err := f.journal.RecordFill(context.Background(), journal.FillObservation{
		OrderID: orderID, Symbol: "005930", Market: "kr", State: "FILLED", Terminal: true,
		Quantity: "3", FilledQuantity: "3", AveragePrice: "55000",
		AccountRef: exitAccount, TradingDay: "2026-03-27", Side: "SELL",
		ObservedAt: f.clk.Now().Format(time.RFC3339),
	}); err != nil {
		t.Fatalf("RecordFill(earlier day): %v", err)
	}
	f.clk.Advance(5 * time.Second)
	f.observe()
	for _, l := range f.lines() {
		if l["event"] == string(obs.EventFillObserved) || l["event"] == "exit.order_ended" {
			t.Fatalf("an earlier day's order under the same id was reported as this stop's fill: %v", l)
		}
	}
	if f.observer.FollowedExitOrdersForTest() != 1 {
		t.Fatal("the order of today is still unfilled and must still be followed")
	}
	f.fill(orderID, "FILLED", true, "10", "67900")
	f.clk.Advance(5 * time.Second)
	f.observe()
	fills := a129Of(f.lines(), obs.EventFillObserved)
	if len(fills) != 1 || fills[0]["filled_quantity"] != "10" {
		t.Fatalf("today's fill lines = %v, want one of 10", fills)
	}
}

// 체결 없이 끝난 청산 주문(취소 · 만료)은 exit.order_ended 한 줄 — 「손절이 한 주도 못 팔았다」 가 로그에 보임.
func TestA129IAnExitOrderThatEndsUnfilledLeavesALine(t *testing.T) {
	f := a129Setup(t, nil)
	f.breach()
	orderID := orderOf(f)
	f.fill(orderID, "CANCELLED", true, "0", "")
	f.clk.Advance(5 * time.Second)
	f.observe()
	ended := a129Of(f.lines(), "exit.order_ended")
	if len(ended) != 1 {
		t.Fatalf("exit.order_ended lines = %d, want one: %s", len(ended), f.sink.String())
	}
	a129Want(t, ended[0], map[string]any{"order_id": orderID, "filled_quantity": "0", "terminal": true,
		"to_state": "CANCELLED", "level": "WARN"})
	if len(a129Of(f.lines(), obs.EventFillObserved)) != 0 {
		t.Fatal("an unfilled order is not a fill")
	}
	// 취소가 발의를 풀면 같은 주기에 손절이 다시 나감(그 새 주문만 따라감) — 끝난 주문은 따라가기에서 빠져야 함.
	if got, want := f.observer.FollowedExitOrdersForTest(), len(f.submit.places)-1; got != want {
		t.Fatalf("followed = %d with %d places, want %d — a terminal order must end the follow", got, len(f.submit.places), want)
	}
}

// 같은 거절이 매 주기 반복되면 첫 줄만 — 결과가 바뀌면 다시 씀(5초마다 같은 줄은 관측이 아님).
func TestA129JARepeatedRefusalIsWrittenOnceUntilTheOutcomeChanges(t *testing.T) {
	f := a129Setup(t, nil)
	f.entry("005930", "10", "70000", "68000", "70000")
	f.quote("005930", 70100)
	f.observe()
	f.submit.placeOutcome = &execgw.Outcome{State: journal.StateFailedConfirmed, Reason: execgw.ReasonBrokerRejected}
	f.quote("005930", 67900)
	for i := 0; i < 3; i++ {
		f.clk.Advance(5 * time.Second)
		f.observe()
	}
	if len(f.submit.places) < 2 {
		t.Fatalf("places = %d — the refused level must be proposed again for this test to mean anything", len(f.submit.places))
	}
	if got := len(a129Of(f.lines(), obs.EventOrderRejected)); got != 1 {
		t.Fatalf("rejected lines = %d over %d identical refusals, want one", got, len(f.submit.places))
	}
	f.submit.placeOutcome = nil // 이제 브로커가 받음
	f.clk.Advance(5 * time.Second)
	f.observe()
	if got := len(a129Of(f.lines(), obs.EventOrderConfirmed)); got != 1 {
		t.Fatalf("confirmed lines = %d after the refusals cleared, want one", got)
	}
}

// a129 의 recover 는 쓰기 하나만 감쌈 — Place 의 패닉은 종전대로 관측 주기를 빠져나감(삼켜서 성공처럼 보이면 안 됨).
func TestA129KAPanicElsewhereInTheCycleIsNotSwallowed(t *testing.T) {
	f := a129Setup(t, nil)
	f.entry("005930", "10", "70000", "68000", "70000")
	f.quote("005930", 70100)
	f.observe()
	f.submitter.panics = true
	f.quote("005930", 67900)
	f.clk.Advance(5 * time.Second)
	recovered := func() (r any) {
		defer func() { r = recover() }()
		f.observe()
		return nil
	}()
	if recovered == nil {
		t.Fatal("a panicking broker surface was swallowed by the exit-order log path")
	}
}

// 원장에 intent 가 없는 주문(정규 신원을 못 얻음)은 exit.fill_follow_failed 한 줄 뒤 따라가기를 끝냄 — 원문 오류 없이, 다시 쓰지 않음.
func TestA129LAnOrderWithNoLedgerIdentityIsReportedOnceAndDropped(t *testing.T) {
	f := a129Setup(t, nil)
	f.entry("005930", "10", "70000", "68000", "70000")
	f.quote("005930", 70100)
	f.observe()
	// placeOutcome 은 원장에 매도를 남기지 않음 — 게이트웨이를 거치지 않은 주문 번호.
	f.submit.placeOutcome = &execgw.Outcome{State: journal.StateConfirmed, BrokerOrderID: "O-ghost"}
	f.quote("005930", 67900)
	f.clk.Advance(5 * time.Second)
	f.observe()
	failed := a129Of(f.lines(), "exit.fill_follow_failed")
	if len(failed) != 1 {
		t.Fatalf("exit.fill_follow_failed lines = %d, want one: %s", len(failed), f.sink.String())
	}
	a129Want(t, failed[0], map[string]any{"order_id": "O-ghost", "reason": "intent_lookup_failed", "level": "WARN"})
	if f.observer.FollowedExitOrdersForTest() != 0 {
		t.Fatal("an order with no ledger identity cannot be followed and must be dropped")
	}
	f.clk.Advance(5 * time.Second)
	f.observe()
	if got := len(a129Of(f.lines(), "exit.fill_follow_failed")); got != 1 {
		t.Fatalf("the follow failure was written %d times, want once", got)
	}
}

// 거절 래치는 연속한 주기만 묶음 — 가격이 회복된 뒤 다시 이탈해 같은 거절이 나면 새 에피소드라 다시 한 줄(구현 리뷰 P1).
func TestA129MTheRefusalLatchDoesNotOutliveItsEpisode(t *testing.T) {
	f := a129Setup(t, nil)
	f.entry("005930", "10", "70000", "68000", "70000")
	f.quote("005930", 70100)
	f.observe()
	f.submit.placeOutcome = &execgw.Outcome{State: journal.StateFailedConfirmed, Reason: execgw.ReasonBrokerRejected}
	f.quote("005930", 67900)
	f.clk.Advance(5 * time.Second)
	f.observe()
	f.clk.Advance(5 * time.Second)
	f.observe()              // 같은 에피소드 — 줄 없음
	f.quote("005930", 70200) // 회복 — 발의 없음
	f.clk.Advance(5 * time.Second)
	f.observe()
	placesBefore := len(f.submit.places)
	f.quote("005930", 67800) // 다시 이탈, 같은 거절
	f.clk.Advance(5 * time.Second)
	f.observe()
	if len(f.submit.places) == placesBefore {
		t.Fatal("the second breach must be submitted for this test to mean anything")
	}
	if got := len(a129Of(f.lines(), obs.EventOrderRejected)); got != 2 {
		t.Fatalf("rejected lines = %d over two separate episodes, want two: %s", got, f.sink.String())
	}
}

// 종결 기록이 수명(24h) 안에 안 오는 주문은 follow_expired 한 줄 뒤 따라가기를 끝냄 — 주기마다 원장 읽기가 재시작까지 남지 않게.
func TestA129NAFollowThatNeverEndsExpires(t *testing.T) {
	f := a129Setup(t, nil)
	f.breach()
	if f.observer.FollowedExitOrdersForTest() != 1 {
		t.Fatal("the confirmed sell must be followed")
	}
	f.clk.Advance(24 * time.Hour)
	f.observe()
	expired := a129Of(f.lines(), "exit.fill_follow_failed")
	if len(expired) != 1 || expired[0]["reason"] != "follow_expired" {
		t.Fatalf("expired lines = %v, want one follow_expired", expired)
	}
	if f.observer.FollowedExitOrdersForTest() != 0 {
		t.Fatal("an expired follow must be dropped")
	}
}

// 해제 갈래를 풀어 쓴 자리(releaseErr)는 해제 실패를 종전대로 돌려줌 — 원장을 닫아 해제를 실패시키고 주기 오류를 잼(구현 리뷰 P2).
func TestA129OAFailedReleaseStillSurfaces(t *testing.T) {
	cases := []struct {
		name    string
		arrange func(f *a129Fixture)
	}{
		{name: "B2 zero floor", arrange: func(f *a129Fixture) {
			f.floor.quantity, f.floor.bound, f.floor.applies = "0", "STALE", true
			f.floorHook.before = func() { _ = f.journal.Close() }
		}},
		{name: "B3 guardian refused", arrange: func(f *a129Fixture) {
			f.issuer.err = errors.New("guardian says no")
			f.issuer.before = func() { _ = f.journal.Close() }
		}},
		{name: "B10 symbol in flight", arrange: func(f *a129Fixture) {
			f.submit.placeOutcome = &execgw.Outcome{Reason: execgw.ReasonSymbolInFlight}
			f.submitter.decorate = func(*execgw.Outcome) { _ = f.journal.Close() }
		}},
		{name: "B13 broker rejected", arrange: func(f *a129Fixture) {
			f.submit.placeOutcome = &execgw.Outcome{State: journal.StateFailedConfirmed, Reason: execgw.ReasonBrokerRejected}
			f.submitter.decorate = func(*execgw.Outcome) { _ = f.journal.Close() }
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f := a129Setup(t, nil)
			f.entry("005930", "10", "70000", "68000", "70000")
			f.quote("005930", 70100)
			f.observe()
			tc.arrange(f)
			f.quote("005930", 67900)
			f.clk.Advance(5 * time.Second)
			cycle := f.observe()
			if cycle.Err == nil || !strings.Contains(cycle.Err.Error(), "releasing the exit proposal") {
				t.Fatalf("cycle error = %v, want the failed release surfaced", cycle.Err)
			}
		})
	}
}

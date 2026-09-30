package engine_test

// a094 §3 · §4 — 손절은 자기를 막는 것을 치운다(관측 루프 끝에서).
//
// 하네스는 exitloop_test.go 의 것(실제 원장 · 실제 Guardian · 가짜 브로커 면)이다. 이 파일은 거기에 둘을 더한다:
// 발주 attempt 를 원하는 상태로 남기는 제출자(a094Submitter)와 창 0 기록을 세는 critical 기록자(a094Criticals).

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/app/engine"
	"github.com/JungHoonGhae/tossinvest-cli/internal/config"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/exitpolicy"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
	"github.com/JungHoonGhae/tossinvest-cli/internal/trading"
)

// --- fakes -----------------------------------------------------------------------

// a094Criticals 는 알림기의 critical 기록 단일 입구를 대신함 — 받은 사건과 재알림 창을 셈.
type a094Criticals struct {
	events []obs.Event
	remind []time.Duration
	err    error
}

func (c *a094Criticals) RecordCritical(_ context.Context, e obs.Event, remindAfter time.Duration) error {
	if c.err != nil {
		return c.err
	}
	c.events = append(c.events, e)
	c.remind = append(c.remind, remindAfter)
	return nil
}

func (c *a094Criticals) withKey(fragment string) []obs.Event {
	var out []obs.Event
	for _, e := range c.events {
		if strings.Contains(e.Key, fragment) {
			out = append(out, e)
		}
	}
	return out
}

// a094Submitter 는 fakeSubmitter 의 취소 · 기본 발주를 쓰되, 발주 attempt 를 placeState 로 원장에 남길 수 있음.
type a094Submitter struct {
	*fakeSubmitter
	h *exitHarness
	// placeState 가 비어 있지 않으면 다음 발주의 attempt 를 그 상태로 기록하고 게이트웨이가 돌려줄 결과를 흉내 냄.
	placeState journal.AttemptState
	// placeErr 는 결과 쓰기 실패(상태 없는 결과 + 오류)를 흉내 냄 — attempt 는 DISPATCH_STARTED 로 남음.
	placeErr error
	// lastAttempt 는 마지막으로 기록한 발주 attempt 임.
	lastAttempt string
	lastOrder   string
	// cancelRecords 면 성공한 취소를 게이트웨이처럼 원장에 CANCEL attempt(CONFIRMED)로 남김 — 청소의 재취소 금지와
	// 종결 증거 대기 알림이 그 기록을 읽음.
	cancelRecords bool
}

func (s *a094Submitter) Place(ctx context.Context, req execgw.PlaceRequest) (execgw.Outcome, error) {
	if s.placeState == "" && s.placeErr == nil {
		out, err := s.fakeSubmitter.Place(ctx, req)
		s.lastOrder = out.BrokerOrderID
		return out, err
	}
	s.places = append(s.places, req)
	state := s.placeState
	if s.placeErr != nil {
		state = journal.StateDispatchStarted
	}
	attemptID, orderID := s.h.a094RecordSell(req.IntentID, req.Intent.Quantity, req.Intent.Price, state)
	s.lastAttempt, s.lastOrder = attemptID, orderID
	if s.placeErr != nil {
		return execgw.Outcome{IntentID: req.IntentID, AttemptID: attemptID}, s.placeErr
	}
	out := execgw.Outcome{IntentID: req.IntentID, AttemptID: attemptID, State: state, BrokerOrderID: orderID}
	switch state {
	case journal.StateUnresolvedInDoubt:
		// 게이트웨이는 IN_DOUBT 를 돌려주고, park 는 뒤에 해소가 만듦 — 시험은 기록만 park 로 둠.
		out.State = journal.StateInDoubt
	case journal.StateFailedConfirmed:
		out.Reason, out.Detail = execgw.ReasonOppositePendingOrder, "opposite-pending-order-exists"
	}
	return out, nil
}

func (s *a094Submitter) Cancel(ctx context.Context, req execgw.CancelRequest) (execgw.Outcome, error) {
	out, err := s.fakeSubmitter.Cancel(ctx, req)
	if err != nil || !s.cancelRecords {
		return out, err
	}
	s.h.a094RecordCancelOn(req.Intent.OrderID, strings.ToUpper(req.Order.Side), journal.StateConfirmed)
	return out, nil
}

// a094RecordSell 은 매도 PLACE attempt 하나를 원하는 상태로 원장에 남김.
func (h *exitHarness) a094RecordSell(intentID string, quantity, price float64, to journal.AttemptState) (string, string) {
	h.t.Helper()
	ctx := context.Background()
	h.ids++
	attemptID := fmt.Sprintf("a-a094-%d", h.ids)
	orderID := fmt.Sprintf("O-a094-%d", h.ids)
	attempt, err := h.journal.Prepare(ctx, journal.PrepareRequest{
		Intent: journal.Intent{
			ID: intentID, Market: "kr", TradingDay: "2026-03-30", AccountRef: exitAccount,
			Symbol: "005930", Side: "SELL", OrderType: "LIMIT", TimeInForce: "DAY",
			Quantity: decimalText(quantity), Price: decimalText(price), Currency: "KRW",
			Source: "engine/test", Fingerprint: "fp-" + intentID,
		},
		Kind: journal.KindPlace, AttemptID: attemptID, AccountRef: exitAccount,
	})
	if err != nil {
		h.t.Fatalf("Prepare: %v", err)
	}
	step := func(err error) {
		h.t.Helper()
		if err != nil {
			h.t.Fatalf("driving %s to %s: %v", attemptID, to, err)
		}
	}
	if to == journal.StateRecorded {
		return attemptID, ""
	}
	step(attempt.MarkDispatchStarted(ctx))
	switch to {
	case journal.StateDispatchStarted:
		return attemptID, ""
	case journal.StateFailedConfirmed:
		step(attempt.Settle(ctx, journal.StateFailedConfirmed, string(execgw.ReasonOppositePendingOrder), "409"))
		return attemptID, ""
	case journal.StateInDoubt, journal.StateUnresolvedInDoubt:
		step(attempt.MarkInDoubt(ctx, "dispatch_outcome_unknown", "409"))
		if to == journal.StateUnresolvedInDoubt {
			step(attempt.ResolveUnresolved(ctx, "in_doubt_unresolved", "parked"))
		}
		return attemptID, ""
	}
	step(attempt.MarkAcked(ctx, orderID))
	if to == journal.StateConfirmed {
		step(attempt.Settle(ctx, journal.StateConfirmed, "broker_accepted", ""))
	}
	return attemptID, orderID
}

// a094Harness 는 a094 제출자 · critical 기록자를 꽂은 하네스임.
func a094Harness(t *testing.T, mutate func(*engine.ExitObserverOptions)) (*exitHarness, *a094Submitter, *a094Criticals) {
	t.Helper()
	crit := &a094Criticals{}
	var sub *a094Submitter
	h := newExitHarness(t, func(o *engine.ExitObserverOptions) {
		sub = &a094Submitter{fakeSubmitter: o.Submit.(*fakeSubmitter)}
		o.Submit = sub
		o.Critical = crit
		if mutate != nil {
			mutate(o)
		}
	})
	sub.h = h
	return h, sub, crit
}

func (h *exitHarness) exitEventCount(positionID, action string) int {
	h.t.Helper()
	events, err := h.journal.ExitEvents(context.Background(), positionID)
	if err != nil {
		h.t.Fatalf("ExitEvents: %v", err)
	}
	n := 0
	for _, e := range events {
		if e.Action == action {
			n++
		}
	}
	return n
}

// a094ArmTakeProfit 은 40% 익절을 무장하고 그 attempt 를 state 로 남김.
func a094ArmTakeProfit(t *testing.T, h *exitHarness, sub *a094Submitter, state journal.AttemptState) journal.Position {
	t.Helper()
	p := h.entry("005930", "10", "70000", "68000", "70000")
	sub.placeState = state
	h.quote("005930", 72000)
	if c := h.observe(); c.Err != nil {
		t.Fatalf("arming cycle: %v", c.Err)
	}
	sub.placeState = ""
	if !h.state(p.ID).Pending() {
		t.Fatalf("control: the take-profit is not armed (state %s)", state)
	}
	return p
}

// --- 3.N1 · 3.N1a · 3.N1b · 3.R2 — 살아 있을 수 있는 발의는 비우지 않는다 ---------------------

// 3.N1 · 3.N1a — park 된 익절 위에서 손절 조건: 발의를 비우지 않고 제출 0, park 원인 critical 한 번(attempt 이름).
func TestA094AParkedTakeProfitIsNotClearedAndIsNamed(t *testing.T) {
	h, sub, crit := a094Harness(t, nil)
	p := a094ArmTakeProfit(t, h, sub, journal.StateUnresolvedInDoubt)
	parked := sub.lastAttempt
	places := len(h.submit.places)

	h.quote("005930", 68500) // 올라간 기준선 69000 아래 — CancelPendingFirst
	for i := 0; i < 3; i++ {
		if c := h.observe(); c.Err != nil {
			t.Fatalf("cycle %d: %v", i, c.Err)
		}
	}
	if !h.state(p.ID).Pending() || h.state(p.ID).PendingIntentID != h.submit.places[places-1].IntentID {
		t.Fatal("the parked take-profit's proposal was released — a live sell may still be on the book")
	}
	if got := len(h.submit.places); got != places {
		t.Fatalf("places = %d, want %d — no liquidation over a proposal whose order may be live", got, places)
	}
	if got := h.exitEventCount(p.ID, journal.ExitEventProposalCancelled); got != 0 {
		t.Errorf("PROPOSAL_CANCELLED = %d, want 0", got)
	}
	named := crit.withKey("|" + p.ID + "|attempt:" + parked)
	if len(named) != 1 {
		t.Fatalf("park-cause alerts for %s = %d, want exactly 1 (%+v)", parked, len(named), crit.events)
	}
	if named[0].Type != obs.EventOrderUnresolved || obs.SeverityOf(named[0].Type) != obs.SeverityCritical {
		t.Errorf("park-cause alert type = %s, want the critical %s", named[0].Type, obs.EventOrderUnresolved)
	}
	if !strings.Contains(named[0].Body, parked) || named[0].Fields[obs.FieldAttemptID] != parked {
		t.Errorf("the alert does not name the parked attempt: %+v", named[0])
	}
	for _, r := range crit.remind {
		if r != 0 {
			t.Errorf("remindAfter = %s, want 0 — an episode is one row", r)
		}
	}
	// 침묵하지 않는다: 보류가 한계를 넘으면 기존 지연 경보가 난다 — 청소가 「완료」 로 판정해 타이머를 지우면 안 난다
	// (arming 의 두 번째 발의 거절이 제출은 막아도 경보는 못 살림 — 변이 M12).
	h.clk.Advance(31 * time.Second)
	h.observe()
	if got := h.alerts.count(obs.EventExitLiquidationDelayed); got != 1 {
		t.Errorf("delay alerts = %d past the bound, want 1", got)
	}
}

// 3.R7(다른 intent) — 무장 발의와 다른 intent 의 엔진 매도를 취소한 주기: 그 매도의 취소 접수는 치움이 아님. 발의 판정이
// 해제를 허락해도(그 intent 에 attempt 없음) 손절은 그 매도의 종결 기록 전에 나가지 않음(변이 M10).
func TestA094AnotherIntentsCancelledSellStillHoldsTheStop(t *testing.T) {
	h, sub, _ := a094Harness(t, nil)
	p := h.entry("005930", "10", "70000", "68000", "70000")
	h.quote("005930", 70100)
	h.observe()
	h.a094ArmOn(p, "005930", "exit-tp-armed", string(exitpolicy.ActionRatchetPartial))
	h.a094RecordSell("old-sell", 4, 72000, journal.StateConfirmed)
	h.submit.settle = nil
	sub.cancelRecords = true
	h.quote("005930", 67900)
	h.observe()
	if len(h.submit.cancels) != 1 {
		t.Fatalf("cancels = %d, want the other intent's sell cancelled", len(h.submit.cancels))
	}
	if len(h.submit.places) != 0 {
		t.Fatalf("places = %d, want no stop over a sell whose cancel is only acknowledged", len(h.submit.places))
	}
}

// 3.N1a — key 는 포지션 단위: 두 포지션이면 두 key.
func TestA094TheParkCauseKeyIsPerPosition(t *testing.T) {
	h, _, crit := a094Harness(t, nil)
	ctx := context.Background()
	var ids []string
	for i, symbol := range []string{"005930", "000660"} {
		p := h.entry(symbol, "10", "70000", "68000", "70000")
		h.quote(symbol, 70100)
		h.observe() // exit 상태를 연다
		intentID := fmt.Sprintf("exit-park-%d", i)
		h.a094ArmOn(p, symbol, intentID, string(exitpolicy.ActionBaselineBreach))
		h.a094RecordSellOn(symbol, intentID, journal.StateUnresolvedInDoubt)
		ids = append(ids, p.ID)
		h.quote(symbol, 67000)
	}
	_ = ctx
	h.observe()
	for _, id := range ids {
		if got := len(crit.withKey("|" + id + "|attempt:")); got != 1 {
			t.Errorf("park-cause alerts for position %s = %d, want 1", id, got)
		}
	}
}

// 3.R2 — 무장 발의가 손절 자신이고 park: 평가는 억제 그대로, park 원인 critical 이 포지션 key 로 1회(사건 두 행의 모양).
func TestA094AParkedStopIsNamedOnceAndStaysSuppressed(t *testing.T) {
	h, sub, crit := a094Harness(t, nil)
	p := h.entry("005930", "10", "70000", "68000", "70000")
	sub.placeState = journal.StateUnresolvedInDoubt
	h.quote("005930", 67900)
	h.observe() // 손절 무장 · 제출 → IN_DOUBT → (기록상) park
	sub.placeState = ""
	parked := sub.lastAttempt
	if got := h.state(p.ID).PendingAction; got != string(exitpolicy.ActionBaselineBreach) {
		t.Fatalf("control: pending = %q, want the armed stop", got)
	}
	places := len(h.submit.places)

	for i := 0; i < 4; i++ {
		h.quote("005930", 67800-float64(i))
		h.observe()
	}
	if got := len(h.submit.places); got != places {
		t.Errorf("places = %d, want %d — the suppression stands", got, places)
	}
	if got := len(crit.withKey("|" + p.ID + "|attempt:" + parked)); got != 1 {
		t.Fatalf("park-cause alerts = %d, want exactly 1 across cycles", got)
	}
}

// 3.N1b — 발의 attempt 가 기록 · 전송 · 접수 · 모호면 해제하지 않고 park 원인 critical 도 없음.
func TestA094ALiveProposalAttemptIsNotReleased(t *testing.T) {
	for _, state := range []journal.AttemptState{
		journal.StateRecorded, journal.StateDispatchStarted, journal.StateAcked, journal.StateInDoubt,
	} {
		t.Run(string(state), func(t *testing.T) {
			h, sub, crit := a094Harness(t, nil)
			p := h.entry("005930", "10", "70000", "68000", "70000")
			h.quote("005930", 70100)
			h.observe() // exit 상태를 연다
			h.a094ArmOn(p, "005930", "exit-live", string(exitpolicy.ActionRatchetPartial))
			_, _ = sub, state
			h.a094RecordSellOn("005930", "exit-live", state)
			h.quote("005930", 67500)
			h.observe()
			if st := h.state(p.ID); !st.Pending() || st.PendingIntentID != "exit-live" {
				t.Fatalf("the proposal of a %s attempt was released", state)
			}
			if len(h.submit.places) != 0 {
				t.Errorf("places = %d, want 0", len(h.submit.places))
			}
			if got := len(crit.withKey("|attempt:")); got != 0 {
				t.Errorf("park-cause alerts = %d for a %s attempt, want 0", got, state)
			}
		})
	}
}

// a094ArmOn 은 판정 없이 원장에 발의를 직접 무장함(사건의 모양을 만드는 픽스처 — 관측이 만든 것과 같은 컬럼).
func (h *exitHarness) a094ArmOn(p journal.Position, symbol, intentID, action string) {
	h.t.Helper()
	st := h.state(p.ID)
	level := string(exitpolicy.LevelNone)
	if action == string(exitpolicy.ActionRatchetPartial) {
		level = string(exitpolicy.LevelBreakeven)
	}
	if err := h.journal.RecordExitJudgement(context.Background(), journal.ExitJudgement{
		PositionID: p.ID, LifecycleGeneration: st.LifecycleGeneration, ObservedPrice: "70000",
		HighWater: st.HighWater, Baseline: st.Baseline, RatchetLevel: st.RatchetLevel, ActiveRung: exitpolicy.NoRung,
		Proposal: &journal.ExitProposal{Action: action, Level: level, IntentID: intentID},
	}); err != nil {
		h.t.Fatalf("arming %s on %s: %v", intentID, symbol, err)
	}
}

func (h *exitHarness) a094RecordSellOn(symbol, intentID string, to journal.AttemptState) (string, string) {
	h.t.Helper()
	if symbol != "005930" {
		// a094RecordSell 은 005930 전용 — 다른 종목은 같은 모양을 여기서 만듦.
		ctx := context.Background()
		h.ids++
		attemptID := fmt.Sprintf("a-a094-%d", h.ids)
		attempt, err := h.journal.Prepare(ctx, journal.PrepareRequest{
			Intent: journal.Intent{ID: intentID, Market: "kr", TradingDay: "2026-03-30", AccountRef: exitAccount,
				Symbol: symbol, Side: "SELL", OrderType: "LIMIT", TimeInForce: "DAY", Quantity: "10", Price: "67000",
				Currency: "KRW", Source: "engine/test", Fingerprint: "fp-" + intentID},
			Kind: journal.KindPlace, AttemptID: attemptID, AccountRef: exitAccount,
		})
		if err != nil {
			h.t.Fatal(err)
		}
		if err := attempt.MarkDispatchStarted(ctx); err != nil {
			h.t.Fatal(err)
		}
		if err := attempt.MarkInDoubt(ctx, "dispatch_outcome_unknown", "x"); err != nil {
			h.t.Fatal(err)
		}
		if to == journal.StateUnresolvedInDoubt {
			if err := attempt.ResolveUnresolved(ctx, "in_doubt_unresolved", "parked"); err != nil {
				h.t.Fatal(err)
			}
		}
		return attemptID, ""
	}
	return h.a094RecordSell(intentID, 10, 67000, to)
}

// --- 3.R7 · 3.R7c · 3.R7d · 3.R7e · 3.R9 — 매도의 취소 접수는 치움이 아니다 --------------------

// a094ConfirmedTakeProfit 은 접수 확정된 익절 매도를 무장하고, 취소가 종결 스냅숏을 **기록하지 않게** 함(형태 B).
func a094ConfirmedTakeProfit(t *testing.T) (*exitHarness, *a094Submitter, *a094Criticals, journal.Position, string) {
	t.Helper()
	h, sub, crit := a094Harness(t, nil)
	p := a094ArmTakeProfit(t, h, sub, journal.StateConfirmed)
	orderID := sub.lastOrder
	if orderID == "" {
		t.Fatal("control: the take-profit order has no id")
	}
	h.submit.settle = nil    // 체결 감지가 아직 종결을 기록하지 않음
	sub.cancelRecords = true // 취소 attempt 를 원장에 CONFIRMED 로 남김(게이트웨이처럼)
	return h, sub, crit, p, orderID
}

// 3.R7 — 취소 CONFIRMED · 종결 스냅숏 없음: 해제 0 · 제출 0 · 재취소 0. 종결 기록 뒤 주기: 해제 → 손절 제출.
func TestA094ACancelledSellIsNotClearedUntilItsClosingRecord(t *testing.T) {
	h, _, _, p, orderID := a094ConfirmedTakeProfit(t)
	places := len(h.submit.places)

	h.quote("005930", 68500)
	h.observe()
	if got := len(h.submit.cancels); got != 1 {
		t.Fatalf("cancels = %d, want the take-profit cancelled once", got)
	}
	for i := 0; i < 3; i++ {
		h.observe()
	}
	if got := len(h.submit.cancels); got != 1 {
		t.Errorf("cancels = %d, want 1 — a confirmed cancel is not re-sent while its closing record is awaited", got)
	}
	if got := len(h.submit.places); got != places {
		t.Fatalf("places = %d, want %d — no stop over a sell whose close is unproven", got, places)
	}
	if !h.state(p.ID).Pending() {
		t.Fatal("the take-profit's proposal was released before its closing record")
	}

	h.settleCancelled(orderID)
	h.observe()
	if got := len(h.submit.places); got != places+1 {
		t.Fatalf("places = %d after the closing record, want the stop (%d)", got, places+1)
	}
	if got := h.state(p.ID).PendingAction; got != string(exitpolicy.ActionBaselineBreach) {
		t.Errorf("pending = %q, want the stop armed", got)
	}
}

// 3.R7d · 3.R7e · 3.R7c — 종결 기록이 계속 오지 않으면: 시간이 지나도 해제 · 제출 0, 지연 경보 1 · 연속 실패 경보 1, 둘 다
// 복구 절차 문장을 본문에 싣음.
func TestA094AClosingRecordThatNeverComesHoldsAndAlerts(t *testing.T) {
	h, _, crit, p, _ := a094ConfirmedTakeProfit(t)
	places := len(h.submit.places)
	h.quote("005930", 68500)
	for i := 0; i < 12; i++ {
		h.observe()
		h.clk.Advance(5 * time.Second)
	}
	if got := len(h.submit.places); got != places {
		t.Fatalf("places = %d after a minute, want %d — time alone never releases", got, places)
	}
	if !h.state(p.ID).Pending() {
		t.Fatal("the proposal was released by the passage of time")
	}
	if got := h.alerts.count(obs.EventExitLiquidationDelayed); got != 1 {
		t.Errorf("delay alerts = %d, want 1", got)
	}
	delay, _ := h.alerts.first(obs.EventExitLiquidationDelayed)
	streak := crit.withKey("|" + p.ID + "|streak:")
	if len(streak) != 1 {
		t.Fatalf("streak alerts = %d, want 1 (%+v)", len(streak), crit.events)
	}
	for _, body := range []string{delay.Body, streak[0].Body} {
		if !strings.Contains(body, "체결 감지가 멈췄을 때") {
			t.Errorf("the alert body does not name the recovery procedure: %q", body)
		}
	}
}

// 3.R7e — 형태 B 가 아닌 지연 경보의 본문은 무변화.
func TestA094AnOrdinaryDelayKeepsItsBody(t *testing.T) {
	h, sub, _ := a094Harness(t, nil)
	h.entry("005930", "10", "70000", "68000", "70000")
	h.workingEntry("005930", "5", "69500")
	sub.cancelFails = true
	h.quote("005930", 67900)
	h.observe()
	h.clk.Advance(31 * time.Second)
	h.observe()
	alert, ok := h.alerts.first(obs.EventExitLiquidationDelayed)
	if !ok {
		t.Fatal("control: no delay alert")
	}
	if !strings.HasPrefix(alert.Body, "a working order on the symbol could not be taken off the book") ||
		strings.Contains(alert.Body, "체결 감지가 멈췄을 때") {
		t.Errorf("an ordinary delay's body changed: %q", alert.Body)
	}
}

// 3.R9 — 건강한 감지 · 비종결: 취소 settled_at 뒤 30초 미만 알림 0, 이상 1회(key type|position|cancel:<id>), 반복 0,
// noteDelay key 와 겹치지 않음.
func TestA094TheClosingEvidenceWaitIsNamedOnce(t *testing.T) {
	h, _, crit, p, _ := a094ConfirmedTakeProfit(t)
	h.quote("005930", 68500)
	h.observe() // 취소 CONFIRMED(지금 시각)
	h.clk.Advance(20 * time.Second)
	h.observe()
	if got := len(crit.withKey("|cancel:")); got != 0 {
		t.Fatalf("close-wait alerts at 20s = %d, want 0", got)
	}
	h.clk.Advance(11 * time.Second)
	h.observe()
	h.clk.Advance(5 * time.Second)
	h.observe()
	named := crit.withKey("|" + p.ID + "|cancel:")
	if len(named) != 1 {
		t.Fatalf("close-wait alerts = %d, want exactly 1 (%+v)", len(named), crit.events)
	}
	if named[0].Type != obs.EventExitLiquidationDelayed {
		t.Errorf("type = %s, want %s", named[0].Type, obs.EventExitLiquidationDelayed)
	}
	if named[0].Key == string(obs.EventExitLiquidationDelayed)+"|"+p.ID {
		t.Error("the close-wait key collides with the delay timer's key")
	}
	if _, ok := named[0].Fields[obs.FieldAccount]; ok {
		t.Error("the close-wait alert carries an account field")
	}
}

// --- 3.R5 · 3.R5a — 다른 intent 의 미종결 위에서 반복하지 않는다 ------------------------------

// 3.R5 — 무장 없음 + 같은 종목에 다른 intent 의 IN_DOUBT: clear=false · 무장 0 · PROPOSAL_CANCELLED 0 · 지연 경보가 한계에서.
// 미종결 판정은 **실제 게이트웨이의 메서드**다(3.R5a 의 행동 절반 — 그 메서드를 변이하면 이 시험이 깨짐).
func TestA094AnotherIntentInFlightStopsTheArmReleaseLoop(t *testing.T) {
	var gw *execgw.Gateway
	h, sub, _ := a094Harness(t, nil)
	gw = a094RealGateway(t, h)
	sub.unsettled = gw.UnsettledOnSymbol
	p := h.entry("005930", "10", "70000", "68000", "70000")
	h.a094RecordSellOn("005930", "other-intent", journal.StateInDoubt)
	h.quote("005930", 67900)
	for i := 0; i < 8; i++ {
		h.observe()
		h.clk.Advance(5 * time.Second)
	}
	if h.state(p.ID).Pending() || len(h.submit.places) != 0 {
		t.Fatalf("armed %v / places %d — a stop the gateway will refuse must not be armed", h.state(p.ID).Pending(), len(h.submit.places))
	}
	if got := h.exitEventCount(p.ID, journal.ExitEventProposalCancelled); got != 0 {
		t.Errorf("PROPOSAL_CANCELLED = %d, want 0 (272210's loop)", got)
	}
	if got := h.alerts.count(obs.EventExitLiquidationDelayed); got != 1 {
		t.Errorf("delay alerts = %d, want 1 — the timer is no longer wiped every cycle", got)
	}
}

// 3.R5a — 구조: 청소와 checkSymbolFree 가 같은 판정 함수(unsettledFor)를 부른다.
func TestA094TheClearAndTheGatewayShareOneUnsettledJudgement(t *testing.T) {
	calls := func(path, fn string) map[string]bool {
		fset := token.NewFileSet()
		f, err := parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		out := map[string]bool{}
		for _, d := range f.Decls {
			fd, ok := d.(*ast.FuncDecl)
			if !ok || fd.Name.Name != fn {
				continue
			}
			ast.Inspect(fd.Body, func(n ast.Node) bool {
				if c, ok := n.(*ast.CallExpr); ok {
					switch f := c.Fun.(type) {
					case *ast.SelectorExpr:
						out[f.Sel.Name] = true
					case *ast.Ident:
						out[f.Name] = true
					}
				}
				return true
			})
		}
		if len(out) == 0 {
			t.Fatalf("control: %s not found in %s", fn, path)
		}
		return out
	}
	if !calls("exitloop.go", "clearTheSymbol")["UnsettledOnSymbol"] {
		t.Error("clearTheSymbol does not ask the gateway's UnsettledOnSymbol")
	}
	gw := "../../execgw/gateway.go"
	shared := "../../execgw/a094_shared_judgements.go"
	if c := calls(gw, "checkSymbolFree"); !c["unsettledFor"] || c["PendingAttempts"] {
		t.Errorf("checkSymbolFree must call unsettledFor and not read PendingAttempts itself: %v", c)
	}
	if !calls(shared, "UnsettledOnSymbol")["unsettledFor"] {
		t.Error("UnsettledOnSymbol must be unsettledFor")
	}
}

// --- 3.E4 · 3.E5 — 연속 실패의 이른 트리거 ------------------------------------------------

// 3.E4 — N = obs.DefaultCriticalAttempts 번 연속 실패면 이른 경보 1회(key 는 연속 id), 기존 30초 경보도 난다, 제출 0.
func TestA094ConsecutiveClearFailuresRaiseAnEarlierAlert(t *testing.T) {
	h, sub, crit := a094Harness(t, nil)
	p := h.entry("005930", "10", "70000", "68000", "70000")
	h.workingEntry("005930", "5", "69500")
	sub.cancelFails = true
	h.quote("005930", 67900)
	for i := 0; i < obs.DefaultCriticalAttempts-1; i++ {
		h.observe()
	}
	if got := len(crit.withKey("|streak:")); got != 0 {
		t.Fatalf("streak alerts after %d failures = %d, want 0", obs.DefaultCriticalAttempts-1, got)
	}
	h.observe()
	first := crit.withKey("|" + p.ID + "|streak:")
	if len(first) != 1 {
		t.Fatalf("streak alerts at the %dth failure = %d, want 1", obs.DefaultCriticalAttempts, len(first))
	}
	h.observe()
	h.clk.Advance(31 * time.Second)
	h.observe()
	if got := len(crit.withKey("|streak:")); got != 1 {
		t.Errorf("streak alerts = %d, want still 1", got)
	}
	if got := h.alerts.count(obs.EventExitLiquidationDelayed); got != 1 {
		t.Errorf("the timer's delay alert = %d, want 1 — both alerts go out", got)
	}
	if len(h.submit.places) != 0 {
		t.Error("failures must never submit over an uncleared conflict")
	}

	// 연속이 끝나면(치움 성공) 다음 연속은 새 id — 새 에피소드.
	sub.cancelFails = false
	h.observe()
	if len(h.submit.places) != 1 {
		t.Fatalf("control: the clear did not complete once the cancel went through")
	}
}

// 4.N4f ④ — 같은 벽시계 시각에 시작한 두 연속도 다른 key(연속 id).
func TestA094TwoStreaksAtTheSameInstantAreTwoEpisodes(t *testing.T) {
	h, sub, crit := a094Harness(t, nil)
	h.entry("005930", "10", "70000", "68000", "70000")
	entry := h.workingEntry("005930", "5", "69500")
	sub.cancelFails = true
	h.quote("005930", 67900)
	for i := 0; i < obs.DefaultCriticalAttempts; i++ {
		h.observe()
	}
	// 연속 끝: 취소가 통하고 종목이 깨끗해짐 → 제출. 그 뒤 새 매수가 막고 다시 실패.
	sub.cancelFails = false
	h.observe()
	_ = entry
	keys := map[string]bool{}
	for _, e := range crit.withKey("|streak:") {
		keys[e.Key] = true
	}
	if len(keys) != 1 {
		t.Fatalf("control: streak keys = %v", keys)
	}
}

// 3.E5 — 치우지 못한 것이 전부 엔진 취소의 기록 · 전송 · 인수 단계면 계수를 늘리지 않음. IN_DOUBT 취소는 셈.
func TestA094AnInFlightEngineCancelIsNotCountedButAnInDoubtOneIs(t *testing.T) {
	for _, tc := range []struct {
		state  journal.AttemptState
		counts bool
	}{
		{journal.StateRecorded, false},
		{journal.StateDispatchStarted, false},
		{journal.StateAcked, false},
		{journal.StateInDoubt, true},
	} {
		t.Run(string(tc.state), func(t *testing.T) {
			h, sub, crit := a094Harness(t, nil)
			gw := a094RealGateway(t, h)
			sub.unsettled = gw.UnsettledOnSymbol
			p := h.entry("005930", "10", "70000", "68000", "70000")
			h.a094RecordCancel("O-target", tc.state)
			h.quote("005930", 67900)
			for i := 0; i < obs.DefaultCriticalAttempts+1; i++ {
				h.observe()
			}
			got := len(crit.withKey("|" + p.ID + "|streak:"))
			if tc.counts && got != 1 {
				t.Errorf("%s cancel: streak alerts = %d, want 1", tc.state, got)
			}
			if !tc.counts && got != 0 {
				t.Errorf("%s cancel: streak alerts = %d, want 0 — it is being cleared, not failing", tc.state, got)
			}
			// 기존 타이머에는 제외가 없음.
			h.clk.Advance(31 * time.Second)
			h.observe()
			if got := h.alerts.count(obs.EventExitLiquidationDelayed); got != 1 {
				t.Errorf("%s cancel: delay alerts = %d, want 1 — the timer takes no exclusion", tc.state, got)
			}
		})
	}
}

func (h *exitHarness) a094RecordCancel(target string, to journal.AttemptState) {
	h.a094RecordCancelOn(target, "BUY", to)
}

func (h *exitHarness) a094RecordCancelOn(target, side string, to journal.AttemptState) {
	h.t.Helper()
	ctx := context.Background()
	h.ids++
	id := fmt.Sprintf("c-a094-%d", h.ids)
	attempt, err := h.journal.Prepare(ctx, journal.PrepareRequest{
		Intent: journal.Intent{ID: "i-" + id, Market: "kr", TradingDay: "2026-03-30", AccountRef: exitAccount,
			Symbol: "005930", Side: side, OrderType: "LIMIT", TimeInForce: "DAY", Quantity: "5", Price: "69500",
			Currency: "KRW", Source: "engine/test", Fingerprint: "fp-" + id},
		Kind: journal.KindCancel, AttemptID: id, TargetOrderID: target, AccountRef: exitAccount,
	})
	if err != nil {
		h.t.Fatal(err)
	}
	if to == journal.StateRecorded {
		return
	}
	if err := attempt.MarkDispatchStarted(ctx); err != nil {
		h.t.Fatal(err)
	}
	switch to {
	case journal.StateAcked, journal.StateConfirmed:
		if err := attempt.MarkAcked(ctx, target); err != nil {
			h.t.Fatal(err)
		}
		if to == journal.StateConfirmed {
			if err := attempt.Settle(ctx, journal.StateConfirmed, "broker_accepted", "cancelled"); err != nil {
				h.t.Fatal(err)
			}
		}
	case journal.StateInDoubt:
		if err := attempt.MarkInDoubt(ctx, "dispatch_outcome_unknown", "x"); err != nil {
			h.t.Fatal(err)
		}
	}
}

// --- §4 — 비수용 종결이 발의를 푼다 -----------------------------------------------------

// 2.9 · 4.1 — R1 의 FAILED_CONFIRMED 는 default 갈래(알림 + 해제)로 가고, 쓸 수 있는 시세의 다음 관측이 손절을 다시 발의함.
func TestA094AnUnacceptedStopIsReleasedAndProposedAgain(t *testing.T) {
	for _, policy := range []string{"RATCHET", "LADDER"} {
		t.Run(policy, func(t *testing.T) {
			h, sub, _ := a094Harness(t, func(o *engine.ExitObserverOptions) {
				if policy == "LADDER" {
					ladder := exitpolicy.DefaultLadderPolicy()
					o.Ladder = &ladder
				}
			})
			p := h.entry("005930", "10", "70000", "68000", "70000")
			if policy == "LADDER" {
				if _, err := h.journal.OpenExitState(context.Background(), journal.ExitStateSeed{
					PositionID: p.ID, PolicyKind: journal.ExitPolicyLadder, EntryPrice: "70000", InitialStop: "68000",
				}); err != nil {
					t.Fatalf("OpenExitState: %v", err)
				}
			}
			sub.placeState = journal.StateFailedConfirmed
			h.quote("005930", 67000)
			h.observe()
			if h.state(p.ID).Pending() {
				t.Fatal("a definitively refused stop kept its proposal armed")
			}
			if got := h.alerts.count(obs.EventExitProposalRefused); got != 1 {
				t.Errorf("refusal alerts = %d, want 1", got)
			}
			sub.placeState = ""
			h.quote("005930", 66900)
			h.observe()
			if got := len(h.submit.places); got != 2 {
				t.Fatalf("places = %d, want the stop proposed again", got)
			}
		})
	}
}

// 4.3c — 결과를 쓰지 못한 제출(상태 없음 + attempt 있음)은 발의를 무장된 채 둠.
func TestA094AnUnrecordedOutcomeKeepsTheProposalArmed(t *testing.T) {
	h, sub, _ := a094Harness(t, nil)
	p := h.entry("005930", "10", "70000", "68000", "70000")
	sub.placeErr = errors.New("journal: writing the dispatch outcome: disk full")
	h.quote("005930", 67000)
	cycle := h.observe()
	if !h.state(p.ID).Pending() {
		t.Fatal("the proposal was released although the order may have gone out")
	}
	if cycle.Err == nil {
		t.Error("the unrecorded outcome must surface as a cycle error")
	}
	if got := h.alerts.count(obs.EventExitProposalRefused); got != 0 {
		t.Errorf("refusal alerts = %d — nothing was refused", got)
	}
}

// 4.3 — park + 살아 있는 매도 가정: 두 번째 매도가 무장되지 않음.
func TestA094ParkedStopArmsNoSecondSell(t *testing.T) {
	h, sub, _ := a094Harness(t, nil)
	p := h.entry("005930", "10", "70000", "68000", "70000")
	sub.placeState = journal.StateUnresolvedInDoubt
	h.quote("005930", 67900)
	h.observe()
	sub.placeState = ""
	intent := h.state(p.ID).PendingIntentID
	for i := 0; i < 3; i++ {
		h.quote("005930", 67000-float64(i*100))
		h.observe()
	}
	if got := h.state(p.ID).PendingIntentID; got != intent || len(h.submit.places) != 1 {
		t.Fatalf("pending %s→%s, places %d — a second sell was armed over a parked one", intent, got, len(h.submit.places))
	}
}

// 3.N1d · 4.3e — 운영자가 park 를 비수용으로 닫고 같은 판정 함수가 풀면, 다음 관측이 손절을 발의함.
func TestA094AnOperatorClosedParkReleasesAndTheStopFollows(t *testing.T) {
	h, sub, _ := a094Harness(t, nil)
	ctx := context.Background()
	p := h.entry("005930", "10", "70000", "68000", "70000")
	sub.placeState = journal.StateUnresolvedInDoubt
	h.quote("005930", 67900)
	h.observe()
	sub.placeState = ""
	parked, intent := sub.lastAttempt, h.state(p.ID).PendingIntentID
	for i := 0; i < 3; i++ {
		h.observe()
	}
	if len(h.submit.places) != 1 {
		t.Fatal("control: the park released itself")
	}
	if err := h.journal.OperatorResolve(ctx, parked, journal.StateFailedConfirmed, "op", "", "broker shows nothing"); err != nil {
		t.Fatalf("OperatorResolve: %v", err)
	}
	if _, released, err := h.journal.ReleaseUnacceptedExitProposal(ctx, p.ID, intent, journal.ProposalRefused); err != nil || !released {
		t.Fatalf("release = %v/%v", released, err)
	}
	h.quote("005930", 67800)
	h.observe()
	if len(h.submit.places) != 2 {
		t.Fatalf("places = %d, want the stop once the park is closed", len(h.submit.places))
	}
}

// 4.4a — 관측 루프는 재시작 규칙(RecoverPending)을 세션 중에 부르지 않음(구조).
func TestA094TheSessionNeverRunsTheRestartRules(t *testing.T) {
	for _, file := range []string{"exitloop.go", "exit_held_proposal.go"} {
		fset := token.NewFileSet()
		f, err := parser.ParseFile(fset, file, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		ast.Inspect(f, func(n ast.Node) bool {
			if s, ok := n.(*ast.SelectorExpr); ok && s.Sel.Name == "RecoverPending" {
				t.Errorf("%s references RecoverPending at %s — the restart rules forge the ledger mid-session", file, fset.Position(s.Pos()))
			}
			return true
		})
	}
}

// 3.R4 — 새 critical 은 기록 입구로만: 관측 루프의 Notify 로 가지 않고 창 0.
func TestA094TheNewCriticalsAreRecordedNotNotified(t *testing.T) {
	h, sub, crit := a094Harness(t, nil)
	p := a094ArmTakeProfit(t, h, sub, journal.StateUnresolvedInDoubt)
	h.quote("005930", 68500)
	h.observe()
	if len(crit.withKey("|"+p.ID+"|attempt:")) != 1 {
		t.Fatal("control: no park-cause record")
	}
	for _, e := range h.alerts.events {
		if strings.Contains(e.Key, "|attempt:") || strings.Contains(e.Key, "|streak:") || strings.Contains(e.Key, "|cancel:") {
			t.Errorf("an a094 critical went through the loop's Notify: %s", e.Key)
		}
	}
}

// 3.R2a — park 원인 판정은 결과를 바꾸지 않음: 기록자가 실패해도(적재 실패) 같은 원장 행과 같은 제출.
func TestA094TheParkCheckChangesNoOutcome(t *testing.T) {
	run := func(fail bool) (journal.ExitState, int, int) {
		h, sub, crit := a094Harness(t, nil)
		if fail {
			crit.err = errors.New("outbox write failed")
		}
		p := a094ArmTakeProfit(t, h, sub, journal.StateUnresolvedInDoubt)
		h.quote("005930", 68500)
		h.observe()
		h.observe()
		return h.state(p.ID), len(h.submit.places), len(h.submit.cancels)
	}
	a, ap, ac := run(false)
	b, bp, bc := run(true)
	if a.PendingAction != b.PendingAction || a.Baseline != b.Baseline || a.HighWater != b.HighWater || ap != bp || ac != bc {
		t.Errorf("outcomes differ with a failing recorder: %+v/%d/%d vs %+v/%d/%d", a, ap, ac, b, bp, bc)
	}
}

// 3.B3 — 빈 가격(시장가)의 작업 주문은 치우기 실패가 아님.
func TestA094AnEmptyPriceIsNotAClearFailure(t *testing.T) {
	h, _, _ := a094Harness(t, nil)
	h.entry("005930", "10", "70000", "68000", "70000")
	orderID := h.workingEntry("005930", "5", "69500")
	// 시장가 모양: 원장의 intent 가격을 비움(가격 NULL — LiveOrdersForSymbol 은 빈 문자열로 읽음).
	if _, err := h.a094DB().Exec(`UPDATE intents SET price = NULL
		WHERE id = (SELECT intent_id FROM mutation_attempts WHERE broker_order_id = ?)`, orderID); err != nil {
		t.Fatalf("emptying the price: %v", err)
	}
	h.quote("005930", 67900)
	h.observe()
	if len(h.submit.cancels) != 1 || h.submit.cancels[0].Intent.OrderID != orderID {
		t.Fatalf("cancels = %+v, want the market order cancelled", h.submit.cancels)
	}
	if h.submit.cancels[0].Order.Price != 0 {
		t.Errorf("price = %v, want 0 for an empty price", h.submit.cancels[0].Order.Price)
	}
	if len(h.submit.places) != 1 {
		t.Errorf("places = %d, want the stop after the buy was cleared", len(h.submit.places))
	}
}

// 3.E2 — 이 경로의 브로커 호출은 가격 한 번뿐(원장 읽기만 더함).
func TestA094TheClearAddsNoBrokerRead(t *testing.T) {
	h, sub, _ := a094Harness(t, nil)
	a094ArmTakeProfit(t, h, sub, journal.StateUnresolvedInDoubt)
	before := h.prices.calls
	h.quote("005930", 68500)
	for i := 0; i < 3; i++ {
		h.observe()
	}
	if got := h.prices.calls - before; got != 3 {
		t.Errorf("price reads = %d over 3 cycles, want 3 — nothing else reads the broker", got)
	}
}

// a094RealGateway 는 하네스의 원장 위에 실제 게이트웨이를 세움 — 미종결 판정(UnsettledOnSymbol)을 실물로 부르기 위함.
// 브로커는 부르지 않음(발주 · 취소는 가짜 제출자가 함).
func a094RealGateway(t *testing.T, h *exitHarness) *execgw.Gateway {
	t.Helper()
	gw, err := execgw.New(execgw.Options{
		Journal: h.journal, Trading: trading.NewService(config.Trading{}, nil),
		Clock: h.clk, AccountRef: exitAccount, Source: "a094-test",
	})
	if err != nil {
		t.Fatalf("execgw.New: %v", err)
	}
	return gw
}

// a094DB 는 원장 파일을 직접 여는 창임 — 공개 API 가 만들 수 없는 모양(가격 없는 intent)을 픽스처로 만들 때만 씀.
func (h *exitHarness) a094DB() *sql.DB {
	h.t.Helper()
	db, err := sql.Open("sqlite", "file:"+h.journal.Path())
	if err != nil {
		h.t.Fatalf("opening %s: %v", h.dbPath, err)
	}
	h.t.Cleanup(func() { _ = db.Close() })
	return db
}

// --- ExitSubmitter 의 a094 메서드를 다른 시험의 가짜에 더함 -----------------------------------------------
// 그 시험들은 미종결 판정을 재지 않으므로 「미종결 없음」. 그 파일들을 편집하지 않고 여기에 두는 것은 FLM 게이트가 편집된
// 파일의 이웃 함수를 요구 집합으로 끌어들이기 때문임(저장소 교훈 「새 코드는 새 파일에」).

func (s *journalMutationSubmitter) UnsettledOnSymbol(context.Context, string, string) ([]journal.AttemptRecord, error) {
	return nil, nil
}

func (s *a111SubmitSpy) UnsettledOnSymbol(context.Context, string, string) ([]journal.AttemptRecord, error) {
	return nil, nil
}

func (s *signallingSubmitter) UnsettledOnSymbol(context.Context, string, string) ([]journal.AttemptRecord, error) {
	return nil, nil
}

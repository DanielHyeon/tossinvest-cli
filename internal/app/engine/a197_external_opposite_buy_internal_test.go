package engine

// a197 — 단계의 순수 판정 · 문장을 직접 잼: 열 조건(§4-8 · §4-9), 실행 직전 재검증(§4-15a, Manager 결정 M-3),
// 치우지 못한 취소 알림의 문장 · key(r1 #3 · gstack 리뷰 r2), 거절 알림 본문(D9).

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/exitpolicy"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
)

func TestA197TheStageOpensOnlyOnAProtectiveOppositePendingRefusal(t *testing.T) {
	stop := exitpolicy.Proposal{Action: exitpolicy.ActionBaselineBreach}
	ladderStop := exitpolicy.Proposal{Action: exitpolicy.ActionLadderStop}
	takeProfit := exitpolicy.Proposal{Action: exitpolicy.ActionLadderTakeProfit}
	opposite := execgw.Outcome{State: journal.StateFailedConfirmed, Reason: execgw.ReasonOppositePendingOrder}
	for name, tc := range map[string]struct {
		p    exitpolicy.Proposal
		out  execgw.Outcome
		want bool
	}{
		"stop refused for an opposite order":        {stop, opposite, true},
		"ladder stop refused for an opposite order": {ladderStop, opposite, true},
		"take profit refused for an opposite order": {takeProfit, opposite, false},
		"stop refused for another reason":           {stop, execgw.Outcome{State: journal.StateFailedConfirmed, Reason: execgw.ReasonCode(journal.ReasonDispatchRejected)}, false},
		"stop not dispatched with that reason":      {stop, execgw.Outcome{State: journal.StateNotDispatched, Reason: execgw.ReasonOppositePendingOrder}, false},
	} {
		if got := externalBuyOpens(tc.p, tc.out); got != tc.want {
			t.Errorf("%s: opens = %v, want %v", name, got, tc.want)
		}
	}
}

func TestA197TheItemMustStillBeTheSameOpenInstance(t *testing.T) {
	item := externalBuyItem{positionID: "p-1", instanceSeq: 2, market: "kr", symbol: "005930"}
	open := journal.Position{ID: "p-1", InstanceSeq: 2, State: "OPEN", Quantity: "10"}
	for name, tc := range map[string]struct {
		positions []journal.Position
		want      bool
	}{
		"same open instance": {[]journal.Position{open}, true},
		"closed":             {[]journal.Position{{ID: "p-1", InstanceSeq: 2, State: journal.PositionClosed, Quantity: "0"}}, false},
		"zero quantity":      {[]journal.Position{{ID: "p-1", InstanceSeq: 2, State: "OPEN", Quantity: "0"}}, false},
		"new instance":       {[]journal.Position{{ID: "p-1", InstanceSeq: 3, State: "OPEN", Quantity: "10"}}, false},
		"gone":               {nil, false},
		"another position":   {[]journal.Position{{ID: "p-2", InstanceSeq: 2, State: "OPEN", Quantity: "10"}}, false},
	} {
		if got := externalBuyItemCurrent(item, tc.positions); got != tc.want {
			t.Errorf("%s: current = %v, want %v", name, got, tc.want)
		}
	}
}

// a197InternalRecorder 는 알림 · critical 입구 둘을 함께 받는 기록기임(순수 문장 시험용).
type a197InternalRecorder struct{ events []obs.Event }

func (r *a197InternalRecorder) Notify(_ context.Context, e obs.Event) error {
	r.events = append(r.events, e)
	return nil
}

func (r *a197InternalRecorder) RecordCritical(ctx context.Context, e obs.Event, _ time.Duration) error {
	return r.Notify(ctx, e)
}

// gstack 리뷰 r2 — 기록됐지만 종결되지 않은 취소 attempt 만 「손절까지 막힘」임. IN_DOUBT 는 park 갈래가 먼저 받으므로 여기 표에서도 막힘 쪽.
func TestA197OnlyAnUnfinishedRecordedCancelHoldsTheSymbol(t *testing.T) {
	for name, tc := range map[string]struct {
		out  execgw.Outcome
		want bool
	}{
		"not recorded":              {execgw.Outcome{}, false},
		"recorded, write failed":    {execgw.Outcome{AttemptID: "a-1"}, true},
		"left dispatch started":     {execgw.Outcome{AttemptID: "a-1", State: journal.StateDispatchStarted}, true},
		"left acked":                {execgw.Outcome{AttemptID: "a-1", State: journal.StateAcked}, true},
		"refused for certain":       {execgw.Outcome{AttemptID: "a-1", State: journal.StateFailedConfirmed}, false},
		"refused before dispatch":   {execgw.Outcome{AttemptID: "a-1", State: journal.StateNotDispatched}, false},
		"confirmed with a late err": {execgw.Outcome{AttemptID: "a-1", State: journal.StateConfirmed}, false},
	} {
		if got := externalBuyCancelHeld(tc.out); got != tc.want {
			t.Errorf("%s: held = %v, want %v", name, got, tc.want)
		}
	}
}

// 치우지 못한 취소의 알림은 손절을 무엇이 푸는지를 사실대로 말함 — 기록 전 실패 · 종결되지 않은 attempt · 확정 거절 셋(r1 #3 · r2).
func TestA197TheUnclearedAlertSaysWhatReleasesTheStop(t *testing.T) {
	item := externalBuyItem{positionID: "p-1", instanceSeq: 1, market: "kr", symbol: "005930"}
	target := externalBuyRow{orderID: "X-1", buy: true}
	for name, tc := range map[string]struct {
		attemptID  string
		held       bool
		keySuffix  string
		wantText   string
		forbidText string
	}{
		"not sent":            {"", false, "|X-1|not-sent", "다시 시도할 수 있다", "재시작"},
		"held by the engine":  {"a-1", true, "|X-1", "재시작 복구", "취소해야 손절이 나간다"},
		"refused for certain": {"a-1", false, "|X-1", "취소해야 손절이 나간다", "재시작"},
	} {
		rec := &a197InternalRecorder{}
		o := &ExitObserver{opts: ExitObserverOptions{Critical: rec, NotificationsEnabled: true}}
		o.recordExternalBuyUncleared(context.Background(), item, target, tc.attemptID, tc.held, "synthetic")
		if len(rec.events) != 1 {
			t.Fatalf("%s: criticals = %d, want 1", name, len(rec.events))
		}
		e := rec.events[0]
		if e.Key != A197UnclearedKeyPrefix+"p-1"+tc.keySuffix {
			t.Errorf("%s: key = %q, want suffix %q", name, e.Key, tc.keySuffix)
		}
		if !strings.Contains(e.Body, tc.wantText) || strings.Contains(e.Body, tc.forbidText) {
			t.Errorf("%s: body = %q, want %q and not %q", name, e.Body, tc.wantText, tc.forbidText)
		}
	}
}

// D9 — 확정 거절은 사유와 무관하게 「브로커가 거절」로, 전송 전 거절만 「브로커에 닿기 전」으로 말함(토글 무관 문장).
func TestA197TheRefusalBodyNamesWhoRefused(t *testing.T) {
	m := managed{position: journal.Position{ID: "p-1", Symbol: "005930"}}
	stop := exitpolicy.Proposal{Action: exitpolicy.ActionBaselineBreach, Level: "L0"}
	for name, tc := range map[string]struct {
		state      journal.AttemptState
		reason     execgw.ReasonCode
		wantText   string
		forbidText string
	}{
		"opposite pending":  {journal.StateFailedConfirmed, execgw.ReasonOppositePendingOrder, "매수 미체결", "닿기 전에"},
		"another broker no": {journal.StateFailedConfirmed, execgw.ReasonCode(journal.ReasonDispatchRejected), "브로커가 거절했다", "닿기 전에"},
		"refused in-engine": {journal.StateNotDispatched, execgw.ReasonOppositePendingOrder, "닿기 전에", "브로커가 거절"},
	} {
		rec := &a197InternalRecorder{}
		o := &ExitObserver{opts: ExitObserverOptions{Alerts: rec}}
		o.alertProposalRefused(context.Background(), m, stop, "synthetic", tc.state, tc.reason)
		if len(rec.events) != 1 {
			t.Fatalf("%s: alerts = %d, want 1", name, len(rec.events))
		}
		if body := rec.events[0].Body; !strings.Contains(body, tc.wantText) || strings.Contains(body, tc.forbidText) {
			t.Errorf("%s: body = %q, want %q and not %q", name, body, tc.wantText, tc.forbidText)
		}
	}
}

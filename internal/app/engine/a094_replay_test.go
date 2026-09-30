package engine_test

// a094 §6 — 2026-08-06~07 사건의 재생(fixture). 기대 결말은 5판 이후 의미론이다(tasks 6.1 · 6.1a · 6.2):
// 손절이 나가는 것은 이 change 의 기대가 아니다(반대 매수는 엔진 밖 주문 — 3.X 사용자 결정). 기대는 「영구 동결」 이
// 「보고되는 반복」 또는 「명명된 보류」 로 바뀌고, 272210 의 PROPOSAL_CANCELLED 반복이 멈추는 것이다.

import (
	"context"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/app/engine"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/exitpolicy"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
)

// 6.1 — 새 409(원 발주 응답 · 목록 안 code)는 R1 으로 FAILED_CONFIRMED → 판정 함수가 풂 → 다음 관측이 다시 발의 → 반대
// 매수(엔진 밖)가 그대로라 같은 409 로 다시 거절되고 보고됨. 매 주기 거절 보고 · 발의 영구 무장 0.
// (게이트웨이 끝의 409 → FAILED_CONFIRMED 는 execgw 의 TestA094RefusalCodeClassifiesTheAttempt 가 실제 HTTP 로 잰다.)
func TestA094ReplayTheNew409BecomesAReportedRepetition(t *testing.T) {
	h, sub, _ := a094Harness(t, nil)
	p := h.entry("005930", "10", "70000", "68000", "70000")
	sub.placeState = journal.StateFailedConfirmed // 409 opposite-pending-order-exists 를 게이트웨이가 분류한 결과
	for i := 0; i < 3; i++ {
		h.quote("005930", 67000-float64(i*10))
		if c := h.observe(); c.Err != nil {
			t.Fatalf("cycle %d: %v", i, c.Err)
		}
		if h.state(p.ID).Pending() {
			t.Fatalf("cycle %d: the refused stop stayed armed — the freeze is back", i)
		}
	}
	if len(h.submit.places) != 3 {
		t.Errorf("places = %d, want one attempt per usable observation (the repetition is visible)", len(h.submit.places))
	}
	if got := h.exitEventCount(p.ID, journal.ExitEventProposalRefused); got != 3 {
		t.Errorf("PROPOSAL_REFUSED = %d, want 3", got)
	}
	if got := h.exitEventCount(p.ID, journal.ExitEventProposalCancelled); got != 0 {
		t.Errorf("PROPOSAL_CANCELLED = %d, want 0 (6.2 across all three branches)", got)
	}
	if got := h.alerts.count(obs.EventExitProposalRefused); got != 3 {
		t.Errorf("refusal alerts = %d, want each refusal reported", got)
	}
	if last := h.submit.places[len(h.submit.places)-1]; last.Intent.Side != "sell" {
		t.Errorf("last place %+v, want a sell", last.Intent)
	}
	_ = execgw.ReasonOppositePendingOrder
}

// 6.1a — 원장의 park 된 두 행(475150 · 080220)의 모양: 사다리 손절 발의(level 0 / -1)가 무장된 채 attempt 가 park.
// 발의는 무장된 채 남고 제출 0, park 원인 critical(무장 발의가 손절이라 청소가 아니라 판정 진입에서). 해동(비수용 종결 +
// 같은 판정) 뒤에야 6.1 과 같은 결말로 감.
func TestA094ReplayTheTwoParkedRows(t *testing.T) {
	for _, level := range []string{"0", "-1"} {
		t.Run("level "+level, func(t *testing.T) {
			h, _, crit := a094Harness(t, func(o *engine.ExitObserverOptions) {
				ladder := exitpolicy.DefaultLadderPolicy()
				o.Ladder = &ladder
			})
			ctx := context.Background()
			p := h.entry("005930", "10", "70000", "68000", "70000")
			if _, err := h.journal.OpenExitState(ctx, journal.ExitStateSeed{PositionID: p.ID, PolicyKind: journal.ExitPolicyLadder,
				EntryPrice: "70000", InitialStop: "68000"}); err != nil {
				t.Fatal(err)
			}
			h.quote("005930", 70100)
			h.observe()
			st := h.state(p.ID)
			if err := h.journal.RecordExitJudgement(ctx, journal.ExitJudgement{
				PositionID: p.ID, LifecycleGeneration: st.LifecycleGeneration, ObservedPrice: "67700",
				HighWater: st.HighWater, Baseline: st.Baseline, RatchetLevel: st.RatchetLevel, ActiveRung: st.ActiveRung,
				Proposal: &journal.ExitProposal{Action: string(exitpolicy.ActionLadderStop), Level: level, IntentID: "exit-park"},
			}); err != nil {
				t.Fatalf("arming the incident proposal: %v", err)
			}
			parked, _ := h.a094RecordSell("exit-park", 10, 67700, journal.StateUnresolvedInDoubt)
			for i := 0; i < 4; i++ {
				h.quote("005930", 67600-float64(i))
				h.observe()
			}
			if !h.state(p.ID).Pending() || len(h.submit.places) != 0 {
				t.Fatalf("armed %v places %d, want the parked proposal to hold and nothing submitted", h.state(p.ID).Pending(), len(h.submit.places))
			}
			if got := len(crit.withKey("|" + p.ID + "|attempt:" + parked)); got != 1 {
				t.Fatalf("park-cause alerts = %d, want 1", got)
			}
			if got := h.exitEventCount(p.ID, journal.ExitEventProposalCancelled); got != 0 {
				t.Fatalf("PROPOSAL_CANCELLED = %d on the parked branch, want 0", got)
			}
			// 해동: 운영자 비수용 종결 + 같은 판정 함수.
			if err := h.journal.OperatorResolve(ctx, parked, journal.StateFailedConfirmed, "op", "", "broker shows nothing"); err != nil {
				t.Fatal(err)
			}
			if _, released, err := h.journal.ReleaseUnacceptedExitProposal(ctx, p.ID, "exit-park", journal.ProposalRefused); err != nil || !released {
				t.Fatalf("thaw release = %v/%v", released, err)
			}
			h.quote("005930", 67500)
			h.observe()
			if len(h.submit.places) != 1 {
				t.Fatalf("places = %d after the thaw, want the stop proposed again", len(h.submit.places))
			}
		})
	}
}

// 6.2 — 272210 의 라이브락(STOP_LOSS → PROPOSAL_CANCELLED, 중앙값 5초): 무장 발의 없음 + 같은 종목에 **다른 intent 의**
// IN_DOUBT attempt(옛 409 로 얼어붙은 손절). 재생은 D−4.7 의 셋째 기전으로 그 반복을 재현한다 — 변이 M9(청소가 미종결을
// 안 봄)에서 이 시험이 PROPOSAL_CANCELLED 반복으로 실패하는 것이 인과 확정이다(변이 원장). 결말 불문 단언: 같은 포지션의
// PROPOSAL_CANCELLED 반복 기록 0.
func TestA094ReplayThe272210LivelockStops(t *testing.T) {
	h, sub, _ := a094Harness(t, nil)
	gw := a094RealGateway(t, h)
	sub.unsettled = gw.UnsettledOnSymbol
	// 게이트웨이가 그 제출을 SymbolInFlight 로 거절하는 모양(청소가 막지 못하면 여기까지 옴).
	sub.placeOutcome = &execgw.Outcome{Reason: execgw.ReasonSymbolInFlight, Detail: "attempt on 005930 has not settled yet"}
	p := h.entry("005930", "10", "70000", "68000", "70000")
	h.a094RecordSellOn("005930", "old-stop-409", journal.StateInDoubt)
	for i := 0; i < 36; i++ { // 3분 = 5초 × 36
		h.quote("005930", 67900-float64(i%3))
		h.observe()
		h.clk.Advance(5_000_000_000)
	}
	if got := h.exitEventCount(p.ID, journal.ExitEventProposalCancelled); got != 0 {
		t.Fatalf("PROPOSAL_CANCELLED = %d over 3 minutes, want 0 — the livelock is back", got)
	}
	if got := h.alerts.count(obs.EventExitLiquidationDelayed); got != 1 {
		t.Errorf("delay alerts = %d, want 1 — the held stop is now audible", got)
	}
}

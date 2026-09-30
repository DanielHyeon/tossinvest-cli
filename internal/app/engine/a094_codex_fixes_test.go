package engine_test

// a094 codex 구현 리뷰(i1) 수리의 시험 — F1(발의 해제 뒤 종결 대기 매도 위의 손절) · F2(판정 진입 해제 뒤 옛 상태 사본 평가).

import (
	"context"
	"fmt"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/app/engine"
	"github.com/JungHoonGhae/tossinvest-cli/internal/exitpolicy"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
)

// F1 — 청소가 취소한 매도(다른 intent)가 종결 기록을 기다리는 동안 무장 발의가 풀려도(withPending=false 주기) 손절은 그
// 매도의 종결 기록 전에 나가지 않음. 취소 전 체결량을 모르는 채 전량 손절을 내면 초과 매도가 됨.
func TestA094ACancelledSellStillHoldsTheStopAfterTheProposalIsReleased(t *testing.T) {
	h, sub, _ := a094Harness(t, nil)
	p := h.entry("005930", "10", "70000", "68000", "70000")
	h.quote("005930", 70100)
	h.observe()
	h.a094ArmOn(p, "005930", "exit-x", string(exitpolicy.ActionRatchetPartial)) // attempt 없음 — 판정 진입이 풀지 않음
	_, sellOrder := h.a094RecordSell("old-sell", 4, 72000, journal.StateConfirmed)
	h.submit.settle = nil
	sub.cancelRecords = true
	h.quote("005930", 67900)
	h.observe() // 발의가 있어 청소가 매도를 취소 — 종결 대기
	if len(h.submit.cancels) != 1 || len(h.submit.places) != 0 {
		t.Fatalf("control: cancels %d places %d", len(h.submit.cancels), len(h.submit.places))
	}
	// 발의 intent 가 비수용으로 끝남 → 판정 진입이 풂.
	h.a094RecordSell("exit-x", 4, 72000, journal.StateFailedConfirmed)
	h.observe()
	if h.state(p.ID).Pending() {
		t.Fatal("control: the unaccepted proposal was not released")
	}
	for i := 0; i < 3; i++ {
		h.quote("005930", 67800-float64(i))
		h.observe()
	}
	if len(h.submit.places) != 0 {
		t.Fatalf("places = %d — a stop went out over a cancelled sell whose fills are unknown", len(h.submit.places))
	}
	if len(h.submit.cancels) != 1 {
		t.Errorf("cancels = %d, want the acknowledged cancel not re-sent", len(h.submit.cancels))
	}
	h.settleCancelled(sellOrder)
	h.observe()
	if len(h.submit.places) != 1 {
		t.Fatalf("places = %d after the closing record, want the stop", len(h.submit.places))
	}
}

// F2 — 판정 진입이 입증된 비수용 발의(사다리 익절)를 풀면 그 주기의 판정을 건너뛰어, 되돌린 rung 을 옛 사본으로 다시 쓰지
// 않음 — 익절이 같은 rung 으로 다시 발의됨.
func TestA094AReleasedLadderRungIsProposedAgain(t *testing.T) {
	h, sub, _ := a094Harness(t, func(o *engine.ExitObserverOptions) {
		ladder := exitpolicy.DefaultLadderPolicy()
		o.Ladder = &ladder
	})
	ctx := context.Background()
	p := h.entry("005930", "10", "70000", "68000", "70000")
	if _, err := h.journal.OpenExitState(ctx, journal.ExitStateSeed{PositionID: p.ID, PolicyKind: journal.ExitPolicyLadder,
		EntryPrice: "70000", InitialStop: "68000"}); err != nil {
		t.Fatal(err)
	}
	h.quote("005930", 70500)
	h.observe()
	h.quote("005930", 71100) // rung 1(1.5%) — 상태만 승격(분할 0)
	h.observe()
	// 둘째 rung(2.5%, 분할 25%) 교차 — 익절 발의 · 제출이 확정 거절로 오는데 해제 쓰기가 실패(발의 · rung 이 남음).
	h.a094Exec(t, fmt.Sprintf(`CREATE TRIGGER a094_refuse_release BEFORE UPDATE ON exit_states
		WHEN OLD.position_id = '%s' AND OLD.pending_action IS NOT NULL AND NEW.pending_action IS NULL
		BEGIN SELECT RAISE(ABORT, 'a094 fixture'); END`, p.ID))
	sub.placeState = journal.StateFailedConfirmed
	h.quote("005930", 71800)
	h.observe()
	armed := h.state(p.ID)
	if !armed.Pending() || len(h.submit.places) != 1 {
		t.Fatalf("control: pending %v places %d — the rung was not proposed", armed.Pending(), len(h.submit.places))
	}
	h.a094Exec(t, `DROP TRIGGER a094_refuse_release`)
	sub.placeState = ""
	h.observe() // 판정 진입이 풀고 이 주기 판정은 건너뜀
	if st := h.state(p.ID); st.Pending() || st.ActiveRung == armed.ActiveRung {
		t.Fatalf("after the release: pending %v rung %d (was %d) — the rung must be rolled back", st.Pending(), st.ActiveRung, armed.ActiveRung)
	}
	h.observe()
	if len(h.submit.places) != 2 {
		t.Fatalf("places = %d, want the same rung proposed again", len(h.submit.places))
	}
}

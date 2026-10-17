package strategyhandoff

// a112 태스크 5.2.2.1 — 소유자 범위별 Admit 경계. 동결 골든의 queue 블록은 "market_wide_single_proposal_assumption_forbidden":
// true · "selected_limit": "at most one selected proposal per owner scope" 다. 활성화된 시장의 주문 경로는 시장 하나가 아니라
// **소유자 범위 하나**마다 handoff 를 받는다: 각 handoff 는 선택 하나를 싣고 기존 단일 값 문(Single · Deliver)과 그 불변식
// (Single 이 값을 내주는 것 ⇔ Refusal 이 Admitted)을 그대로 지킨다. 시장 단위 거절(닫힘 · 선택 없음)은 handoff 하나로,
// 같은 소유자 범위에 둘이 실리면(범위당 상한 초과) 시장 전체를 거절한다 — 하나를 골라 조용히 버리지 않는다.

import (
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyflow"
)

func scopedSelection(identity, symbol string, generation uint64) strategyflow.Result {
	return strategyflow.Result{Lineage: strategyflow.Lineage{Identity: identity, AccountRef: "acct", Market: "KR",
		Symbol: symbol, PositionGeneration: generation}}
}

func TestEachOwnerScopeCrossesTheSeamOnItsOwn(t *testing.T) {
	selected := []strategyflow.Result{scopedSelection("a", "005930", 1), scopedSelection("b", "000660", 1)}
	handoffs := AdmitEachOwnerScope(true, selected)
	if len(handoffs) != 2 {
		t.Fatalf("handoffs=%d, want one per owner scope", len(handoffs))
	}
	for i, handoff := range handoffs {
		result, ok := handoff.Single()
		if !ok || handoff.Refusal() != Admitted || result.Lineage.Identity != selected[i].Lineage.Identity {
			t.Fatalf("scope %d: ok=%v refusal=%q identity=%q, want %q admitted in coordinator order",
				i, ok, handoff.Refusal(), result.Lineage.Identity, selected[i].Lineage.Identity)
		}
		if handoff.Pending() != 1 {
			t.Errorf("scope %d pending=%d, want its own single scope", i, handoff.Pending())
		}
	}
	// 부르는 쪽 배열을 들고 있지 않는다.
	selected[0] = scopedSelection("mutated", "005930", 1)
	if result, _ := handoffs[0].Single(); result.Lineage.Identity != "a" {
		t.Fatalf("the admitted scope aliases the caller's slice: %q", result.Lineage.Identity)
	}
}

func TestMarketLevelRefusalsStayOneHandoff(t *testing.T) {
	two := []strategyflow.Result{scopedSelection("a", "005930", 1), scopedSelection("b", "000660", 1)}
	for name, tc := range map[string]struct {
		ready    bool
		selected []strategyflow.Result
		want     Refusal
		pending  int
	}{
		"closed":    {false, two, MarketClosed, 2},
		"empty":     {true, nil, NoSelection, 0},
		"duplicate": {true, []strategyflow.Result{scopedSelection("a", "005930", 1), scopedSelection("b", "005930", 1)}, OverCapacity, 2},
		// 표기만 다른 같은 범위(strategyrouter.NewOwnerKey 의 정규화 — 종목 대문자 · 공백 제거)는 둘로 세지 않는다.
		"duplicate symbol spelling": {true, []strategyflow.Result{scopedSelection("a", "aapl", 1), scopedSelection("b", " AAPL", 1)}, OverCapacity, 2},
		"duplicate account spelling": {true, []strategyflow.Result{scopedSelection("a", "005930", 1),
			{Lineage: strategyflow.Lineage{Identity: "b", AccountRef: " acct ", Market: "KR", Symbol: "005930", PositionGeneration: 1}}}, OverCapacity, 2},
	} {
		handoffs := AdmitEachOwnerScope(tc.ready, tc.selected)
		if len(handoffs) != 1 {
			t.Fatalf("%s: handoffs=%d, want a single market-level refusal", name, len(handoffs))
		}
		if _, ok := handoffs[0].Single(); ok || handoffs[0].Refusal() != tc.want || handoffs[0].Pending() != tc.pending {
			t.Fatalf("%s: ok=%v refusal=%q pending=%d, want refusal %q pending %d",
				name, ok, handoffs[0].Refusal(), handoffs[0].Pending(), tc.want, tc.pending)
		}
	}
}

// 소유자 범위는 종목만이 아니다: 같은 종목이라도 포지션 세대가 다르면 다른 범위다(strategyrouter.OwnerKey).
func TestAnotherPositionGenerationIsAnotherOwnerScope(t *testing.T) {
	handoffs := AdmitEachOwnerScope(true, []strategyflow.Result{scopedSelection("a", "005930", 1), scopedSelection("b", "005930", 2)})
	if len(handoffs) != 2 || handoffs[0].Refusal() != Admitted || handoffs[1].Refusal() != Admitted {
		t.Fatalf("two generations of one symbol were not two scopes: %d handoffs", len(handoffs))
	}
}

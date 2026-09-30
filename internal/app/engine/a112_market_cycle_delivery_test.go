//go:build tossos_testseams

package engine

// a112 5.2.2.1 리뷰 수리 3차(codex 재확인 #1, Manager 판정 2026-10-01) — 주기 함수 `runProductionStrategyMarketCycle` **자체**를
// 돌려 「dispatch 가 이 시장의 handoff 를 전부 받았다」를 잰다.
//
// 앞 두 판의 구조 못은 철자(이름 → 식별자 해소)를 조였지만 매번 우회가 남았다 — 호출부에서 `clear(hs[1:])` 로 뒤 handoff 를
// 지워도 통과했다(codex 재확인 #1). 「축별 행동 시험은 종료하지 않는다」의 결론대로 여기서는 모양이 아니라 **의미**를 잰다:
// 주기 함수를 실제로 부르고, dispatch 가 몇 번 불렸는지 Gateway 스파이의 보호 관측 수로 센다. 어떤 철자의 우회든 handoff 하나를
// 빼면 이 수가 준다.
//
// 조립: 권한 새로 고침은 Context 의 1초 캐시(`strategyRefresh` · `strategyRefreshAt`)에 fixture 조립을 넣어 주입하고(원격 0),
// 레인 런타임은 생산 생성자가 fixture 원장 위에 세운다. dispatch 는 fixture 의 실제 dispatch 주기(Gateway 는 스파이 — 실주문 아님).
//
// 시나리오: 서명 활성화된 KR 시장, 소유자 범위 둘(원래 005930 먼저 · 000660 뒤). 1차 레그 권한은 **원래 범위 하나**만 안다 —
// 그래서 첫 범위는 끝까지 가서 스파이 주문 1 건, 둘째 범위는 1차 레그 identity 대조(:221)에서 거절된다. 두 범위 모두 dispatch 에
// 들어갔다는 사실이 보호 관측 수 2 로 드러나고, 둘째 범위의 거절이 주기의 반환 오류로 드러난다. 「첫 오류에서 멈춤」 때문에
// 순서를 거꾸로 하면 둘째가 안 불린다 — 그래서 이 순서다.
//
// 이 시험은 5.6.2.2 의 행동 커버 요구 중 **주기 전달 부분**의 선납이다(tasks 5.6.2.2).

import (
	"context"
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyrouter"
)

func TestTheProductionCycleHandsEveryOwnerScopeToTheDispatch(t *testing.T) {
	cycle, proposals, j, spy := pairedStrategyDispatchCycleFixture(t)
	loader, ok := cycle.firstLeg.loader.(*productionStrategyFirstLegAuthorityLoader)
	if !ok {
		t.Fatalf("arrangement: first-leg loader is %T", cycle.firstLeg.loader)
	}
	activation := strategyrouter.FamilyActivationForTest(strategyrouter.MarketKR, 1,
		strategyrouter.AllFourFamiliesForTest(strategyrouter.MarketKR))
	two := a112TwoScopeKR(t, proposals.kr, proposals.observedAt)
	two.activation = activation
	// dispatch 주기는 두 범위를 본다(활성화 · 보호 하한 판정). 1차 레그 권한은 원래 범위 하나만 안다(위 시나리오).
	cycle.proposals.kr = two
	single := loader.proposals.kr
	single.activation = activation
	loader.proposals.kr = single

	pair := proposals
	pair.kr = two
	now := loader.clk.Now()
	c := &Context{Journal: j, AccountRef: "acct-risk-loader"}
	c.strategyRefresh = &StrategyEntryProductionAssembly{dispatch: cycle, proposals: pair, schedule: cycle.schedule}
	c.strategyRefreshAt = now

	err := c.runProductionStrategyMarketCycle(context.Background(), loader.clk, StrategyMarketKR)

	spy.mu.Lock()
	dispatched, placed := spy.observed["protection-kr"], len(spy.calls)
	spy.mu.Unlock()
	if dispatched != 2 {
		t.Fatalf("the cycle handed %d owner-scope handoff(s) to the dispatch, want 2 (err=%v)", dispatched, err)
	}
	if placed != 1 {
		t.Fatalf("gateway spy places=%d, want 1 (the scope the first-leg authority knows)", placed)
	}
	if err == nil || !strings.Contains(err.Error(), "production proposal identity changed") {
		t.Fatalf("cycle err=%v, want the second scope refused by the first-leg identity check", err)
	}
}

// 대조: 활성화 없는 같은 두 범위 시장(오늘 생산의 갈래)은 주기 함수를 통째로 돌아도 dispatch 에 아무것도 넘기지 않는다
// (시장 단위 OverCapacity — Deliver 가 몸통을 부르지 않음). 토글 OFF = upstream 을 주기 함수 수준에서 잰다.
func TestWithoutAnActivationTheProductionCycleHandsNothingToTheDispatch(t *testing.T) {
	cycle, proposals, j, spy := pairedStrategyDispatchCycleFixture(t)
	loader, ok := cycle.firstLeg.loader.(*productionStrategyFirstLegAuthorityLoader)
	if !ok {
		t.Fatalf("arrangement: first-leg loader is %T", cycle.firstLeg.loader)
	}
	two := a112TwoScopeKR(t, proposals.kr, proposals.observedAt)
	cycle.proposals.kr = two
	pair := proposals
	pair.kr = two
	c := &Context{Journal: j, AccountRef: "acct-risk-loader"}
	c.strategyRefresh = &StrategyEntryProductionAssembly{dispatch: cycle, proposals: pair, schedule: cycle.schedule}
	c.strategyRefreshAt = loader.clk.Now()

	if err := c.runProductionStrategyMarketCycle(context.Background(), loader.clk, StrategyMarketKR); err != nil {
		t.Fatalf("cycle err=%v, want nil (a refused handoff is not an error)", err)
	}
	spy.mu.Lock()
	defer spy.mu.Unlock()
	if spy.observed["protection-kr"] != 0 || len(spy.calls) != 0 {
		t.Fatalf("an unactivated two-scope market reached the dispatch: observed=%v places=%d", spy.observed, len(spy.calls))
	}
}

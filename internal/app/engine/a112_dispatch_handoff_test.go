//go:build tossos_testseams

package engine

import (
	"context"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/continuationlane"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyhandoff"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyrouter"
)

// 여기 있는 시험은 경계의 **규칙**을 다시 검사하지 않는다. 그 규칙과 그
// fail-closed 성질은 internal/strategyhandoff 의 시험이 값으로 지킨다.
// 이 파일이 지키는 것은 **배선**이다: 엔진의 레인 권한이 실제로 그 경계에
// 무엇을 건네고, 그 답이 승격과 dispatch 에 어떻게 닿는가.

// 종목이 둘인 시장은 지금도 아무것도 내보내지 않는다. 달라지는 것은 그 이유를
// 말한다는 점이다.
//
// 이 시장은 고장난 시장이 아니다. 조정자가 소유자 범위마다 정확히 하나씩
// 골라 둘을 돌려준, 계약대로 동작한 결과다. 그런데도 공유 dispatch 는
// 아무것도 받지 못한다. 그 사실이 어디에도 적히지 않으면 운영자는 "오늘
// 후보가 없었나 보다"와 구별할 수 없다.
func TestAMarketWithTwoSelectedScopesNamesWhyNothingWasHandedOff(t *testing.T) {
	now := time.Date(2026, 8, 29, 1, 2, 3, 0, time.UTC)
	pair := collectArbitrated(t, now, familyScoresForTest(strategyrouter.MarketKR), "005930",
		continuationlane.KRContinuationLaneID)
	if !pair.kr.snapshot.Ready || len(pair.kr.entries) != 2 {
		t.Fatalf("KR=%+v entries=%d, want a ready market with two selected scopes", pair.kr.snapshot, len(pair.kr.entries))
	}
	handoff := pair.kr.dispatchHandoff()
	result, handedOff := handoff.Single()
	if handedOff {
		t.Fatalf("handoff admitted %q from a two-scope market", result.Lineage.Symbol)
	}
	if handoff.Refusal() != strategyhandoff.OverCapacity {
		t.Fatalf("refusal=%q, want %q", handoff.Refusal(), strategyhandoff.OverCapacity)
	}
	if handoff.Pending() != 2 {
		t.Fatalf("pending=%d, want 2 — 몇 개라서 막혔는지 말하지 못하면 상한이 보이지 않는다", handoff.Pending())
	}
	if result.Lineage.Symbol != "" {
		t.Fatalf("a refused handoff still carries %q", result.Lineage.Symbol)
	}
}

// 조정자가 하나만 고른 시장은 그 하나가 그대로 건너간다.
func TestASingleSelectedScopeIsTheValueThatCrossesTheSeam(t *testing.T) {
	_, proposals, _, _ := pairedStrategyDispatchCycleFixture(t)
	for _, market := range []StrategyMarket{StrategyMarketKR, StrategyMarketUS} {
		authority := proposals.forMarket(market)
		if len(authority.entries) != 1 {
			t.Fatalf("%s fixture entries=%d, want one", market, len(authority.entries))
		}
		handoff := authority.dispatchHandoff()
		result, handedOff := handoff.Single()
		if !handedOff || handoff.Pending() != 1 {
			t.Fatalf("%s handoff refusal=%q pending=%d, want an admitted single selection",
				market, handoff.Refusal(), handoff.Pending())
		}
		want := authority.entries[0].authority.Proposal()
		if result.Lineage.Identity != want.Lineage.Identity {
			t.Fatalf("%s handed off %q, want %q", market, result.Lineage.Identity, want.Lineage.Identity)
		}
	}
}

// 닫힌 시장에서는 항목이 남아 있어도 아무것도 건너가지 않는다.
//
// 이 상태는 지금 생산 코드가 만들지 않는다 —
// TestEveryMarketAuthorityCarryingEntriesAlsoReportsReady 가 그것을 지킨다.
// 그래도 검사를 두고 시험하는 이유는, 그 불변식이 깨지는 날 조용히 통과하는
// 것보다 여기서 막히는 편이 낫기 때문이다.
func TestAClosedMarketHandsOffNothingEvenWhenAnEntryIsStillAttached(t *testing.T) {
	_, proposals, _, _ := pairedStrategyDispatchCycleFixture(t)
	closed := proposals.kr
	closed.snapshot.Ready = false
	closed.snapshot.Reason = StrategyProposalArbitrationRefused
	handoff := closed.dispatchHandoff()
	if result, handedOff := handoff.Single(); handedOff {
		t.Fatalf("a closed market handed off %q", result.Lineage.Symbol)
	}
	if handoff.Refusal() != strategyhandoff.MarketClosed {
		t.Fatalf("refusal=%q, want %q", handoff.Refusal(), strategyhandoff.MarketClosed)
	}
	if handoff.Pending() != 1 {
		t.Fatalf("pending=%d, want the attached entry counted", handoff.Pending())
	}
}

// 고른 것이 없는 시장은 "없음"이라고 말한다. 상한에 걸린 것과 같은 이름을
// 쓰면 후보가 없던 날과 상한이 막은 날이 한 글자로 뭉쳐진다.
func TestAMarketThatSelectedNothingSaysSoInsteadOfBorrowingTheCapacityName(t *testing.T) {
	handoff := strategyProposalMarketAuthority{market: StrategyMarketKR,
		snapshot: StrategyProposalMarketSnapshot{Market: StrategyMarketKR, Ready: true, Reason: StrategyProposalReady}}.dispatchHandoff()
	if _, handedOff := handoff.Single(); handedOff {
		t.Fatal("an empty market handed something off")
	}
	if handoff.Refusal() != strategyhandoff.NoSelection || handoff.Pending() != 0 {
		t.Fatalf("refusal=%q pending=%d, want an empty-selection refusal", handoff.Refusal(), handoff.Pending())
	}
}

// 경계가 거절하면 worker 는 Effective 로 올라가지 않는다.
//
// 이것이 이 태스크가 동작을 바꾸지 않았다는 증거다. 상한을 한 자리로 모은
// 것이지 푼 것이 아니다 — 푸는 것은 태스크 5.2 다.
func TestARefusedHandoffLeavesTheWorkerDormant(t *testing.T) {
	cycle, proposals, _, spy := pairedStrategyDispatchCycleFixture(t)
	loader, ok := cycle.firstLeg.loader.(*productionStrategyFirstLegAuthorityLoader)
	if !ok {
		t.Fatal("production first-leg authority loader unavailable")
	}
	now := loader.schedule.observedAt
	candidates := strategyCandidateAuthorityPair{observedAt: now,
		kr: readyCandidateAuthority(StrategyMarketKR), us: readyCandidateAuthority(StrategyMarketUS)}
	routes := strategyRouteAuthorityPair{observedAt: now,
		kr: readyRouteAuthority(StrategyMarketKR), us: readyRouteAuthority(StrategyMarketUS)}
	cycleFn := func(context.Context) error { return nil }
	build := func(pair strategyProposalAuthorityPair, market StrategyMarket) StrategyMarketWorker {
		return buildProductionStrategyMarketWorker(context.Background(), loader.clk, market, true, spy,
			loader.schedule, candidates, routes, loader.fx, pair, loader.risk, loader.accounts, cycleFn)
	}
	if worker := build(proposals, StrategyMarketKR); !worker.Effective {
		t.Fatalf("baseline KR worker=%+v, want Effective before the capacity refusal", worker)
	}
	// 같은 항목을 하나 더 붙여 상한을 넘긴다. 다른 것은 아무것도 바꾸지 않는다.
	overCapacity := proposals
	overCapacity.kr.entries = append(append([]strategyProposalEntryAuthority{}, proposals.kr.entries...), proposals.kr.entries[0])
	if refusal := overCapacity.kr.dispatchHandoff().Refusal(); refusal != strategyhandoff.OverCapacity {
		t.Fatalf("refusal=%q, want the capacity refusal this test rests on", refusal)
	}
	if worker := build(overCapacity, StrategyMarketKR); worker.Effective {
		t.Fatalf("KR worker=%+v went Effective while the handoff refused", worker)
	}
	if worker := build(overCapacity, StrategyMarketUS); !worker.Effective {
		t.Fatalf("US worker=%+v lost its promotion to a KR-local capacity refusal", worker)
	}
}

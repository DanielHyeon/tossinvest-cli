package engine

import (
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyarbiter"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyproposal"
)

// arbitrateProposalScope 는 한 종목의 모든 가족 제안을 보정 중재자에게 넘긴다.
//
// 여기서는 아무것도 고르지 않는다. 고르는 규칙은 strategyarbiter 한 곳에만 두고,
// 이 함수는 그 규칙에 필요한 값을 모아 주는 일만 한다. 규칙이 두 곳에 있으면
// 언젠가 두 곳이 서로 다른 답을 낸다.
//
// 기대 범위의 종목은 경로 권한이 아니라 승인된 후보에서 읽는다. 권한이 스스로
// 말한 종목을 그 권한을 검사하는 데 다시 쓰면, 어긋남을 잡아내려던 검사가
// 언제나 참이 되어 아무것도 잡지 못한다.
func arbitrateProposalScope(accountRef string, market StrategyMarket, route strategyRouteEntryAuthority,
	lanes []strategyproposal.ProductionAuthority, observedAt time.Time,
) strategyarbiter.Outcome {
	proposals := make([]strategyarbiter.Proposal, 0, len(lanes))
	for _, lane := range lanes {
		proposals = append(proposals, strategyarbiter.Proposal{Result: lane.Proposal(), Authority: route.route})
	}
	return strategyarbiter.Arbitrate(strategyarbiter.Request{AccountRef: accountRef, Market: strategyRouterMarket(market),
		Symbol: route.approved.Symbol(), PositionGeneration: route.route.Request().Key.PositionGeneration,
		ObservedAt: observedAt, Proposals: proposals})
}

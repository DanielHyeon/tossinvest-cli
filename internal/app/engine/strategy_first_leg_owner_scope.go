package engine

import (
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyflow"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyproposal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyrouter"
)

// strategyFirstLegOwnerScopeRefusal 은 1차 레그 권한이 건너온 결과의 소유자 범위를 조립의 권한 쌍에서 **정확히 하나로** 찾지
// 못했을 때의 거절 문구임. identity 거절 문구를 머리로 담아, 봉투가 조립이 중재하지 않은 것을 날랐다는 같은 사실을 같은 머리로
// 부름(뒤의 설명이 어느 판정이 거절했는지 가름).
const strategyFirstLegOwnerScopeRefusal = "production proposal identity changed: owner scope is not uniquely authorized by the assembly"

// authorityForOwnerScope 는 a112 6.2 봉인의 **선택** 절반임(Manager 판정 2026-10-01 — 안 (C) 「의미 봉인」).
//
// 1차 레그 권한은 건너온 봉투를 믿지 않고, 조립이 새로 고침 때 중재해 둔 이 권한 쌍에서 항목을 다시 꺼내 봉인된 identity 를 대조함
// (collectStrategyFirstLegAuthority 의 identity 가드). 여기서는 그 항목을 **소유자 범위**(계좌 · 시장 · 종목 · 포지션 세대 —
// strategyrouter.OwnerKey 정규형)로 고름:
//   - 범위는 건너온 결과의 계보에서 읽지만 **선택 기준일 뿐 대조 대상이 아님.** identity 로 고르면(accepted 와 맞는 항목을 골라
//     accepted 와 비교하면) 대조가 공허해짐 — HANDOFF §4 의 자기 참조 함정. 범위로 고르면 같은 범위의 다른 가족 · 다른 캠페인 ·
//     조건 재작성은 선택되지만 identity 대조에서 거절됨.
//   - 정확히 하나만 받음. 없으면(조립이 중재하지 않은 범위 · 다른 시장) 거절, 둘 이상이면(조정자의 범위당 하나 보장이 깨짐)
//     아무것도 고르지 않고 거절 — 하나를 골라 조용히 버리지 않음.
//   - 항목의 계보가 OwnerKey 로 정규화되지 않으면 그 쌍 전체를 믿지 않고 거절(보수 방향).
//
// 왜 조정자 토큰이 아니라 이것인가: 엔진이 조정자를 만들고 · 먹이고 · 돌리므로 조정자가 찍은 값은 「어떤 중재가 돌았다」만
// 증명함. 위조가 주문이 되는 유일한 자리가 1차 레그 권한이고, 여기서 **조립의 중재 결과**와 대조하면 봉투가 어떤 철자로
// 만들어졌든 의미로 잡힘.
func (authority strategyProposalMarketAuthority) authorityForOwnerScope(lineage strategyflow.Lineage) (strategyproposal.ProductionAuthority, bool) {
	want, err := strategyrouter.NewOwnerKey(lineage.AccountRef, lineage.Market, lineage.Symbol, lineage.PositionGeneration)
	if err != nil {
		return strategyproposal.ProductionAuthority{}, false
	}
	var chosen strategyproposal.ProductionAuthority
	matches := 0
	for _, entry := range authority.entries {
		held := entry.authority.Proposal().Lineage
		key, err := strategyrouter.NewOwnerKey(held.AccountRef, held.Market, held.Symbol, held.PositionGeneration)
		if err != nil {
			return strategyproposal.ProductionAuthority{}, false
		}
		if key == want {
			chosen = entry.authority
			matches++
		}
	}
	if matches != 1 {
		return strategyproposal.ProductionAuthority{}, false
	}
	return chosen, true
}

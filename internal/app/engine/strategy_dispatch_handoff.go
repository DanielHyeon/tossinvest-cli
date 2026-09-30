package engine

import (
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyflow"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyhandoff"
)

// dispatchHandoff 는 한 시장의 조정 결과를 공유 dispatch 로 건너가는 경계
// 값으로 바꾼다.
//
// 판단은 여기서 하지 않는다. 상한도 거절 이름도 strategyhandoff 한 곳에 있고,
// 그 패키지는 엔진 밖에 있어서 "여기서는 주문을 낼 수 없다"를 import 목록으로
// 증명할 수 있다. 이 함수가 하는 일은 엔진의 레인 권한에서 봉인된 제안을 꺼내
// 조정자가 정한 순서 그대로 건네주는 것뿐이다.
//
// Proposal 은 값을 읽기만 하는 접근자다(strategyproposal/production.go:131).
// 그래서 거절될 시장의 것까지 미리 꺼내도 부작용이 없다.
func (authority strategyProposalMarketAuthority) dispatchHandoff() strategyhandoff.Handoff {
	selected := make([]strategyflow.Result, 0, len(authority.entries))
	for _, entry := range authority.entries {
		selected = append(selected, entry.authority.Proposal())
	}
	return strategyhandoff.Admit(authority.snapshot.Ready, selected)
}

// dispatchHandoffs 는 주문 경로(runProductionStrategyMarketCycle)가 이 시장에서 받는 handoff 들임(태스크 5.2.2.1).
//
//   - 서명 활성화가 **없는** 시장(오늘 생산 전부 — 배포된 서명 매니페스트 0건): dispatchHandoff 하나 그대로. 시장 단위
//     상한(Capacity=1)과 그 거절 이름이 바뀌지 않으므로 토글 OFF = upstream 동작 불변.
//   - 서명 활성화가 **있는** 시장: 소유자 범위마다 handoff 하나(strategyhandoff.AdmitEachOwnerScope). 상한이 시장 단위에서
//     소유자 범위 단위로 오름. 활성화 판정은 dispatch 주기의 보호 세대 하한(strategy_dispatch_cycle.go 의
//     `familyActivation().Verified()`)과 같은 값을 읽음 — 두 자리가 다른 활성화를 보지 않게 함.
//
// 하류(결과 권한 · 위험 · 계좌 · 1차 레그 권한 · worker 승격 · projection)도 이 목록을 소유자 범위 단위로 읽음(5.2.2.2 — 시장
// 단위 개수 관문 제거, 범위별 재유도). 한 범위의 위험 · 계좌 권한이 없으면 그 범위만 타입 거절되고(strategyScopeRefusal) 나머지
// 범위는 계속 감. 같은 파도의 둘째 범위는 공유 버킷 사용량 CAS(BUCKET_USAGE_STALE)가 거절하고 다음 파도에 발급됨 — 설계된 보호.
func (authority strategyProposalMarketAuthority) dispatchHandoffs() []strategyhandoff.Handoff {
	if !authority.familyActivation().Verified() {
		return []strategyhandoff.Handoff{authority.dispatchHandoff()}
	}
	// 조정자가 정한 순서 그대로 꺼냄(dispatchHandoff 와 같은 읽기 — Proposal 은 부작용 없는 접근자).
	selected := make([]strategyflow.Result, 0, len(authority.entries))
	for _, entry := range authority.entries {
		selected = append(selected, entry.authority.Proposal())
	}
	// 가독 계약(a112 6.2 A-lite — 2차 방어, **봉인이 아님**): 건네는 목록이 조립이 중재 때 적어 둔 제안 집합 digest 와 같아야 함.
	// 다르면 준비 안 됨으로 닫음(범위를 쪼개지 않고 시장 단위 MarketClosed 하나 — 경계에 없는 거절 이름을 지어내지 않음).
	// 이것이 **못 보는 모양**: 엔진 코드는 digest 도 스스로 계산할 수 있으므로(strategyProposalSetDigest 는 엔진 함수) 권한 값을
	// 통째로 위조하면 이 대조를 맞출 수 있고, digest 가 (종목, 계보 identity)만 담으므로 같은 계보의 조건 재작성은 보지 못한다. 봉인은
	// 1차 레그 권한의 소유자 범위 재유도(authorityForOwnerScope + identity 가드)가 진다 — 그 권한은 구성 때 제안 쌍을 떼어 내므로
	// (detachedStrategyProposalPair) 이 목록의 재할당도 제자리 원소 교체도 그 대조 원본에 닿지 않는다(6.2 리뷰 A#2 · codex #1 뒤 정정).
	// 여기는 「중재 결과와 다른 목록」이 섞이는 것을 이름 붙여 막는 읽히는 계약이다.
	ready := authority.snapshot.Ready && strategyProposalSetDigest(authority.entries) == authority.snapshot.ProposalSetDigest
	return strategyhandoff.AdmitEachOwnerScope(ready, selected)
}

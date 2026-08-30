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

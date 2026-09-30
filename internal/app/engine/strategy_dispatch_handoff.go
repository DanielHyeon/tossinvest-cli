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
// 이 함수가 올리는 것은 **경계의 상한뿐**임. 하류의 세 권한(결과 권한 · 1차 레그 권한 · worker 승격/projection)은 여전히
// 시장당 제안 하나를 요구하고, 1차 레그의 다섯 줄(strategy_account_first_leg_authority.go)은 결정 (1) 에 따라 L6 6.2 봉인
// 전까지 바꾸지 않음. 그래서 오늘 소유자 범위가 둘인 활성화 시장은 경계를 지나도 하류에서 거절돼 주문이 0 임 — 생산에서는
// 결과 권한 · 계좌 B1 · 위험 권한 재수집도 거절하고, 1차 레그 B2 가 **유일한** 방어인 것은 오늘-동등성 핀의 의도적 최악
// 조건(fixture 순서)에서뿐임(a112_owner_scope_handoff_test.go 머리말). 하류를 소유자 범위 단위로 옮기는 일은 5.2.2.2 임.
func (authority strategyProposalMarketAuthority) dispatchHandoffs() []strategyhandoff.Handoff {
	if !authority.familyActivation().Verified() {
		return []strategyhandoff.Handoff{authority.dispatchHandoff()}
	}
	// 조정자가 정한 순서 그대로 꺼냄(dispatchHandoff 와 같은 읽기 — Proposal 은 부작용 없는 접근자).
	selected := make([]strategyflow.Result, 0, len(authority.entries))
	for _, entry := range authority.entries {
		selected = append(selected, entry.authority.Proposal())
	}
	return strategyhandoff.AdmitEachOwnerScope(authority.snapshot.Ready, selected)
}

// deliverEachStrategyHandoff 는 handoff 들을 조정자 순서대로 하나씩 몸통에 건넴(태스크 5.2.2.1).
//
// **첫 오류에서 멈춤.** 한 소유자 범위의 몸통이 오류(원장 읽기 실패 · dispatch 거절 등)를 내면 같은 주기의 뒤 범위로
// 넘어가지 않고 그 오류를 그대로 돌려줌 — 예상 밖 오류 뒤에 같은 주기에서 주문을 더 내지 않는 보수 방향. 범위마다 고장을
// 격리해 뒤 범위를 계속 내보낼지는 하류 권한이 소유자 범위 단위가 되는 5.2.2.2 의 결정으로 남김. handoff 가 하나뿐인
// 오늘 경로에서는 `handoffs[0].Deliver(body)` 와 같은 값을 돌려줌(감싸지 않음 — 오류 분류가 그대로 보존됨).
//
// 거절된 handoff 는 Deliver 가 몸통을 부르지 않고 nil 을 돌려주므로 다음 범위로 넘어감(거절은 오류가 아님, Deliver 머리말).
func deliverEachStrategyHandoff(handoffs []strategyhandoff.Handoff, body func(strategyhandoff.Delivered) error) error {
	for _, handoff := range handoffs {
		if err := handoff.Deliver(body); err != nil {
			return err
		}
	}
	return nil
}

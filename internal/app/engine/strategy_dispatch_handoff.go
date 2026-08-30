package engine

import "github.com/JungHoonGhae/tossinvest-cli/internal/strategyflow"

// strategyMarketHandoffCapacity 는 한 시장이 한 주기에 공유 dispatch 로
// 넘길 수 있는 선택의 수다.
//
// **이 값은 고른 것이 아니라 지금 코드에서 읽은 것이다.** 이 태스크 전까지
// 엔진은 같은 규칙 `len(entries) != 1` 을 다섯 자리(worker 승격, dispatch 주기,
// 결과 권한, 읽기 전용 projection, first-leg 권한)에 따로 적어 두고 있었다.
// 다섯 개의 사본은 각각 조금씩 다른 동반 조건을 달고 있었고, 하나가 바뀌어도
// 나머지는 조용히 옛 규칙을 유지한다.
//
// 그리고 이 값은 계약이 아니라 **빚**이다. 동결 골든
// `analysis/goldens/four-family-runtime-v1.json` 은
// `queue.market_wide_single_proposal_assumption_forbidden = true` 이고
// `queue.selected_limit` 는 "시장마다 하나"가 아니라 "**소유자 범위마다** 하나"다.
// 즉 종목이 둘인 시장이 아무 말 없이 아무것도 안 하는 지금 동작은 없어져야
// 한다. 그것을 실제로 없애는 것은 태스크 5.2(시장 단위 단일 제안 준비 상태를
// 시장마다 네 개의 독립 레인 worker 로 교체)다. 이 태스크는 흩어진 사본을
// 한 자리로 모으고 그 거절에 이름을 붙이는 데까지만 한다 — 그래야 5.2 가
// 상수 하나를 고치는 일이 되고, 그때까지 무엇이 막혀 있는지 말할 수 있다.
const strategyMarketHandoffCapacity = 1

// strategyHandoffRefusal 은 조정자가 고른 것이 공유 dispatch 까지 가지 못한
// 이유다. 빈 값은 거절이 없다는 뜻이다.
//
// 이 이름들은 중재 거절(`strategyarbiter.Refusal`)을 빌려 쓰지 않는다. 중재는
// "누가 이 범위를 가져가는가"를 정하다 실패한 것이고, 여기는 중재가 끝난 뒤
// 그 결과를 넘기는 자리에서 막힌 것이다. 같은 이름을 쓰면 운영자가 둘을
// 구별할 수 없다.
type strategyHandoffRefusal string

const (
	strategyHandoffAdmitted     strategyHandoffRefusal = ""
	strategyHandoffMarketClosed strategyHandoffRefusal = "HANDOFF_MARKET_CLOSED"
	strategyHandoffNoSelection  strategyHandoffRefusal = "HANDOFF_NO_SELECTION"
	strategyHandoffOverCapacity strategyHandoffRefusal = "HANDOFF_OVER_CAPACITY"
)

// strategyDispatchHandoff 는 시장 조정자가 고른 것이 공유 dispatch 로 건너가는
// 유일한 값이다.
//
// 목록을 그대로 건네지 않는 이유: 목록을 받는 쪽마다 "몇 개까지 괜찮은가"를
// 스스로 정하게 되고, 그 판단이 서로 달라져도 아무도 모른다. 여기서 하나로
// 줄여 건네면 그 판단은 한 번만 내려진다.
type strategyDispatchHandoff struct {
	// result 는 넘어가는 봉인된 제안이다. refusal 이 있으면 비어 있다.
	result strategyflow.Result
	// refusal 은 못 넘긴 이유다.
	refusal strategyHandoffRefusal
	// pending 은 조정자가 이 시장에서 고른 소유자 범위의 수다. 거절일 때도
	// 채운다 — "몇 개라서 막혔는지"를 말하지 못하면 운영자는 상한을 볼 수 없다.
	pending int
}

// Admitted 는 이 handoff 가 공유 dispatch 로 건너갈 수 있는지 답한다.
func (handoff strategyDispatchHandoff) Admitted() bool {
	return handoff.refusal == strategyHandoffAdmitted
}

// dispatchHandoff 는 한 시장의 조정 결과에서 공유 dispatch 로 넘길 하나를 고른다.
//
// 순서를 여기서 정하지 않는다. `entries` 는 이미 조정자가 정한 소유자 범위
// 순서로 와 있고, 그 순서가 `ProposalSetDigest` 에 들어간다. 여기서 다시
// 정렬하면 같은 시장이 주기마다 다른 것을 고를 수 있다.
func (authority strategyProposalMarketAuthority) dispatchHandoff() strategyDispatchHandoff {
	pending := len(authority.entries)
	// 닫힌 시장에서 항목이 새어 나오지 않게 준비 상태를 먼저 본다. 지금은
	// 항목이 있으면 반드시 준비된 시장이지만(그 불변식은
	// TestEveryMarketAuthorityCarryingEntriesAlsoReportsReady 가 지킨다),
	// 그 불변식이 깨지는 날 조용히 통과하는 것보다 여기서 막는 것이 낫다.
	if !authority.snapshot.Ready {
		return strategyDispatchHandoff{refusal: strategyHandoffMarketClosed, pending: pending}
	}
	if pending == 0 {
		return strategyDispatchHandoff{refusal: strategyHandoffNoSelection, pending: pending}
	}
	if pending > strategyMarketHandoffCapacity {
		return strategyDispatchHandoff{refusal: strategyHandoffOverCapacity, pending: pending}
	}
	return strategyDispatchHandoff{result: authority.entries[0].authority.Proposal(), pending: pending}
}

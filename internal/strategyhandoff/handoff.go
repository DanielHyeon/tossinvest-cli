// Package strategyhandoff 은 한 시장의 조정 결과가 공유 dispatch 로 건너가는
// 유일한 경계다.
//
// 이 패키지가 엔진 **밖에** 있는 이유는 하나다. "여기서는 주문을 낼 수 없다"는
// 약속은 같은 패키지 안에서는 증명할 수 없다. 엔진 안에 있으면 import 를
// 하나도 늘리지 않고도 같은 패키지의 브로커 타입(officialBroker)을 손에 넣어
// 손절 주문을 취소할 수 있고, import 를 세는 어떤 검사도 그것을 보지 못한다.
// 밖으로 나오면 그 약속은 import 목록이 지킨다 — dependency_closure_test.go
// 가 그 목록을 읽고, 전이 의존까지 따라간다.
package strategyhandoff

import "github.com/JungHoonGhae/tossinvest-cli/internal/strategyflow"

// Capacity 는 이 경계가 한 시장·한 주기에 실제로 건네줄 수 있는 선택의 수다.
//
// **고른 값이 아니라 건네주는 코드에서 읽은 값이다.** Single 은 하나를
// 돌려주므로 이 값은 1 이다. 그래서 이 상수만 올려서는 두 번째 선택이 건너가지
// 않는다 — 올리면 Single 이 값을 내주지 않고, 그날 dispatch 는 하나를 내보내는
// 대신 아무것도 내보내지 않는다. 조용히 하나를 버리는 것보다 낫고, 그것이
// 다중 dispatch 를 구현하라는 신호다.
// TestARaisedCapacityHandsOffNothingRatherThanDroppingTheSecondSelection 이
// 그 성질을 지킨다.
//
// 이 값은 계약이 아니라 **빚**이다. 동결 골든
// analysis/goldens/four-family-runtime-v1.json 의 queue 블록은
// "market_wide_single_proposal_assumption_forbidden": true 이고
// "selected_limit": "at most one selected proposal per owner scope" 다.
// 즉 종목이 둘인 시장이 아무 말 없이 아무것도 안 하는 지금 동작은 없어져야
// 한다. 그것을 실제로 없애는 것은 태스크 5.2(시장 단위 단일 제안 준비 상태를
// 시장마다 네 개의 독립 레인 worker 로 교체)이며, 그 로트는 이 상수와 Single
// 의 서명을 **함께** 바꿔야 한다.
const Capacity = 1

// Refusal 은 조정자가 고른 것이 공유 dispatch 까지 가지 못한 이유다.
// 빈 값은 거절이 없다는 뜻이다.
//
// 이 이름들은 중재 거절(strategyarbiter.Refusal)을 빌려 쓰지 않는다. 중재는
// "누가 이 범위를 가져가는가"를 정하다 실패한 것이고, 여기는 중재가 끝난 뒤
// 그 결과를 넘기는 자리에서 막힌 것이다. 같은 이름을 쓰면 운영자가 둘을
// 구별할 수 없다.
type Refusal string

const (
	Admitted     Refusal = ""
	MarketClosed Refusal = "HANDOFF_MARKET_CLOSED"
	NoSelection  Refusal = "HANDOFF_NO_SELECTION"
	OverCapacity Refusal = "HANDOFF_OVER_CAPACITY"
)

// Handoff 는 시장 조정자가 고른 것이 공유 dispatch 로 건너가는 유일한 값이다.
//
// 필드는 모두 비공개이고, 실린 제안은 Single 로만 나간다. 그래서 "거절 여부를
// 안 보고 값을 읽는" 코드는 쓸 수 없다 — 값을 받으려면 두 번째 반환값을 함께
// 받게 되고, 그것을 무시하려면 명시적으로 버려야 한다. 이 성질을 AST 검사가
// 아니라 타입이 지킨다는 점이 요점이다. 검사는 우회할 수 있지만 서명은 못
// 우회한다.
type Handoff struct {
	// selected 는 승인되었을 때 실린 선택들이다. 거절이면 비어 있다.
	selected []strategyflow.Result
	// refusal 은 못 넘긴 이유다.
	refusal Refusal
	// pending 은 조정자가 이 시장에서 고른 소유자 범위의 수다. 거절일 때도
	// 채운다 — "몇 개라서 막혔는지"를 말하지 못하면 운영자는 상한을 볼 수 없다.
	pending int
}

// Admit 은 한 시장의 조정 결과를 경계 값으로 바꾼다.
//
// 순서를 여기서 정하지 않는다. selected 는 이미 조정자가 정한 소유자 범위
// 순서로 오고, 그 순서가 ProposalSetDigest 에 들어간다. 여기서 다시 정렬하면
// 같은 시장이 주기마다 다른 것을 고를 수 있다.
func Admit(ready bool, selected []strategyflow.Result) Handoff {
	pending := len(selected)
	// 닫힌 시장에서 항목이 새어 나오지 않게 준비 상태를 먼저 본다.
	if !ready {
		return Handoff{refusal: MarketClosed, pending: pending}
	}
	if pending == 0 {
		return Handoff{refusal: NoSelection, pending: pending}
	}
	if pending > Capacity {
		return Handoff{refusal: OverCapacity, pending: pending}
	}
	// 부르는 쪽의 배열을 그대로 들고 있지 않는다. 나중에 그 배열이 바뀌면
	// 이미 판정이 끝난 handoff 가 다른 값을 건네게 된다.
	return Handoff{selected: append([]strategyflow.Result(nil), selected...), pending: pending}
}

// Refusal 은 왜 못 넘겼는지 답한다. 승인이면 빈 값이다.
func (handoff Handoff) Refusal() Refusal { return handoff.refusal }

// Pending 은 조정자가 이 시장에서 고른 소유자 범위의 수다.
func (handoff Handoff) Pending() int { return handoff.pending }

// Single 은 공유 dispatch 로 건너갈 하나를 돌려준다. 두 번째 값이 false 면
// 값은 없다.
//
// 승인되었더라도 실린 것이 하나가 아니면 내주지 않는다. 그 자리가 이 경계의
// fail-closed 지점이다: 건네줄 수 있는 것보다 많이 실렸다는 말은 이 서명이
// 상한과 어긋났다는 뜻이고, 그때 하나만 골라 내주면 나머지는 거절 이름도
// 계수기도 없이 사라진다.
func (handoff Handoff) Single() (strategyflow.Result, bool) {
	if handoff.refusal != Admitted || len(handoff.selected) != 1 {
		return strategyflow.Result{}, false
	}
	return handoff.selected[0], true
}

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

import (
	"errors"

	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyflow"
)

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

// deliverable 은 이 경계가 **실제로 건네주는** 선택의 수다. Single 과 Deliver 가
// 하나를 건네므로 1 이다.
//
// Capacity 와 따로 두는 이유: 두 값이 갈라질 수 있고, 갈라지는 순간이 위험한
// 순간이기 때문이다. Capacity 만 올리면 Admit 은 승인하는데 건네줄 수는 없다.
// 그 상태는 무명으로 두지 않고 OverCarried 라고 부른다.
const deliverable = 1

// ErrNoDelivery 는 Deliver 에 몸통 없이 부른 프로그래밍 오류다. 조용히
// 성공시키면 "건넸다"와 "건넬 곳이 없었다"가 같은 답이 된다.
var ErrNoDelivery = errors.New("strategyhandoff: delivery body is nil")

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
	// OverCarried 는 승인됐지만 건네줄 수 있는 것보다 많이 실린 상태다.
	// Capacity 가 deliverable 보다 커지는 날에만 생긴다. 적대 리뷰가 지적한
	// 대로, 이름 없는 fail-closed 는 이 change 가 없애려던 것과 같은 모양이다.
	OverCarried Refusal = "HANDOFF_OVER_CARRIED"
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

// refusalNow 는 이 경계의 **유일한 술어**다. Refusal·Single·Deliver 가 모두
// 이것을 부른다.
//
// 첫 수정 판본은 Refusal 과 Single 이 각각 다른 조건을 썼다. 상한을 올리면
// Refusal 은 Admitted 라고 하면서 Single 은 값을 안 줬다 — 주문 경로는
// fail-closed 였지만 이름과 계수기는 "승인"이라고 보고했다.
func (handoff Handoff) refusalNow() Refusal {
	if handoff.refusal != Admitted {
		return handoff.refusal
	}
	// 적재 미달과 과적재를 가른다. 앞 판본은 `!= deliverable` 하나로 갈라서
	// 실린 것이 **없는** 영값 Handoff 까지 "많이 실렸다"라고 불렀다. 둘 다
	// 거절인 것은 같지만, 이름이 반대면 운영자가 보는 것이 거짓이 된다.
	//
	// 적재 미달에 새 이름을 만들지 않는다. 실린 것이 없는 상태는 고른 것이
	// 없는 상태와 같으므로 NoSelection 이 그대로 맞고, 동결 골든에 없는 거절
	// 어휘를 영수증 없이 하나 더 늘리지 않는다.
	if len(handoff.selected) < deliverable {
		return NoSelection
	}
	if len(handoff.selected) > deliverable {
		return OverCarried
	}
	return Admitted
}

// Refusal 은 왜 못 넘겼는지 답한다. 승인이면 빈 값이다.
func (handoff Handoff) Refusal() Refusal { return handoff.refusalNow() }

// Pending 은 조정자가 이 시장에서 고른 소유자 범위의 수다.
func (handoff Handoff) Pending() int { return handoff.pending }

// Single 은 공유 dispatch 로 건너갈 하나를 돌려준다. 두 번째 값이 false 면
// 값은 없다.
//
// 주문을 내는 자리에서는 이것 대신 Deliver 를 쓴다. 여기서 돌려주는 bool 은
// 부르는 쪽이 무시할 수 있고, 적대 리뷰가 실제로 무시하는 편집을 통과시켰다.
// 이 서명이 남아 있는 세 자리(worker 승격, 결과 권한, 읽기 전용 projection)는
// 값을 쓰기 전에 반드시 ValidProposal 을 다시 보므로, 답을 무시해도 영값이
// 걸린다 — 그 성질은 우연이 아니라 TestARefusedSingleReturnsTheZeroResult
// 와 strategyflow 의 시험이 값으로 지킨다.
func (handoff Handoff) Single() (strategyflow.Result, bool) {
	if handoff.refusalNow() != Admitted {
		return strategyflow.Result{}, false
	}
	return handoff.selected[0], true
}

// Delivered 는 이 경계를 지나온 값이라는 증거다. 밖에서는 채울 수 없다.
//
// **왜 값 대신 봉투인가.** 적대 리뷰 세 라운드가 각각 `rawSelection()`,
// `relay()`, `rawTailProposal()` 로 같은 구멍을 뚫었다. 셋 다 경계를 지나지
// 않은 strategyflow.Result 를 만들어 공유 dispatch 에 넘겼고, 셋 다 두 스위트를
// 통과했다. 막던 것이 "그 값을 묶어 준 **이름**을 쓰는가"를 보는 AST 검사였고,
// 이름은 언제나 다시 쓸 수 있기 때문이다.
//
// 봉투에서는 그 편집이 컴파일되지 않는다. result 는 비공개 필드라서 이 패키지
// 밖에서는 `Delivered{}` 라는 영값밖에 만들 수 없고, 그 영값은 dispatch 의
// 첫 줄 validateStrategyFirstLegResult 에서 걸린다 — 가정이 아니라
// TestAForgedEnvelopeIsRefusedBeforeAnyGatewayCall 이 값으로 확인한다.
//
// **이 타입이 증명하지 못하는 것도 적는다.** 봉투는 "dispatch 된 값이 Admit 을
// 거쳤다"만 증명한다. 그 Admit 을 부른 것이 시장 조정자였는지는 증명하지
// 않는다 — 엔진 패키지는 스스로 Admit 을 부를 수 있다. 그 자리는 타입이 아니라
// TestExactlyOneProductionSiteAdmitsIntoTheSeam 이 지키고, 그 검사는 함수
// 본문이 아니라 패키지 전체의 호출을 센다.
type Delivered struct {
	// result 는 경계가 승인한 값이다. 비공개인 것이 이 타입의 전부다.
	result strategyflow.Result
}

// Result 는 봉투가 실은 값을 돌려준다. 밖에서 만든 봉투는 영값을 돌려준다.
func (delivered Delivered) Result() strategyflow.Result { return delivered.result }

// Deliver 는 건너간 것이 있을 때만 몸통을 부른다.
//
// **이 서명의 요점은 부르는 쪽에 무시할 boolean 을 주지 않는 것이다.**
// Single 판본에서는 답을 받아 `if !handedOff { }` 같은 빈 조건에 넣으면
// 관문이 사라지면서도 "답을 썼다"는 검사가 통과했다. 여기서는 몸통이 도는지를
// 부르는 쪽이 정하지 않는다.
//
// 거절은 오류가 아니다 — 그 주기에 낼 것이 없었을 뿐이다. 몸통의 오류만
// 그대로 밖으로 나간다.
func (handoff Handoff) Deliver(to func(Delivered) error) error {
	if to == nil {
		return ErrNoDelivery
	}
	result, ok := handoff.Single()
	if !ok {
		return nil
	}
	return to(Delivered{result: result})
}

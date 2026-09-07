package positioncampaign

// fill.go 는 "원장에 쓰는 쪽과 원장을 다시 읽는 쪽이 **같은 답**을 내야 하는 규칙"을
// 한 곳에 모은 파일이다.
//
// 왜 따로 두는가
// -------------
// 규칙 하나가 두 자리에 서로 다른 방법으로 구현돼 있으면, 시험은 자기가 보고 쓴 쪽을
// 통과시킨다. 그래서 두 구현이 갈려 있어도 스위트는 초록이고, 변이를 심어도 반대쪽이
// 대신 막아 주어 살아남는다. a065 적대 리뷰가 찾은 결함 대부분이 이 기전이었다:
//
//   - leg 상태:      journal 의 함수 안 if  vs  TransitionLeg 표
//   - entry_blocked: 체결이 상태에서 재계산  vs  UpdateCampaignStop 의 단조 래치
//   - successor 잔량: cap 기준(journal)      vs  leg 잔여 기준(LegLedger)
//
// 그래서 세 규칙을 각각 함수 **하나**로 만들고, 쓰는 쪽과 읽는 쪽이 모두 그 함수를
// 부른다. 갈릴 자리가 없으면 갈리지 않는다.

import "github.com/JungHoonGhae/tossinvest-cli/internal/riskcalc"

// LegFillFacts 는 누적 체결 관측 하나가 나르는 사실 전부다.
//
// 수량은 전부 10진 문자열이다. 부동소수점은 원장 수량에 쓰지 않는다.
type LegFillFacts struct {
	// LegState 는 이 관측 **이전** 의 leg 상태다.
	LegState LegState
	// Delta 는 이 관측이 새로 더하는 수량이다 (>= 0).
	Delta string
	// LegFilled 는 이 관측을 반영한 뒤의 leg 누적 체결 합계다.
	LegFilled string
	// LegRequested 는 leg 가 계획한 수량이다.
	LegRequested string
	// OrderTerminal 은 이 관측으로 그 broker order 가 종결되는지다.
	OrderTerminal bool
	// OrderHasSuccessor 는 그 broker order 를 잇는 교체 주문이 이미 있는지다.
	// 후속이 있으면 잔량은 취소된 것이 아니라 넘어간 것이므로 leg 는 닫히지 않는다.
	OrderHasSuccessor bool
}

// LegEventForFill 은 사실을 D4 leg 표가 아는 사건 하나로 분류한다.
//
// 순서가 규칙이다. 아래에서 위로 읽으면 안 된다.
func LegEventForFill(facts LegFillFacts) (LegEvent, error) {
	deltaCmp, err := riskcalc.CompareDecimal(facts.Delta, "0")
	if err != nil {
		return "", err
	}
	// 1. leg 가 이미 종결이면 어떤 관측도 그것을 되돌리지 않는다. 뒤늦은 양수 delta 는
	//    수량으로만 남고(수량 권위는 Position 이다) 상태는 terminal 로 유지된다.
	if facts.LegState == LegFilled || facts.LegState == LegCancelled {
		if deltaCmp > 0 {
			return LegLatePositiveFill, nil
		}
		return LegTerminalRetry, nil
	}
	filledCmp, err := riskcalc.CompareDecimal(facts.LegFilled, facts.LegRequested)
	if err != nil {
		return "", err
	}
	// 2. 계획 수량을 채웠다.
	if filledCmp >= 0 {
		return LegFullFill, nil
	}
	// 3. 후속 없이 주문이 종결됐다 — 잔량은 취소다. 여기서 delta 가 양수여도 종결이
	//    우선한다. "3주 체결 + 잔량 취소" 가 관측 하나로 오는 경우가 정확히 여기다.
	if facts.OrderTerminal && !facts.OrderHasSuccessor {
		zeroCmp, err := riskcalc.CompareDecimal(facts.LegFilled, "0")
		if err != nil {
			return "", err
		}
		if zeroCmp == 0 {
			return LegZeroFillCancelled, nil
		}
		return LegResidualCancelled, nil
	}
	// 4. 새 수량이 들어왔다.
	if deltaCmp > 0 {
		return LegPartialFill, nil
	}
	// 5. 아무것도 늘지 않았다.
	return LegDuplicateObservation, nil
}

// LegStateAfterFill 은 분류와 표를 한 번에 적용한다. 원장에 쓰는 쪽과 재구성하는 쪽이
// 모두 이 함수를 부른다.
func LegStateAfterFill(facts LegFillFacts) (LegState, error) {
	event, err := LegEventForFill(facts)
	if err != nil {
		return "", err
	}
	return TransitionLeg(facts.LegState, event)
}

// LatchEntryBlocked 는 진입 차단이 **단조**임을 정하는 유일한 자리다.
//
// entry_blocked 는 campaign 상태만의 함수가 아니다. stop 증거가 없거나 무효일 때
// UpdateCampaignStop 이 상태를 ACTIVE 로 둔 채 이 열만 1 로 건다(spec: "stop evidence 가
// missing 또는 invalid … 새 exposure-raising leg 는 fail closed 된다"). 상태에서
// 되계산하면 그 래치가 평범한 체결 하나에 지워진다 — 그래서 이미 걸린 래치는
// 전이 결과와 **OR** 로 합친다. 푸는 것은 전이표가 아니라 reconcile 해소의 일이다.
func LatchEntryBlocked(stored bool, next CampaignTransition) CampaignTransition {
	next.EntryBlocked = next.EntryBlocked || stored
	return next
}

// SuccessorRemaining 은 교체 주문이 아직 더 사도 되는 수량이다.
//
// 두 상한 중 **작은 쪽**이다.
//   - 자기 cap 잔여: 이 주문 자신이 broker 에 요청한 상한을 넘을 수 없다.
//   - leg 잔여:      leg 가 더는 필요로 하지 않는 수량을 사면 안 된다. predecessor 의
//     늦은 체결이 leg 잔여를 줄이면 후속도 같이 줄어야 한다(design D5).
//
// cap 만 보면 predecessor 의 늦은 체결이 두 피연산자 어디에도 안 들어가서 재계산이
// 증명 가능한 no-op 이 된다. leg 잔여만 보면 주문 자신의 상한을 넘길 수 있다.
func SuccessorRemaining(orderCap, orderCumulative, legResidual string) (string, error) {
	capRemaining, err := remainingFromCap(orderCap, orderCumulative)
	if err != nil {
		return "", err
	}
	residual, err := nonNegative(legResidual)
	if err != nil {
		return "", err
	}
	return riskcalc.MinDecimal(capRemaining, residual)
}

// StoredOrderRemaining 은 campaign_order_watermarks.remaining_quantity 가 가져야 할
// 값이다. 원장에 쓰는 쪽과 재구성하는 쪽이 모두 이 함수를 부른다.
//
// 종결된 주문은 더 살 수 없으므로 leg 잔여가 줄어도 그 행은 움직이지 않는다 —
// 자기 cap 잔여가 그대로 마지막 사실이다.
func StoredOrderRemaining(hasPredecessor, terminal bool, orderCap, orderCumulative, legResidual string) (string, error) {
	if hasPredecessor && !terminal {
		return SuccessorRemaining(orderCap, orderCumulative, legResidual)
	}
	return remainingFromCap(orderCap, orderCumulative)
}

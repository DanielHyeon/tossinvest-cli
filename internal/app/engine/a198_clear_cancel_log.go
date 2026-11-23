package engine

// a198_clear_cancel_log.go 는 보호 청산의 길 치우기가 엔진 주문을 취소하지 못한 사실을 운영 로그(ExitOrderLog)에 남김.
//
// 운영 조립에서 관측자 전체 Log 는 nil 이라(exit_order_log.go 머리 주석) warnEpisode 는 아무것도 남기지 않음. 그래서 a129 가
// 운영에 배선한 전용 로거에 **직접** 한 줄을 씀(a198 design D4). a129 의 대기열 · 래치 · 줄 타입은 쓰지 않음 — 그 줄 타입은
// 매도 지정가를 고정 출력하고 래치 서명에 주문 번호가 없어 같은 포지션의 두 취소를 한 줄로 합침.
//
// 줄에는 종류 · 포지션 · 종목 · 시장 · 주문 방향 · 대상 주문 번호 · 결과 상태 · 사유 코드만 실음. Detail · 원문 오류 · 계좌 ·
// 수량 · 가격은 싣지 않음(불변식 8). 판정은 바꾸지 않음 — 이 함수는 기록만 함.

import (
	"strings"

	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
)

// exitClearCancelDecisionFailed 는 취소 결정을 원장에 기록하지 못했다는 고정 사유 코드임 — 오류 원문 대신 실음.
const exitClearCancelDecisionFailed = "cancel_decision_failed"

// noteClearCancelFailure 는 취소 하나의 실패를 한 줄 씀. 같은 주문의 같은 결과(상태 · 사유)가 주기마다 반복되면 처음 한 번만 —
// 관측 주기마다 같은 줄을 쓰는 것은 관측이 아님(a074). 결과가 바뀌면 다시 씀.
func (o *ExitObserver) noteClearCancelFailure(m managed, order journal.LiveOrder, state journal.AttemptState, reason string) {
	logger := o.opts.ExitOrderLog
	if logger == nil {
		return
	}
	reason = exitOrderToken(reason)
	key := m.position.ID + "|" + strings.TrimSpace(order.OrderID)
	mark := string(state) + "|" + reason
	if o.clearCancelLogged[key] == mark {
		return
	}
	if o.clearCancelLogged == nil {
		o.clearCancelLogged = map[string]string{}
	}
	o.clearCancelLogged[key] = mark

	// 이벤트 이름은 submit 의 청산 주문 줄과 같은 기준(exitloop.go submit): 브로커 거부 · 결과 미상 · 그 밖(브로커 전 거절).
	event := obs.EventOrderNotDispatched
	switch state {
	case journal.StateFailedConfirmed:
		event = obs.EventOrderRejected
	case journal.StateInDoubt:
		event = obs.EventOrderInDoubt
	}
	// 쓰기 대상이 패닉해도 손절 루프는 계속 돎(writeExitOrder 와 같은 규율).
	defer func() { _ = recover() }()
	logger.Warn(event,
		"origin", "exit",
		"kind", "cancel",
		"position_id", m.position.ID,
		obs.FieldSymbol, order.Symbol,
		obs.FieldMarket, order.Market,
		obs.FieldSide, strings.ToUpper(strings.TrimSpace(order.Side)),
		obs.FieldOrderID, order.OrderID,
		obs.FieldToState, exitOrderToken(string(state)),
		obs.FieldReason, reason,
	)
}

// noteClearCancelConfirmed 는 취소가 접수 확정된 주문의 표시를 지움 — 그 주문이 나중에 다시 실패하면 다시 씀.
func (o *ExitObserver) noteClearCancelConfirmed(m managed, order journal.LiveOrder) {
	delete(o.clearCancelLogged, m.position.ID+"|"+strings.TrimSpace(order.OrderID))
}

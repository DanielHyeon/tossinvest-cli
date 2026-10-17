package engine

// trade_outcome_refusal_log.go 는 a130 design D3 의 엔진 쪽임 — 성과 동결 거절을 engine.log 에 한 줄로 남기는 싱크.
// NewContext 가 bindApplyHooks(jrn, tradeOutcomeRefusalSink(opts.Logger)) 로 묶음.
//
// # 왜 별도 goroutine 인가
//
// 싱크는 체결 트랜잭션이 커밋된 **뒤** RecordFill 안에서 불림(쓰기 잠금은 이미 풀림). 그래도 막힌 로그 쓰기(디스크 정체)가 체결
// 감지 루프를 붙잡으면 다음 체결 반영이 늦어지므로, 쓰기를 새 goroutine 으로 넘기고 곧바로 돌아옴. 거절은 포지션 종결마다 많아야
// 한 번이라 goroutine 수는 종결 수로 묶임. goroutine 안의 패닉도 거둠.
//
// # 줄의 필드
//
// journal.TradeOutcomeRefusal 의 넷(종목 · 시장 · 회차 · 사유 코드)만 실음 — 계좌 참조 · 포지션 id 없음. 사유는 journal 의 고정
// 토큰 상수 중 하나임.

import (
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
)

// eventTradeOutcomeNotFrozen 은 a130 이 더하는 이벤트 이름임(obs 등급표에 없으므로 normal). internal/obs 는 이 change 의 편집
// 범위 밖이라 여기 둠(선례: exit_order_log.go 의 eventExitOrderEnded).
const eventTradeOutcomeNotFrozen obs.EventType = "trade_outcome.not_frozen"

// tradeOutcomeRefusalSink 는 로거가 있으면 거절을 줄 하나로 남기는 싱크를, 없으면 nil(a130 이전과 같은 바인딩)을 돌려줌.
func tradeOutcomeRefusalSink(log *obs.Logger) func(journal.TradeOutcomeRefusal) {
	if log == nil {
		return nil
	}
	return func(r journal.TradeOutcomeRefusal) {
		go func() {
			defer func() { _ = recover() }()
			log.Warn(eventTradeOutcomeNotFrozen,
				"symbol", r.Symbol,
				"market", r.Market,
				"instance_seq", r.InstanceSeq,
				"reason", string(r.Reason))
		}()
	}
}

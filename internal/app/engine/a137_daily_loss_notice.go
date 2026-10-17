package engine

// a137_daily_loss_notice.go 는 측정 불가로 조용히 끝난 전략 주기를 감독자 스냅숏 계수 · 계좌 일 키별 알림 한 번으로 남기는 자리임
// (재리뷰 P1-N1, Manager 결정 ① 의 보완). 잠금 · 삼킴 계수 · 운영 모드는 건드리지 않음 — 진입은 매 주기 다시 거절되는 채로 둠.

import (
	"context"
	"math"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
)

type strategyDailyLossRecorderKey struct{}

// withStrategyDailyLossRecorder 는 한 시장 주기의 ctx 에 기록자를 실음(감독자 runMarket 만 부름).
func withStrategyDailyLossRecorder(ctx context.Context, record func(error)) context.Context {
	return context.WithValue(ctx, strategyDailyLossRecorderKey{}, record)
}

// recordStrategyDailyLossUnmeasurable 은 ctx 의 기록자에 측정 불가를 넘김. 기록자가 없으면(시험 · 감독자 밖) 아무것도 안 함.
func recordStrategyDailyLossUnmeasurable(ctx context.Context, cause error) {
	if ctx == nil || cause == nil {
		return
	}
	if record, ok := ctx.Value(strategyDailyLossRecorderKey{}).(func(error)); ok && record != nil {
		record(cause)
	}
}

// recordDailyLossUnmeasurable 은 계수를 올리고 첫 원인을 남기며, 계좌 일 키가 바뀐 첫 번째에만 알림을 냄.
func (s *StrategyEntrySupervisor) recordDailyLossUnmeasurable(worker *strategyMarketRuntime, cause error) {
	if s == nil || worker == nil || cause == nil {
		return
	}
	day, err := clock.AccountTradingDay(s.clk.Now())
	if err != nil {
		day = ""
	}
	s.mu.Lock()
	if worker.dailyLossUnmeasurable < math.MaxUint64 {
		worker.dailyLossUnmeasurable++
	}
	if worker.firstDailyLossCause == "" {
		worker.firstDailyLossCause = cause.Error()
	}
	announce := day != "" && worker.dailyLossNoticeDay != day
	if announce {
		worker.dailyLossNoticeDay = day
	}
	notice, market := s.dailyLossNotice, worker.descriptor.Market
	s.mu.Unlock()
	if announce && notice != nil {
		notice(market, day, cause.Error())
	}
}

// strategyDailyLossNotice 는 생산 조립의 알림 자리 — 경고 줄 하나(계좌 일 키마다). 로그가 없으면 nil(스냅숏 계수만).
func (c *Context) strategyDailyLossNotice() func(StrategyMarket, string, string) {
	if c == nil || c.Log == nil {
		return nil
	}
	log := c.Log
	return func(market StrategyMarket, day, cause string) {
		log.Warn(obs.EventDailyLossUnmeasurable, obs.FieldMarket, string(market), "account_day", day, obs.FieldDetail,
			"new strategy entries are refused every cycle until today's realised loss can be measured; "+
				"sells, stops and exits are unaffected: "+cause)
	}
}

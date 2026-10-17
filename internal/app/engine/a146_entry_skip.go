package engine

// a146_entry_skip.go 는 Guardian 상한 0(한 단위도 못 사는 전략 진입)을 그 진입 하나의 건너뜀으로 다루는 자리임(사용자 결정 2026-10-15).
//
// 결함이 아니라 정책 결과이므로 전달 루프는 오류 없이 다음 범위로 가고(잠금 · 삼킴 없음), 흔적은 감독자 스냅숏의 계수 · 첫 원인과
// (계좌 일 키, 종목)마다 경고 줄 하나로 남김. 원장 · 운영 모드 · 진입 게이트는 건드리지 않음 — 같은 후보는 다음 주기에 다시 판정됨.

import (
	"context"
	"errors"
	"fmt"
	"math"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyrouter"
)

// strategyEntrySkip 은 그 진입 하나의 건너뜀 표지임. 만드는 자리는 첫 다리 승인 하나뿐(언급 census 시험이 고정) — 늘면 진짜 결함이 조용해짐.
type strategyEntrySkip struct {
	market strategyrouter.Market
	symbol string
	// entry · stop 은 봉인된 실행 조건의 시장 가격(major 표기 — 계좌 값 아님). 가격 단위 결함(예: ×100)이 조용한 스킵이 된 경우를 운영자가
	// 가를 수 있게 알림 줄에 실음(freeze 리뷰 P3-2).
	entry, stop string
	// fx 는 그 진입의 봉인 FX 권위 식별(통화 쌍 · digest — 시장 자료, 계좌 값 아님). 봉인 권위는 율을 밖에 내보이지 않으므로 율 대신 digest 를
	// 싣는다(구현 리뷰 P3-4 — 단위 결함으로 생긴 상한 0 을 그 권위의 원문과 맞춰 가를 수 있게).
	fx    string
	cause error
}

// Error 는 고정 문구 + 시장 · 종목 · 코드 · 진입가 · 손절가만 — 계좌 값 · 금액 없음(스냅숏 · 알림 줄에 그대로 실림). 원래 거절은 Unwrap 으로 닿음.
func (s *strategyEntrySkip) Error() string {
	return fmt.Sprintf("engine: %s %s strategy entry skipped (%s) at entry %s stop %s fx %s: the Guardian cap permits fewer than one unit; "+
		"it is judged again next cycle", s.market, s.symbol, StrategyFirstLegGuardianCapZero, s.entry, s.stop, s.fx)
}

func (s *strategyEntrySkip) Unwrap() error { return s.cause }

// strategyEntrySkipOf 는 전달 루프가 「그 진입만 건너뜀」으로 받는 **두 모양만** 인정함(구현 리뷰 P2-1 · Manager 승격 P3-1):
//  1. 오류 자체가 *strategyEntrySkip,
//  2. dispatch 의 승인 감싸개 하나(`engine: first-leg admission GUARDIAN_CAP_ZERO: %w`)가 스킵 하나를 곧바로 감싼 것 — 문구가 정확히 그 감싸개이고
//     단일 Unwrap 이 스킵이어야 함.
//
// errors.As 로 트리 전체를 뒤지지 않는 이유: Join(스킵, 결함) 이나 「lease failed after %w(스킵)」처럼 결함이 스킵을 품은 모양에서 결함이 사라지면
// 안 됨 — 스킵은 결함을 가릴 수 없어야 함. 다중 오류(`Unwrap() []error`)는 어느 자리에 있든 스킵이 아님. 감싸개 문구는 dispatch 의
// `fmt.Errorf("engine: first-leg admission %s: %w", …)` 와 같아야 하고, 실제 경로 시험이 그 짝을 잼(문구가 갈라지면 스킵이 결함으로 — 막는 쪽).
func strategyEntrySkipOf(err error) (*strategyEntrySkip, bool) {
	if skip, ok := err.(*strategyEntrySkip); ok {
		return skip, true
	}
	// 다중 오류(`Unwrap() []error`)는 errors.Unwrap 이 nil 을 돌려주므로 아래에서 스킵이 아님 — 따로 거르지 않음(변이 round3 P1: 같은 판정의 중복).
	skip, ok := errors.Unwrap(err).(*strategyEntrySkip)
	if !ok || err.Error() != fmt.Sprintf("engine: first-leg admission %s: %s", StrategyFirstLegGuardianCapZero, skip.Error()) {
		return nil, false
	}
	return skip, true
}

type strategyEntrySkipRecorderKey struct{}

// strategyEntrySkipAccountDay 는 알림 dedup 의 계좌 일 키 함수(시험이 시간대 자료 결함을 흉내 내려고 바꿀 수 있는 유일한 자리).
// 패키지 변수라 바꾸는 시험은 t.Parallel 을 쓰면 안 되고 t.Cleanup 으로 되돌려야 함(구현 리뷰 P3-6).
var strategyEntrySkipAccountDay = clock.AccountTradingDay

// withStrategyEntrySkipRecorder 는 한 시장 주기의 ctx 에 스킵 기록자를 실음(감독자 runMarket 만 부름).
func withStrategyEntrySkipRecorder(ctx context.Context, record func(*strategyEntrySkip)) context.Context {
	return context.WithValue(ctx, strategyEntrySkipRecorderKey{}, record)
}

// recordStrategyEntrySkip 은 ctx 의 기록자에 스킵을 넘김. 기록자가 없으면(시험 · 감독자 밖) 아무것도 안 함 — 분류는 그대로.
func recordStrategyEntrySkip(ctx context.Context, skip *strategyEntrySkip) {
	if ctx == nil || skip == nil {
		return
	}
	if record, ok := ctx.Value(strategyEntrySkipRecorderKey{}).(func(*strategyEntrySkip)); ok && record != nil {
		record(skip)
	}
}

// recordGuardianCapZeroSkip 은 계수를 올리고(포화) 첫 원인을 지키며, 계좌 일 키가 같은 동안 종목마다 첫 번째에만 알림을 냄.
// 일 키가 바뀌면 알린 종목 기억을 비움 — 기억은 하루 종목 수만큼으로 유계. 일 키를 셀 수 없으면(시간대 자료 결함) 빈 일 키로도 알림 —
// a137 처럼 침묵하면 그때는 스킵이 운영자에게 전혀 보이지 않음(freeze 리뷰 P2-3).
func (s *StrategyEntrySupervisor) recordGuardianCapZeroSkip(worker *strategyMarketRuntime, skip *strategyEntrySkip) {
	if s == nil || worker == nil || skip == nil {
		return
	}
	day, err := strategyEntrySkipAccountDay(s.clk.Now())
	if err != nil {
		day = ""
	}
	s.mu.Lock()
	if worker.guardianCapZeroSkips < math.MaxUint64 {
		worker.guardianCapZeroSkips++
	}
	if worker.firstGuardianCapZeroSkip == "" {
		worker.firstGuardianCapZeroSkip = skip.Error()
	}
	if worker.entrySkipNoticed == nil || worker.entrySkipNoticeDay != day {
		worker.entrySkipNoticeDay, worker.entrySkipNoticed = day, map[string]struct{}{}
	}
	_, noticed := worker.entrySkipNoticed[skip.symbol]
	announce := !noticed
	if announce {
		worker.entrySkipNoticed[skip.symbol] = struct{}{}
	}
	notice, market := s.entrySkipNotice, worker.descriptor.Market
	s.mu.Unlock()
	if announce && notice != nil {
		notice(market, day, skip.symbol, skip.Error())
	}
}

// strategyEntrySkipNotice 는 생산 조립의 알림 자리 — 경고 줄 하나((계좌 일 키, 종목)마다). 로그가 없으면 nil(스냅숏 계수만).
func (c *Context) strategyEntrySkipNotice() func(StrategyMarket, string, string, string) {
	if c == nil || c.Log == nil {
		return nil
	}
	log := c.Log
	return func(market StrategyMarket, day, symbol, cause string) {
		log.Warn(obs.EventStrategyEntrySkipped, obs.FieldMarket, string(market), obs.FieldSymbol, symbol, "account_day", day, obs.FieldDetail,
			"a strategy entry is skipped, not treated as a lane fault, because the Guardian cap permits fewer than one unit; the candidate is "+
				"judged again every cycle and this line appears once per account day and symbol; sells, stops and exits are unaffected: "+cause)
	}
}

package execgw

// a137_daily_loss.go 는 Guardian 이 진입 판정 직전에 원장에서 그날의 실현 손실을 직접 재는 자리와, 그 한도 도달을 전략 경로가 레인 결함이
// 아닌 「계좌 진입 차단」으로 읽게 하는 표지임(a137 design D2 · D3).
//
// # 왜 Guardian 이 재는가
//
// 체인의 daily_loss 단(risk.checkDailyLoss)은 값이 없으면 막고 0 이면 통과함. 생산 입력은 서명 계좌 매니페스트의 값(저장소 안
// 생산자 0)과 tracer 의 리터럴 0 뿐이었음 — 한도는 보호가 아니었음. 그래서 Guardian 이 원장을 직접 재고, 호출자 값과 측정값이 **둘 다
// 유효할 때** 큰 쪽을 체인에 넣음. 호출자 값이 무효면 대체하지 않고(체인이 지금처럼 거절), 측정이 실패하면 호출자 값으로 떨어지지 않음.
//
// # 측정 토큰
//
// PrecheckQFinal* 은 ctx 가 없고 원장 · 브로커를 건드리지 않는 계약이라, 원장 읽기는 그 앞에서 MeasureDailyLoss 가 하고 결과를
// 다른 패키지가 만들 수 없는 토큰으로 넘김. precheck 는 토큰이 이 Guardian 의 것인지 · 10초 안이고 미래가 아닌지(riskcalc 의 계좌
// 스냅샷 신선 한계) · 일 키가 지금과 같은지 · 봉인된 측정 실패가 없는지를 봄.
//
// # 매도 · 손절
//
// 위험 축소 경로(IssueReduction · evaluateReduction)는 이 파일의 어떤 것도 부르지 않음.

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/dailyloss"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/risk"
	"github.com/JungHoonGhae/tossinvest-cli/internal/riskcalc"
)

// dailyLossReadMargin 은 원장 읽기 사전 필터 창의 양쪽 여유임 — 귀속은 dailyloss.Collect 가 closed_at 을 해석해 정함.
const dailyLossReadMargin = time.Hour

// DailyLossMeasurement 는 Guardian 이 잰 그날의 손실(환산 전)임. 칸이 모두 내보내지 않으므로 다른 패키지는 만들 수 없음.
type DailyLossMeasurement struct {
	guardian *RiskGuardian
	dayKey   string
	at       time.Time
	losses   dailyloss.Losses
	err      error
}

// MeasureDailyLoss 는 원장에서 계좌 일 키의 손실을 잼. 실패도 토큰 안에 봉인되어 precheck 의 daily_loss 단까지 감(앞 단 사유 보존).
func (g *RiskGuardian) MeasureDailyLoss(ctx context.Context) DailyLossMeasurement {
	now := g.clk.Now().UTC()
	m := DailyLossMeasurement{guardian: g, at: now}
	day, err := clock.AccountTradingDay(now)
	if err != nil {
		m.err = fmt.Errorf("%w: the account day cannot be computed", dailyloss.ErrUnmeasurable)
		return m
	}
	m.dayKey = day
	from, to, err := dailyLossWindow(day)
	if err != nil {
		m.err = err
		return m
	}
	read := g.dailyLossRows
	if read == nil {
		read = g.journal.DailyLossRows
	}
	rows, err := read(ctx, g.accountRef, from.Add(-dailyLossReadMargin), to.Add(dailyLossReadMargin))
	if err != nil {
		if errors.Is(err, journal.ErrDailyLossLedgerMalformed) {
			m.err = fmt.Errorf("%w: a closed managed position has a malformed closed_at", dailyloss.ErrUnmeasurable)
		} else {
			// 원문 오류는 싣지 않음(a131/a133 경계) — 원장을 읽지 못했다는 사실만.
			m.err = fmt.Errorf("%w: the journal could not be read", dailyloss.ErrUnmeasurable)
		}
		return m
	}
	closed := make([]dailyloss.Closed, 0, len(rows))
	for _, r := range rows {
		closed = append(closed, dailyloss.Closed{PositionID: r.PositionID, Market: r.Market, Symbol: r.Symbol, ClosedAt: r.ClosedAt,
			Managed: true, AdjustmentClosed: r.AdjustmentClosed, EngineSellDispatched: r.EngineSellDispatched, Outcome: r.Outcome})
	}
	m.losses, m.err = dailyloss.Collect(day, closed)
	return m
}

// dailyLossWindow 는 일 키 D 의 창 [D 08:00 KST, D+1 08:00 KST) 를 UTC 로 돌려줌 — clock.AccountTradingDay 와 같은 경계.
func dailyLossWindow(day string) (time.Time, time.Time, error) {
	loc, err := clock.MarketKR.Location()
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("%w: the Seoul time zone cannot be loaded", dailyloss.ErrUnmeasurable)
	}
	midnight, err := time.ParseInLocation("2006-01-02", day, loc)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("%w: the account day key is not a date", dailyloss.ErrUnmeasurable)
	}
	from := midnight.Add(clock.AccountDayRollover).UTC()
	return from, from.Add(24 * time.Hour), nil
}

// measuredDailyLoss 는 토큰을 확인하고 손실을 기준 통화로 환산함. fx 는 요청이 가진 봉인 authority(자기 시장 + 같은 파도 짝 FX)뿐 —
// 하루 매수 총액의 수동 매수 환산과 같은 묶음(qFinalManualBuyFX)이고, 재조회 · 호출자 환율 · 다른 파도는 들어올 길이 없음.
func (g *RiskGuardian) measuredDailyLoss(m DailyLossMeasurement, now time.Time, fx manualBuyFX) (riskcalc.Money, error) {
	if m.guardian == nil || m.guardian != g {
		return riskcalc.Money{}, fmt.Errorf("%w: no measurement taken by this Guardian accompanies the entry", dailyloss.ErrUnmeasurable)
	}
	if m.err != nil {
		return riskcalc.Money{}, m.err
	}
	// 신선도는 손실이 없는 날에도 봄 — riskcalc 는 항목마다만 재므로 빈 측정은 거기서 걸리지 않음. 한계는 riskcalc 의 계좌 스냅샷 한계.
	if age := now.Sub(m.at); age < 0 || age > riskcalc.AccountSnapshotStaleness {
		return riskcalc.Money{}, fmt.Errorf("%w: the measurement is %s old against a %s bound", dailyloss.ErrUnmeasurable,
			age.Round(time.Millisecond), riskcalc.AccountSnapshotStaleness)
	}
	day, err := clock.AccountTradingDay(now)
	if err != nil || day != m.dayKey {
		return riskcalc.Money{}, fmt.Errorf("%w: the measurement belongs to another account day", dailyloss.ErrUnmeasurable)
	}
	base := g.policy.LimitCurrency()
	convert := func(currency, magnitude string) (string, error) {
		market, err := marketOfCurrency(currency)
		if err != nil {
			return "", err
		}
		authority, ok := fx[market]
		if !ok {
			return "", fmt.Errorf("no %s authority in this request's paired FX collection", market)
		}
		bound, err := risk.BindAccountBaseFX(now, market, g.policy, authority)
		if err != nil {
			return "", err
		}
		// 올림(손실을 작게 만들지 않음) — 수동 매수 · 진입 노출과 같은 환산.
		converted, err := risk.AccountBaseOrderNotional(now, market, magnitude, bound)
		if err != nil {
			return "", err
		}
		return converted.Amount, nil
	}
	// AsOf 는 측정 순간, 평가 순간은 now — riskcalc 가 10초 신선 한계와 미래 시각을 판정함.
	return dailyloss.Total(m.losses, base, convert, m.at, now)
}

// dailyLossFloor 는 체인에 넣을 계좌 상태를 만듦. 돌려준 오류는 측정 실패이고, 그때 계좌의 일일 손실은 쓸 수 없는 값이 되어 체인의
// daily_loss 단이 INPUT_UNAVAILABLE 로 거절함(앞 단이 먼저 거절하면 그 사유가 그대로임).
func (g *RiskGuardian) dailyLossFloor(account risk.AccountState, m DailyLossMeasurement, now time.Time, fx manualBuyFX) (risk.AccountState, error) {
	measured, err := g.measuredDailyLoss(m, now, fx)
	if err != nil {
		account.DailyRealizedLoss = riskcalc.Money{}
		return account, err
	}
	caller, ok := validDailyLoss(account.DailyRealizedLoss, g.policy.LimitCurrency())
	if !ok {
		return account, nil // 무효한 호출자 값은 대체하지 않음 — 체인이 지금처럼 거절
	}
	if cmp, err := riskcalc.CompareDecimal(measured.Amount, caller); err == nil && cmp > 0 {
		account.DailyRealizedLoss = riskcalc.Money{Amount: measured.Amount, Currency: measured.Currency}
	}
	return account, nil
}

// validDailyLoss 는 호출자 값이 체인이 받는 모양(기준 통화 · 음수 아닌 decimal)인지 봄.
func validDailyLoss(m riskcalc.Money, base string) (string, bool) {
	if strings.ToUpper(strings.TrimSpace(m.Currency)) != base {
		return "", false
	}
	amount, err := riskcalc.CanonicalDecimal(strings.TrimSpace(m.Amount))
	if err != nil {
		return "", false
	}
	if negative, err := riskcalc.IsNegativeDecimal(amount); err != nil || negative {
		return "", false
	}
	return amount, true
}

// evaluateChainWithDailyLoss 는 체인을 한 번 돌리고, 측정 실패로 daily_loss 단이 입력 불가로 멈췄으면 그 상세를 측정 실패 문구로 바꿈.
func evaluateChainWithDailyLoss(in risk.Input, measureErr error) risk.Decision {
	verdict := evaluateChain(in)
	if !verdict.Allowed && measureErr != nil && verdict.Step == "daily_loss" && verdict.Reason == risk.ReasonInputUnavailable {
		verdict.Detail = "today's realised loss cannot be measured: " + strings.TrimPrefix(measureErr.Error(), dailyloss.ErrUnmeasurable.Error()+": ")
	}
	return verdict
}

// EscalateDailyLossLimit 은 전략 경로의 일일 손실 한도 거절을 운영 모드 강화(ENTRY_BLOCKED)로 영속함 — precheck 는 원장을 바꾸지 않는
// 계약이라 첫 다리 승인이 ctx 를 들고 부름. 성공하면 계좌 진입 차단 표지, 실패하면 거절과 강화 오류를 함께 든 보통 오류(조용히 삼켜지지
// 않게 — 그 경우 주기는 오늘처럼 오류로 끝남).
func (g *RiskGuardian) EscalateDailyLossLimit(ctx context.Context, refusal error) error {
	if err := g.escalateFor(ctx, risk.Decision{Reason: risk.ReasonDailyLossLimitReached}); err != nil {
		return errors.Join(refusal, err)
	}
	return &DailyLossEntryBlock{Cause: refusal}
}

// DailyLossEntryBlock 은 「일일 손실 한도로 계좌의 신규 진입이 막혔다」는 표지임. 전략 경로는 이것을 레인 결함이 아닌 정책 결과로 읽어
// 주기를 오류 없이 끝냄(레인 잠금 없음). 이것을 만드는 생산 자리는 셋뿐: 강화에 성공한 첫 다리 승인, AUTO 일일 손실 모드만 남은 진입
// 관문의 거절, 그 모드 변경이 원인인 제출 직전 펜스.
type DailyLossEntryBlock struct {
	Cause error
}

func (e *DailyLossEntryBlock) Error() string {
	if e == nil || e.Cause == nil {
		return "execgw: new entries are blocked by the daily-loss limit"
	}
	return "execgw: new entries are blocked by the daily-loss limit: " + e.Cause.Error()
}

func (e *DailyLossEntryBlock) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Cause
}

// Unwrap 은 계좌 진입 차단 표지만 드러냄(그 밖의 거절은 cause 가 nil 이라 오늘과 같음).
func (e *RejectedError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.cause
}

// dailyLossOnlyLocked 는 계좌의 진입 차단이 AUTO 일일 손실 모드 latch **하나뿐**이고 그 종목에 종목 latch 도 없는지 봄. g.mu 를 잡고 부름.
func (g *EntryGate) dailyLossOnlyLocked(market, symbol string) (string, bool) {
	if len(g.latches) != 1 || g.modeDailyLossDetail == "" {
		return "", false
	}
	detail, blocked := g.latches[ReasonOperatingModeBlocked]
	if !blocked || detail != g.modeDailyLossDetail {
		return "", false
	}
	if g.checkSymbolEntryLocked(market, symbol) != nil {
		return "", false
	}
	return detail, true
}

// markStrategyRefusal 은 진입 관문 거절이 「AUTO 일일 손실 모드만 남은 차단」이면 표지를 단 사본을 돌려주고, 아니면 그대로 돌려줌.
// 거절의 문구가 지금의 latch 문구와 같을 때만 표지를 닮(거절과 판정이 같은 상태를 가리키게).
func (g *EntryGate) markStrategyRefusal(rejected *RejectedError, market, symbol string) *RejectedError {
	if rejected == nil || rejected.Reason != ReasonOperatingModeBlocked {
		return rejected
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	detail, quiet := g.dailyLossOnlyLocked(market, symbol)
	if !quiet || detail != rejected.Detail {
		return rejected
	}
	marked := *rejected
	marked.cause = &DailyLossEntryBlock{Cause: &RejectedError{Reason: rejected.Reason, Detail: rejected.Detail}}
	return &marked
}

// dailyLossOnlyRejectionLocked 는 펜스 안(g.mu 잡음)에서 쓰는 같은 판정 — 표지를 단 모드 거절을 돌려주거나 nil.
func (g *EntryGate) dailyLossOnlyRejectionLocked(market, symbol string) *RejectedError {
	detail, quiet := g.dailyLossOnlyLocked(market, symbol)
	if !quiet {
		return nil
	}
	return &RejectedError{Reason: ReasonOperatingModeBlocked, Detail: detail,
		cause: &DailyLossEntryBlock{Cause: &RejectedError{Reason: ReasonOperatingModeBlocked, Detail: detail}}}
}

// strategyFenceRejection 은 제출 직전 펜스의 거절을 만듦 — 문구 · 사유는 오늘과 같고, 원인이 계좌 진입 차단이면 그 표지만 이어 붙임.
func strategyFenceRejection(err error) *RejectedError {
	rejected := reject(ReasonStrategyDispatchFenced, "strategy final authority changed before transport: %v", err)
	var block *DailyLossEntryBlock
	if errors.As(err, &block) {
		rejected.cause = block
	}
	return rejected
}

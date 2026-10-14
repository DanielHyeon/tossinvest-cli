// Package dailyloss 는 계좌 단위 하루(clock.AccountTradingDay)의 실현 손실을 원장 행에서 재는 순수 함수 둘을 둠(a137).
//
// # 왜 따로 패키지인가
//
// Guardian(execgw)은 원장 읽기와 환산 권한을 들고, 이 패키지는 「무엇을 손실로 세는가」만 정함 — 시계 · 파일 · 네트워크를 읽지 않음.
// 규칙(사용자 결정 2026-10-14, design D1)은 여기 한 곳에만 있음:
//
//   - 귀속은 계좌 일 키 하나: clock.AccountTradingDay(closed_at) == D. closed_at 은 원장 커밋 시각이라 시장 현지일로 귀속하면
//     늦게 기록된 손실이 이미 끝난 날로 떨어져 영원히 세어지지 않음(fail-open, freeze 리뷰 P1-1).
//   - 포지션 사이에서 이익은 손실을 상쇄하지 않음(Q12) — 음수 행만 합산. 한 포지션 수명 안의 순실현은 성과 행 그대로(U8).
//   - 관리 대상이 아닌 포지션은 세지 않음(U1). 사람의 앱 매도로 조정 종결된 관리 포지션도 세지 않음(Q13) — 단 수명 안에 엔진 매도가
//     브로커에 나갔을 수 있으면 측정 불가(P1-2).
//   - 그날 닫힌 관리 포지션의 성과 행 부재 · 해석 불가 값 · 형식이 어긋난 closed_at 은 측정 불가(Q3, fail-closed).
//
// 측정 불가는 진입만 막음 — 매도 · 손절 경로는 이 패키지를 부르지 않음.
package dailyloss

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/riskcalc"
)

// ErrUnmeasurable 는 그날의 손실을 정직하게 잴 수 없다는 뜻임. 진입은 일일 손실 단에서 입력 불가로 거절됨.
var ErrUnmeasurable = errors.New("dailyloss: today's realised loss cannot be measured")

// Closed 는 원장의 CLOSED 포지션 한 행과 측정에 필요한 사실들임.
type Closed struct {
	// PositionID 는 원장 id 원문 — 오류 문구에는 해시로만 실림(a131/a133 경계).
	PositionID string
	Market     string
	Symbol     string
	// ClosedAt 은 원장 커밋 시각(positions.closed_at) 원문.
	ClosedAt string
	// Managed 는 position.ExitEligible(진입 결정 또는 편입이 있음).
	Managed bool
	// AdjustmentClosed 는 마지막 조정이 수량을 0 으로 만들어 닫힌 포지션(사람의 앱 매도).
	AdjustmentClosed bool
	// EngineSellDispatched 는 그 인스턴스 수명과 겹쳐 엔진 매도가 브로커에 나갔을 수 있다는 기록 — 발송된 SELL PLACE attempt 또는
	// 브로커에 놓였을 수 있는 보호 주문(CREATE · REPLACE).
	EngineSellDispatched bool
	// Outcome 은 trade_outcomes.realized_pnl_after_costs(부호 있음). nil = 성과 행 없음.
	Outcome *string
}

// Losses 는 일 키 하나의 통화별 손실 크기 합(환산 전, 음수 아님)임.
type Losses struct {
	DayKey     string
	ByCurrency map[string]string
}

// unmeasurable 은 측정 불가 사유를 계좌 값 없이 싣음.
func unmeasurable(row Closed, why string) error {
	return fmt.Errorf("%w: %s %s %s — %s", ErrUnmeasurable, strings.ToUpper(strings.TrimSpace(row.Market)),
		strings.TrimSpace(row.Symbol), positionHandle(row.PositionID), why)
}

// positionHandle 은 원장 id 대신 싣는 짧은 손잡이임 — 같은 포지션이면 같은 값, 원문은 복원되지 않음.
func positionHandle(id string) string {
	var h uint32 = 2166136261
	for i := 0; i < len(id); i++ {
		h ^= uint32(id[i])
		h *= 16777619
	}
	return fmt.Sprintf("pos-%08x", h)
}

// canonicalClosedAt 는 closed_at 을 「Z 로 끝나는 정규 RFC3339 UTC」로만 받음(§4.2 P3-2 — 원장 읽기의 문자열 사전 필터와 같은 정의).
func canonicalClosedAt(s string) (time.Time, bool) {
	if !strings.HasSuffix(s, "Z") {
		return time.Time{}, false
	}
	t, err := time.Parse(time.RFC3339Nano, s)
	if err != nil || t.Location() != time.UTC || t.Format(time.RFC3339Nano) != s {
		return time.Time{}, false
	}
	return t, true
}

// marketCurrency 는 측정 대상 시장의 통화임. KR · US 밖은 측정 불가(P3-2 — 하나로 통일).
func marketCurrency(market string) (string, bool) {
	switch strings.ToUpper(strings.TrimSpace(market)) {
	case "KR":
		return "KRW", true
	case "US":
		return "USD", true
	}
	return "", false
}

// Collect 는 일 키 D 에 귀속되는 관리 포지션의 손실을 통화별로 모음.
//
// 형식이 어긋난 closed_at 은 일 키와 무관하게 측정 불가임 — 어느 날의 행인지 알 수 없으므로 오늘의 행일 수 있음.
func Collect(dayKey string, rows []Closed) (Losses, error) {
	dayKey = strings.TrimSpace(dayKey)
	if dayKey == "" {
		return Losses{}, fmt.Errorf("%w: no account day key", ErrUnmeasurable)
	}
	out := Losses{DayKey: dayKey, ByCurrency: map[string]string{}}
	for _, row := range rows {
		if !row.Managed {
			continue // U1 — 외부 · 수동 보유는 세지 않음
		}
		closedAt, ok := canonicalClosedAt(row.ClosedAt)
		if !ok {
			return Losses{}, unmeasurable(row, "closed_at is not a canonical RFC3339 UTC instant")
		}
		day, err := clock.AccountTradingDay(closedAt)
		if err != nil {
			return Losses{}, fmt.Errorf("%w: account day of a close: %v", ErrUnmeasurable, err)
		}
		if day != dayKey {
			continue
		}
		currency, ok := marketCurrency(row.Market)
		if !ok {
			return Losses{}, unmeasurable(row, "market is neither KR nor US")
		}
		if row.AdjustmentClosed {
			if row.EngineSellDispatched {
				return Losses{}, unmeasurable(row, "adjustment-closed while an engine sell may have reached the broker")
			}
			continue // Q13 — 사람의 앱 매도
		}
		if row.Outcome == nil {
			return Losses{}, unmeasurable(row, "closed today without a frozen outcome row")
		}
		amount, err := riskcalc.CanonicalDecimal(strings.TrimSpace(*row.Outcome))
		if err != nil {
			return Losses{}, unmeasurable(row, "outcome amount is not a decimal")
		}
		negative, err := riskcalc.IsNegativeDecimal(amount)
		if err != nil {
			return Losses{}, unmeasurable(row, "outcome amount is not a decimal")
		}
		if !negative {
			continue // Q12 — 이익 · 0 은 상쇄에 쓰지 않음
		}
		magnitude, err := riskcalc.SubDecimal("0", amount)
		if err != nil {
			return Losses{}, unmeasurable(row, "outcome amount cannot be summed")
		}
		sum, err := riskcalc.AddDecimal(orZero(out.ByCurrency[currency]), magnitude)
		if err != nil {
			return Losses{}, unmeasurable(row, "outcome amount cannot be summed")
		}
		out.ByCurrency[currency] = sum
	}
	return out, nil
}

func orZero(s string) string {
	if s == "" {
		return "0"
	}
	return s
}

// Convert 는 다른 통화의 손실 크기를 기준 통화로 환산함 — 봉인된 권한으로만, 올림(Guardian 이 제공).
type Convert func(currency, magnitude string) (string, error)

// Total 은 손실을 기준 통화로 환산해 riskcalc.DailyLoss 로 합산함. asOf 는 측정 순간, now 는 평가 순간 — riskcalc 가 계좌 스냅샷
// 신선 한계(10초)와 미래 시각을 판정함.
func Total(losses Losses, base string, convert Convert, asOf, now time.Time) (riskcalc.Money, error) {
	base = strings.ToUpper(strings.TrimSpace(base))
	currencies := make([]string, 0, len(losses.ByCurrency))
	for currency := range losses.ByCurrency {
		currencies = append(currencies, currency)
	}
	sort.Strings(currencies)
	sum := "0"
	for _, currency := range currencies {
		magnitude := losses.ByCurrency[currency]
		if currency != base {
			if convert == nil {
				return riskcalc.Money{}, fmt.Errorf("%w: a %s loss needs a sealed conversion authority", ErrUnmeasurable, currency)
			}
			converted, err := convert(currency, magnitude)
			if err != nil {
				return riskcalc.Money{}, fmt.Errorf("%w: a %s loss cannot be converted: %v", ErrUnmeasurable, currency, err)
			}
			magnitude = converted
		}
		next, err := riskcalc.AddDecimal(sum, magnitude)
		if err != nil {
			return riskcalc.Money{}, fmt.Errorf("%w: %v", ErrUnmeasurable, err)
		}
		sum = next
	}
	in := riskcalc.DailyLossInputs{Now: now, TradingDay: losses.DayKey, TargetCurrency: base}
	if cmp, err := riskcalc.CompareDecimal(sum, "0"); err == nil && cmp > 0 {
		in.Realized = []riskcalc.RealizedPnL{{TradingDay: losses.DayKey, Amount: "-" + sum, Currency: base,
			CostsDeducted: true, AsOf: asOf}}
	} else if err != nil {
		return riskcalc.Money{}, fmt.Errorf("%w: %v", ErrUnmeasurable, err)
	}
	total, err := riskcalc.DailyLoss(in)
	if err != nil {
		return riskcalc.Money{}, fmt.Errorf("%w: %v", ErrUnmeasurable, err)
	}
	return total, nil
}

package engine

import (
	"fmt"
	"strings"

	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/riskcalc"
)

// DailyBuyLimiter 는 Guardian 이 자기 하루 매수 총액 상한을 진술하는 면(a136). 생산 RiskGuardian 이 구현함.
type DailyBuyLimiter interface {
	DailyBuyNotionalLimit() riskcalc.Money
}

// checkGuardianDailyBuyLimit 은 하루 매수 총액 상한의 단일 출처 검사임.
//
//   - 설정 해석 값이 미설정(USD base 무값 등)이면 한도 부재 → ErrLimitsRequired(D5: 기동 거부).
//   - Guardian 이 진술하지 않거나 값·통화가 다르면 → ErrGuardianLimitsMismatch.
func checkGuardianDailyBuyLimit(guardian execgw.Guardian, status AutomationStatus) error {
	want := status.DailyBuyNotional
	currency := strings.ToUpper(strings.TrimSpace(status.Limits.Currency))
	if !want.Set {
		return fmt.Errorf("%w: max_daily_buy_notional is not set and limit_currency %q has no default; "+
			"set the daily buy ceiling explicitly", ErrLimitsRequired, currency)
	}
	limiter, ok := guardian.(DailyBuyLimiter)
	if !ok {
		return fmt.Errorf("%w: the injected Guardian (%T) does not report its daily buy ceiling",
			ErrGuardianLimitsMismatch, guardian)
	}
	declared := limiter.DailyBuyNotionalLimit()
	cmp, err := riskcalc.CompareDecimal(strings.TrimSpace(declared.Amount), limitString(want.Value))
	if err != nil || cmp != 0 || !strings.EqualFold(strings.TrimSpace(declared.Currency), currency) {
		return fmt.Errorf("%w: Guardian daily buy ceiling %s %s does not match the configured %s %s",
			ErrGuardianLimitsMismatch, declared.Amount, declared.Currency, limitString(want.Value), currency)
	}
	return nil
}

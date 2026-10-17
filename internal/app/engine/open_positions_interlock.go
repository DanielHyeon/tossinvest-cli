package engine

import (
	"fmt"

	"github.com/JungHoonGhae/tossinvest-cli/internal/config"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
)

// OpenPositionLimiter 는 Guardian 이 자기 최대 보유 포지션 수를 진술하는 면(a138). 생산 RiskGuardian 이 구현함.
type OpenPositionLimiter interface {
	MaxOpenPositionsLimit() int
}

// checkGuardianOpenPositionLimit 은 최대 보유 포지션 수의 단일 출처 검사임.
//
//   - 설정 해석 값이 1 이상의 정수가 아니면 한도 부재 → ErrLimitsRequired(기동 거부). 생산 기동은 Guardian factory 가 먼저 같은 판정으로
//     거부하므로 이 갈래는 주입 Guardian 에서만 닿음 — 그래도 단일 출처 대조의 한 몫이라 둠.
//   - Guardian 이 진술하지 않거나 값이 다르면 → ErrGuardianLimitsMismatch.
func checkGuardianOpenPositionLimit(guardian execgw.Guardian, status AutomationStatus) error {
	want, ok := config.WholeOpenPositions(status.MaxOpenPositions)
	if !ok {
		return fmt.Errorf("%w: max_open_positions resolves to %v; it must be a whole number of at least 1",
			ErrLimitsRequired, status.MaxOpenPositions)
	}
	limiter, ok := guardian.(OpenPositionLimiter)
	if !ok {
		return fmt.Errorf("%w: the injected Guardian (%T) does not report its open position maximum",
			ErrGuardianLimitsMismatch, guardian)
	}
	if declared := limiter.MaxOpenPositionsLimit(); declared != want {
		return fmt.Errorf("%w: Guardian open position maximum %d does not match the configured %d",
			ErrGuardianLimitsMismatch, declared, want)
	}
	return nil
}

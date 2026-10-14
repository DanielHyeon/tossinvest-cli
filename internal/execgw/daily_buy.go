package execgw

import (
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/riskcalc"
)

// DailyBuyNotionalLimit 는 이 Guardian 이 진입 예약에 싣는 하루 매수 총액 상한임(a136). 기동 인터록이 감사된 설정의
// 해석 값과 대조함(engine.DailyBuyLimiter) — 5필드 한도 스냅샷과 같은 단일 출처 규칙.
func (g *RiskGuardian) DailyBuyNotionalLimit() riskcalc.Money { return g.policy.MaxDailyBuyNotional }

// dailyBuyLimit 은 진입 발급 문 셋이 예약 요청에 싣는 값을 한 곳에서 만듦 — 문마다 따로 쓰면 하나가 빠지거나 다른 값을 실음.
// journal 은 EXPOSURE_RAISING 예약에 이 값이 없으면 거부하므로, 빠뜨린 문은 조용히 통과하지 못함.
func (g *RiskGuardian) dailyBuyLimit() *journal.DailyBuyLimit {
	return &journal.DailyBuyLimit{
		Amount:   g.policy.MaxDailyBuyNotional.Amount,
		Currency: g.policy.MaxDailyBuyNotional.Currency,
	}
}

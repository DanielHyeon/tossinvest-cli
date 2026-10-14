package execgw

import (
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/costs"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/riskcalc"
)

// DailyBuyNotionalLimit 는 이 Guardian 이 진입 예약에 싣는 하루 매수 총액 상한임(a136). 기동 인터록이 감사된 설정의
// 해석 값과 대조함(engine.DailyBuyLimiter) — 5필드 한도 스냅샷과 같은 단일 출처 규칙.
func (g *RiskGuardian) DailyBuyNotionalLimit() riskcalc.Money { return g.policy.MaxDailyBuyNotional }

// dailyBuyLimit 은 진입 발급 문 셋이 예약 요청에 싣는 값을 한 곳에서 만듦 — 문마다 따로 쓰면 하나가 빠지거나 다른 값을 실음.
// journal 은 EXPOSURE_RAISING 예약에 이 값(수동 매수 합계 포함)이 없으면 거부하므로, 빠뜨린 문은 조용히 통과하지 못함.
// 수동 매수 합계를 정할 수 없으면 오류 — 그 진입은 결정 · 예약 없이 거부됨.
func (g *RiskGuardian) dailyBuyLimit(now time.Time, fx manualBuyFX) (*journal.DailyBuyLimit, error) {
	manual, err := g.manualBuyTotal(now, fx)
	if err != nil {
		return nil, err
	}
	return &journal.DailyBuyLimit{
		Amount:    g.policy.MaxDailyBuyNotional.Amount,
		Currency:  g.policy.MaxDailyBuyNotional.Currency,
		ManualBuy: manual,
	}, nil
}

// qFinalManualBuyFX 는 q_final 요청이 가진 봉인 authority 를 시장별로 모음 — 자기 시장(FXAuthority)과 짝 FX 의 다른 시장
// (PairedFXAuthority). 다른 출처(재조회 · caller 환율 · 다른 파도)는 여기 들어올 길이 없음.
func qFinalManualBuyFX(request QFinalEntryIssuance, market costs.Market) manualBuyFX {
	fx := manualBuyFX{market: request.FXAuthority}
	other := costs.MarketUS
	if market == costs.MarketUS {
		other = costs.MarketKR
	}
	if request.PairedFXAuthority.Digest() != "" {
		fx[other] = request.PairedFXAuthority
	}
	return fx
}

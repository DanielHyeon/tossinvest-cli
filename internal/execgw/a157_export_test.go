package execgw

import (
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
)

// SealForTest 는 첫 다리 precheck 봉인의 내보내지 않은 칸(전략 계획 · 캠페인 · 주간 결속 · 발급 시각)을 시험에 드러냄(a157 구현 리뷰
// codex P3-2 · Opus P3-4 — 통과 봉인 값을 정확한 기대값과 대조하려고). 빌드된 제품에는 없음.
func (p QFinalCampaignFirstLegPrecheck) SealForTest() (journal.StrategyPlanRequest, journal.FirstLegCampaignRequest, *journal.WeeklyFirstLegReservationBinding, time.Time) {
	return p.strategy, p.campaign, p.weekly, p.issuedAt
}

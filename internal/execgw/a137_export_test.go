package execgw

import (
	"context"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
)

// SetDailyLossRowsForTest 는 일일 손실 측정의 원장 읽기를 시험 행으로 바꿈(a137) — 빌드된 제품에는 없음.
func (g *RiskGuardian) SetDailyLossRowsForTest(rows func(context.Context, string, time.Time, time.Time) ([]journal.DailyLossRow, error)) {
	g.dailyLossRows = rows
}

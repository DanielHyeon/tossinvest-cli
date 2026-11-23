package engine

// reconcile_block_report.go — RED 단계 껍데기(a194 3.1). 시험이 컴파일되고 단언에서 실패하게 하는 이름만 둠.

import (
	"context"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/reconcile"
)

const (
	factBlockedOperatorOnly = "blocked_operator_only"
	factBlockedStale        = "blocked_stale"
	blockedHoldingGrace     = 10 * time.Minute
	blockedReportsPerCycle  = 3
)

type blockedHolding struct {
	position       journal.Position
	market, symbol string
}

func (d *ReconcileDriver) reportBlockedWith(context.Context, []blockedHolding, []reconcile.Block, time.Time) {}

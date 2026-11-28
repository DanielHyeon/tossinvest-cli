package main

import (
	"fmt"
	"strings"

	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/reconcile"
)

// reconcileResolveRefusedCause 는 운영자 해제 근거가 정의되지 않은 활성 원인이 하나라도 있으면 그 원인을 이름으로 말하며 거절함(a195).
// 허용 원인 목록을 이 한 곳에만 둠 — 원인을 더할 때 고칠 자리를 세지 않게 하려는 것임.
// 빈 목록 거절은 호출하는 validate 클로저가 맡음.
func reconcileResolveRefusedCause(blocks []reconcile.Block) error {
	for _, block := range blocks {
		switch {
		case block.Cause == journal.ReconcileCauseQuantityMismatch:
			continue
		case block.Cause == journal.ReconcileCauseAttributionFailed &&
			block.Scope == reconcile.ScopeSymbol && strings.TrimSpace(block.Symbol) != "":
			// 투영이 거부한 체결의 종목 차단만 받음 — 계좌 전체 귀속 실패는 쓰는 자리가 없는 모양이라 규칙을 두지 않음(design D5).
			continue
		}
		return fmt.Errorf(
			"engine reconcile-resolve: active cause %s on %s requires cause-specific recovery evidence",
			block.Cause, block.Symbol)
	}
	return nil
}

package main

// verify_reconcile_expired.go 는 `tossctl verify reconcile --expired`(a164)의 cmd 쪽 by-id 위임이다. 판정은
// internal/verifylive/reconcile_expired.go 에 있고, 이 파일은 좁은 리더의 넷째 GET 하나만 더함.

import (
	"context"

	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
)

// ReconcileConditionalOrder 는 by-id 상세 GET 임(a164). 계좌 범위 읽기라 다른 두 메서드처럼 신원 재확인 뒤에만 위임함(P2-4) —
// 만료 모드에서는 이것이 첫 계좌 범위 호출이므로 Accounts 재조회가 그보다 먼저 옴.
func (r *reconcileReads) ReconcileConditionalOrder(ctx context.Context, id string) (official.ReconcileConditionalDetail, error) {
	if err := r.ensureIdentity(ctx); err != nil {
		return official.ReconcileConditionalDetail{}, err
	}
	return r.client.ReconcileConditionalOrder(ctx, id)
}

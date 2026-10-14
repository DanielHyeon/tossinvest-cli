//go:build tossos_testseams

package journal

import (
	"context"
	"fmt"
)

// InsertClosedManagedPositionForTest 는 진입 결정이 있는 CLOSED 포지션 하나와(pnl 이 비지 않으면) 그 동결 성과 행을 씀(a137 —
// 다른 패키지 시험이 실제 원장 읽기로 일일 손실 창을 재게). 태그 밖 빌드에는 없음.
func (j *Journal) InsertClosedManagedPositionForTest(ctx context.Context, id, account, market, symbol, opened, closed, pnl string) error {
	decision := "a137-seam-" + id
	if _, err := j.db.ExecContext(ctx, `INSERT INTO decisions
	  (id, account_ref, generation, safety_class, preimage_kind, risk_preimage,
	   risk_hash, client_order_id, limits_json, nonce, issued_at, expires_at)
	VALUES (?,?,?,?,?,?,?,?,?,?,?,?)`, decision, account, 0, SafetyClassExposureRaising, PreimageKindRiskIntent,
		`{"symbol":"`+symbol+`"}`, "hash-"+id, nil, `{}`, "nonce-"+id, opened, closed); err != nil {
		return fmt.Errorf("seam decision: %w", err)
	}
	if _, err := j.db.ExecContext(ctx, `INSERT INTO positions (id, account_ref, market, symbol, instance_seq, entry_decision_id, state,
		quantity, avg_price, opened_at, closed_at) VALUES (?,?,?,?,1,?,'CLOSED','0','0',?,?)`,
		id, account, market, symbol, decision, opened, closed); err != nil {
		return fmt.Errorf("seam position: %w", err)
	}
	if pnl == "" {
		return nil
	}
	if _, err := j.db.ExecContext(ctx, `INSERT INTO trade_outcomes (position_id, realized_pnl_after_costs, realized_r, initial_risk,
		initial_quantity, held_seconds, exit_ratchet_level, exit_rung, closed_at) VALUES (?,?, '0','1','1',0,NULL,NULL,?)`,
		id, pnl, closed); err != nil {
		return fmt.Errorf("seam outcome: %w", err)
	}
	return nil
}

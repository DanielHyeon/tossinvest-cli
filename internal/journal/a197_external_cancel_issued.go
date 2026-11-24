package journal

// a197_external_cancel_issued.go 는 손절을 막는 외부 매수 취소의 「이미 냈다」 판독임(a197 design D7 · Manager 결정 M-1).
//
// 표시는 새 상태 · 스키마가 아니라 원장의 CANCEL attempt 행 자체임: 같은 계좌 · 시장 · 종목에서 그 생산자(source)가 그 번호를 겨눈
// 취소 시도가 한 번이라도 기록됐으면 「냈다」. 상태(접수 · 확정 거절 · park · RECORDED 에서 멈춤)와 거래일은 묻지 않음 — 주문 생애당
// 한 번이라 날짜 경계를 넘은 같은 주문을 다시 취소하지 않음. 번호는 바이트 비교(불투명 식별자 — resolution.go 의 규칙).

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

// ExternalCancelIssued 는 source 가 낸 취소 시도가 같은 계좌 · 시장 · 종목에서 그 주문 번호를 겨눈 적이 있는지임.
func (j *Journal) ExternalCancelIssued(ctx context.Context, accountRef, market, symbol, orderID, source string) (bool, error) {
	var one int
	err := j.db.QueryRowContext(ctx, `SELECT 1 FROM mutation_attempts a JOIN intents i ON i.id = a.intent_id
		 WHERE a.kind = ? AND a.target_order_id = ? AND TRIM(i.source) = ?
		   AND TRIM(i.account_ref) = ? AND LOWER(TRIM(i.market)) = ? AND UPPER(TRIM(i.symbol)) = ?
		 LIMIT 1`,
		string(KindCancel), orderID, strings.TrimSpace(source),
		strings.TrimSpace(accountRef), strings.ToLower(strings.TrimSpace(market)), strings.ToUpper(strings.TrimSpace(symbol))).Scan(&one)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("journal: reading the issued cancels of order %s: %w", orderID, err)
	}
	return true, nil
}

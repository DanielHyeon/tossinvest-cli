package journal

// a197_external_cancel_issued.go 는 손절을 막는 외부 매수 취소의 「이미 냈다」 판독임(a197 design D7 · Manager 결정 M-1).

import "context"

// ExternalCancelIssued 는 source 가 낸 취소 시도가 같은 계좌 · 시장 · 종목에서 그 주문 번호를 겨눈 적이 있는지임(골격 — a197 3.1 RED).
func (j *Journal) ExternalCancelIssued(ctx context.Context, accountRef, market, symbol, orderID, source string) (bool, error) {
	return false, nil
}

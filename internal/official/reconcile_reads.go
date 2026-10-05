package official

// reconcile_reads.go 는 a121(reconcile-stale-verification-artifacts) 대사 전용 읽기의 **골격**이다.
//
// RED 로트 산출물: 타입·시그니처만 세우고 동작은 비워 둔다(모든 호출이 errReconcileReadUnimplemented 반환).
// 구현은 GREEN 로트(tasks 3.2)의 몫이다.
//
// # 왜 기존 읽기(ProtectionConditionalOrdersRaw · OrdersPageRaw)를 그대로 쓰지 않는가
//
// design G1-6 codex F2·R2-1 — 기존 경로는 unwrapAndDecode 로 구조체에 바로 풀어서 `{"result":null}`·`{}`·
// null 값 필드를 무오류 빈 페이지로 접는다(client.go:213-227). 대사는 "빈 목록" 을 부재의 근거로 쓰므로
// 결측·null·형 불일치를 따로 거절해야 하고, 그러려면 result 원문 바이트를 받아야 한다. 또 codex F6 — 채택
// 리더 adaptProtectionConditional 은 Second 를 버린다(protection_reads.go:49-57). 같은 endpoint·같은
// getAcct 경로를 쓰되 해독만 새로 한다. 기존 함수는 편집하지 않는다.

import (
	"context"
	"errors"
)

// ErrReconcileSchema 는 대사 읽기 응답이 스키마 존재·형 검증을 통과하지 못했음을 뜻함.
// 빈 목록으로 접지 않고 거절로 다루기 위한 표지임.
var ErrReconcileSchema = errors.New("official: reconcile read failed schema validation")

// errReconcileReadUnimplemented 는 RED 골격의 자리표시 오류임 — ErrReconcileSchema 와 일부러 다름.
var errReconcileReadUnimplemented = errors.New("official: a121 reconcile read is not implemented (RED skeleton)")

// ReconcileConditionalRow 는 대사 판정에 쓰는 조건주문 한 행임.
// HasSecond 는 응답의 second 가 null 이 아닌지(OCO/OTO 다리 존재)를 그대로 노출함(codex F6).
type ReconcileConditionalRow struct {
	ID               string
	Symbol           string
	Market           string
	Status           string
	TriggeredOrderID string
	HasSecond        bool
}

// ReconcileConditionalPage 는 조건주문 목록 한 페이지임.
type ReconcileConditionalPage struct {
	Rows       []ReconcileConditionalRow
	NextCursor string
	HasNext    bool
}

// ReconcileOrderRow 는 대사 판정에 쓰는 일반 주문 한 행임(일반 주문 응답에는 market 이 없음).
type ReconcileOrderRow struct {
	ID     string
	Symbol string
	Status string
}

// ReconcileOrderPage 는 일반 주문 목록 한 페이지임.
type ReconcileOrderPage struct {
	Rows       []ReconcileOrderRow
	NextCursor string
	HasNext    bool
}

// DecodeReconcileConditionalPage 는 조건주문 목록의 result 원문을 스키마 검증과 함께 해독함.
// 골격: 항상 자리표시 오류 반환.
func DecodeReconcileConditionalPage(result []byte) (ReconcileConditionalPage, error) {
	return ReconcileConditionalPage{}, errReconcileReadUnimplemented
}

// DecodeReconcileOrderPage 는 일반 주문 목록의 result 원문을 스키마 검증과 함께 해독함.
// 골격: 항상 자리표시 오류 반환.
func DecodeReconcileOrderPage(result []byte) (ReconcileOrderPage, error) {
	return ReconcileOrderPage{}, errReconcileReadUnimplemented
}

// ReconcileConditionalOrdersPage 는 GET /api/v1/conditional-orders 한 페이지를 대사용으로 읽음.
// 골격: 네트워크 호출 없이 자리표시 오류 반환.
func (c *Client) ReconcileConditionalOrdersPage(ctx context.Context, status, symbol, cursor string, limit int) (ReconcileConditionalPage, error) {
	return ReconcileConditionalPage{}, errReconcileReadUnimplemented
}

// ReconcileOpenOrdersPage 는 GET /api/v1/orders?status=OPEN 한 페이지를 대사용으로 읽음.
// 골격: 네트워크 호출 없이 자리표시 오류 반환.
func (c *Client) ReconcileOpenOrdersPage(ctx context.Context, symbol, cursor string, limit int) (ReconcileOrderPage, error) {
	return ReconcileOrderPage{}, errReconcileReadUnimplemented
}

// ReconcileInstrument 는 종목 조회 GET(양성 대조, design G1 P1-3)으로 응답이 되돌린 심볼을 돌려줌.
// 골격: 네트워크 호출 없이 자리표시 오류 반환.
func (c *Client) ReconcileInstrument(ctx context.Context, symbol string) (string, error) {
	return "", errReconcileReadUnimplemented
}

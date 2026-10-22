package official

// reconcile_detail_reads.go 는 a164(verify-reconcile-closes-broker-expired-conditional) 만료 대사 전용 by-id 상세 읽기임.
//
// 같은 endpoint(GET /api/v1/conditional-orders/{id})·같은 getAcct 경로를 쓰되 result 원문을 받아 스키마를 따로 검증하고,
// 404 중 본문 code 가 정확히 conditional-order-not-found 인 것만 「silent」 로 가름(design D3 「by-id 응답 분류」).
// 기존 ConditionalOrder·ConditionalOrderRaw 는 null 을 제로값으로 접고(F2) 404 를 오류로만 돌려주므로 쓰지 않음.

import (
	"context"
	"errors"
	"strings"
)

// ReconcileConditionalDetail 은 by-id 상세 GET 한 번의 분류 결과임.
// Silent 면 404 conditional-order-not-found(진술 아님 — 근거로 세지 않음) 이고 나머지 판정 필드는 비어 있음.
type ReconcileConditionalDetail struct {
	Silent bool
	// RequestID 는 silent 응답 본문의 requestId 임(계좌 정보 없음) — 거절 문구에만 씀.
	RequestID        string
	ID               string
	Type             string
	Status           string
	Symbol           string
	Market           string
	FirstStatus      string
	TriggeredOrderID string
	HasSecond        bool
	// ResultDigest 는 statement 의 result 원문 지문임(sha256 16바이트 절단). 성공 응답 requestId 를 클라이언트가
	// 노출하지 않아서 그 대신 적음(design R3).
	ResultDigest string
}

// errReconcileDetailUnimplemented 는 골격 단계 표지임 — by-id 읽기가 아직 판정 없이 서 있음.
var errReconcileDetailUnimplemented = errors.New("official: the reconcile by-id read is not implemented yet")

// ReconcileConditionalOrder 는 GET /api/v1/conditional-orders/{id} 를 대사용으로 읽음(계좌 헤더 포함).
func (c *Client) ReconcileConditionalOrder(ctx context.Context, id string) (ReconcileConditionalDetail, error) {
	if strings.TrimSpace(id) == "" {
		return ReconcileConditionalDetail{}, reconcileSchema("the by-id read names no identifier")
	}
	return ReconcileConditionalDetail{}, errReconcileDetailUnimplemented
}

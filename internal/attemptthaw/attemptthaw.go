// Package attemptthaw 는 a094 D−4.5 park 해동(운영자 해소) 명령의 전송 계약(요청 · 결과 · 오류 어휘)임.
//
// journal · engine 어느 쪽도 import 하지 않음 — tossctl 과 엔진 제어 endpoint 가 같은 모양을 공유하기 위한 자리일 뿐이고,
// 판정은 전부 엔진 프로세스 안에서 함(riskrelaxation 과 같은 구성, a066/a092 완화 명령 가족 계약): 원장은 단일 writer(엔진)
// 이고 CLI 가 원장을 직접 열면 새 바이너리가 도는 엔진 밑에서 마이그레이션할 수 있음. 엔진이 없으면 거절함. 엔진을 멈추게
// 하지 않음 — 엔진 정지는 손절 정지임.
package attemptthaw

import "errors"

var (
	// ErrInvalidRequest 는 모양이 틀린 요청임(attempt · 목표 · 운영자 · 승인 참조 · note · 접수 확정의 주문 번호 누락).
	ErrInvalidRequest = errors.New("attempt thaw: invalid request")
	// ErrStale 는 명령 시점에 대상 attempt 가 park(UNRESOLVED_IN_DOUBT) 상태가 아님 — 아무것도 바뀌지 않음.
	ErrStale = errors.New("attempt thaw: the attempt is not parked (UNRESOLVED_IN_DOUBT) any more")
	// ErrNotFound 는 그 attempt 가 원장에 없음.
	ErrNotFound = errors.New("attempt thaw: no such attempt")
	// ErrAuditUnavailable 는 audit 줄을 먼저 쓸 수 없음 — 원장 전이 전이므로 아무것도 바뀌지 않음.
	ErrAuditUnavailable = errors.New("attempt thaw: the resolution could not be audited, so nothing was changed")
	// ErrUnwired 는 이 명령을 제공하지 않는 엔진(a094 이전 빌드)임.
	ErrUnwired = errors.New("attempt thaw: the running engine does not offer this resolution")
)

// Target 은 운영자가 확인한 종결임.
const (
	TargetFailedConfirmed = "FAILED_CONFIRMED" // 브로커에 그 주문이 없음(비수용) — 발의를 같은 명령 안에서 해제함
	TargetConfirmed       = "CONFIRMED"        // 브로커에 그 주문이 있음 — 주문 번호 필수, 발의는 체결 경로가 끝냄
)

// Request 는 park 된 attempt 하나의 운영자 해소 요청임.
type Request struct {
	AttemptID     string `json:"attempt_id"`
	Target        string `json:"target"`
	BrokerOrderID string `json:"broker_order_id,omitempty"`
	Operator      string `json:"operator"`
	Approval      string `json:"approval"`
	Note          string `json:"note"`
}

// Result 는 커밋된 해소임. ProposalReleased 는 그 attempt 가 무장한 발의를 같은 명령 안에서 해제했는지임 —
// 거짓이고 ReleaseError 가 있으면 해소는 유효하고 해제만 실패했음(다음 기동 따라잡기가 푼다).
type Result struct {
	AttemptID        string `json:"attempt_id"`
	State            string `json:"state"`
	PositionID       string `json:"position_id,omitempty"`
	ProposalReleased bool   `json:"proposal_released"`
	ReleaseError     string `json:"release_error,omitempty"`
}

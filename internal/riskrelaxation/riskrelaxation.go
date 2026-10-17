// Package riskrelaxation 은 a066 5.5(design D8) 운영자 해제의 전송 계약(요청·결과·오류 어휘)임.
//
// journal·engine 어느 쪽도 import 하지 않음 — tossctl 과 엔진 제어 endpoint 가 같은 모양을 공유하기 위한 자리일 뿐이고,
// 해제 판정은 전부 엔진 프로세스 안의 journal API(ReleaseEntryLossLock · ReleaseRiskOverageLatch)가 함(exitquarantine
// 패키지와 같은 구성, a079 선례).
//
// 경로 원칙(Manager 판정 2026-09-29, a092 완화 명령 가족 계약): 해제는 엔진 제어 endpoint 를 거침 — 원장은 단일 writer
// (엔진)이고, CLI 가 원장을 직접 열면 새 바이너리가 도는 엔진 밑에서 마이그레이션할 수 있음. 엔진이 없으면 거절함
// (엔진이 없으면 진입도 없음).
package riskrelaxation

import "errors"

var (
	// ErrInvalidRequest 는 모양이 틀린 요청임(범위·결속 값·승인·운영자·사유 누락).
	ErrInvalidRequest = errors.New("risk relaxation: invalid request")
	// ErrStale 는 승인자가 본 상태 뒤에 상태가 바뀐 해제임 — 현재 상태로 다시 승인받아야 함(보수 쪽 승리).
	ErrStale = errors.New("risk relaxation: the state changed after the approver read it")
	// ErrAuditUnavailable 는 해제를 audit 로그에 기록할 수 없음(엔진에 audit 로그가 없거나 쓰기 실패) — 커밋 전이므로
	// 아무것도 바뀌지 않음.
	ErrAuditUnavailable = errors.New("risk relaxation: the release could not be audited, so nothing was changed")
	// ErrStateMismatch 는 owner 원장이 자기 마지막 봉인과 맞지 않아 결속을 검증할 수 없는 해제임 — 아무것도 바뀌지
	// 않음. 재시도로 풀리지 않음(상태를 다시 봉인하는 사건이 먼저 필요함).
	ErrStateMismatch = errors.New("risk relaxation: the owner state does not match its seal, so the binding cannot be verified")
	// ErrUnwired 는 이 해제를 제공하지 않는 엔진(a066 5.5 이전 빌드)임.
	ErrUnwired = errors.New("risk relaxation: the running engine does not offer this release")
)

// EntryLockReleaseRequest 는 진입 손실 잠금 하나의 해제 요청임. LockSeq·ExpectedLastEvent 는 `engine risk-latch-show`
// 가 보여 준 값(승인자가 본 상태)임.
type EntryLockReleaseRequest struct {
	AccountRef        string `json:"account_ref"`
	Market            string `json:"market"`
	Horizon           string `json:"horizon"`
	LockSeq           int64  `json:"lock_seq"`
	ExpectedLastEvent int64  `json:"expected_last_event"`
	Operator          string `json:"operator"`
	Approval          string `json:"approval"`
	Reason            string `json:"reason"`
}

// LatchReleaseRequest 는 owner generation 하나의 RISK_OVERAGE latch 해제 요청임. ExpectedState 는 show 가 보여 준
// owner 상태 봉인 digest 임.
type LatchReleaseRequest struct {
	AccountRef    string `json:"account_ref"`
	Market        string `json:"market"`
	Symbol        string `json:"symbol"`
	Generation    string `json:"generation"`
	ExpectedState string `json:"expected_state"`
	Operator      string `json:"operator"`
	Approval      string `json:"approval"`
	Reason        string `json:"reason"`
}

// Result 는 커밋된 해제임. Notified=false 는 「완화됨·통지 실패」 — 해제는 유효하고, 커밋 뒤 alert 기록만 실패했음.
type Result struct {
	ReleaseSeq  int64  `json:"release_seq"`
	Target      string `json:"target"`
	ReleasedAt  string `json:"released_at"`
	Notified    bool   `json:"notified"`
	NotifyError string `json:"notify_error,omitempty"`
}

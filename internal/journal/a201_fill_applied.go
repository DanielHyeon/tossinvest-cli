package journal

// ReconcileReleaseFillApplied 는 대사의 로컬 읽기 뒤에 커밋된 체결이 조정을 대신했고 그 뒤의 새 대사가 일치해 수량 불일치
// 차단을 푼 원인임(a201 design V6).
//
// ADJUSTMENT_APPLIED 와 짝이지만 근거가 다름 — 조정 tx 가 아니라 체결이 투영을 계좌 쪽으로 옮겼음. 그래서 「보유 0 의
// 근거」가 아니므로 위험 버킷 broker-zero 허용 목록(RECHECK_MATCHED · ADJUSTMENT_APPLIED)에는 넣지 않음 — 그 경로는 이
// 원인으로 풀린 행을 released_reconciliation 으로 거절함(fail-closed). release_cause 에는 CHECK 가 없어 스키마는 무변.
const ReconcileReleaseFillApplied = "FILL_APPLIED"

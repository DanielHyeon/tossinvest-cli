package obs

import "github.com/JungHoonGhae/tossinvest-cli/internal/journal"

// isPreemption 은 정산 · 반납 결과가 「원장이 이름을 준 선점」인지 — 승인(이미 정산됨) 또는 다른 발송자(임차 상실)만.
// 행 없음 · 모르는 결과는 선점이 아님(a092 델타 「행 없음·모르는 결과·원장 오류는 선점이 아니다」, 26라운드 codex P2 #6):
// 동기 발송의 시도 기록 · 반납 두 자리가 이 한 판정을 씀(판정이 둘이면 서로를 가림).
func isPreemption(outcome journal.SettleOutcome) bool {
	return outcome == journal.SettleAlreadySettled || outcome == journal.SettleLeaseLost
}

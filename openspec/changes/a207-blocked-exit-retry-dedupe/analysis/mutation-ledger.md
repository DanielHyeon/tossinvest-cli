# a207 변이 원장 (design §7.2)

- 하네스: `analysis/a207_mutation.py` — 별도 detached worktree `~/tossos-work/a207-mut`(사본)에서만 변이를 걸고, 시작 HEAD 단언 · 깨끗함 단언 ·
  **무변이 대조군 GREEN** 확인 뒤 변이마다 원문 치환 1 회를 단언하고 `git checkout --` 로 되돌린 뒤 다시 깨끗함을 단언함. 원본 worktree 무접촉.
- 시험 명령: journal `go test ./internal/journal/ -count=1 -run 'A207|TestExitSnapshotDuplicateDecisionIsNotRearmed|TestARefusalReArmsTheLevel'`,
  engine `go test ./internal/app/engine/ -count=1 -run 'A207|TestARefusedProposalReleasesTheLevelAndAlerts'`.

## 1 회차 — HEAD `6f58010bf` (2026-11-27)

대조군 journal rc=0 · engine rc=0. 15 변이 중 CAUGHT 13 · **SURVIVED 2**:
M6(압축 갈래의 action 조건 제거) · M6b(출처 recomputed 조건 제거). 원인: 그 갈래의 조건 넷(출처 · action · intent · arm suppression)이 서로를 가려,
하나를 지워도 나머지가 같은 행을 막음(saved_monotone 은 발의를 지우고, 무장은 늘 recomputed, 무장 행은 늘 intent 를 가짐) — 「우연이 지키는 안전」.
처리(2단계 리뷰 r1 A-P2 와 같은 지적): 조건을 `row.ProposedIntentID != ""` 하나로 줄이고 술어 호출 계수 시험을 더함 → 커밋 `092b15d48`.
M10(술어를 트랜잭션 밖 연결에서 읽음)은 **NA — 구조상 불가**: `repeatsARefusedArmTx` 는 `*sql.Tx` 만 받고 `Journal` 수신자가 없어 다른 연결을 쥘 수 없음.

## 2 회차 — HEAD `092b15d48` (수리 뒤, 최종)

control journal: rc=0
control engine: rc=0
| # | 변이 | 파일 | 결과 | 첫 실패 |
|---|---|---|---|---|
| M1 | 해제 훅 확인 제거 | `internal/journal/a207_compact_evidence.go` | CAUGHT | --- FAIL: TestA207TheRepeatPredicateNeedsEveryTerm (0.26s) |
| M2 | 직전 무장 action 비교 제거 | `internal/journal/a207_compact_evidence.go` | CAUGHT | --- FAIL: TestA207TheRepeatPredicateNeedsEveryTerm (0.24s) |
| M3 | 직전 무장 level_after 비교 제거 | `internal/journal/a207_compact_evidence.go` | CAUGHT | --- FAIL: TestA207TheRepeatPredicateNeedsEveryTerm (0.24s) |
| M4 | 직전 무장 lifecycle_generation 비교 제거 | `internal/journal/a207_compact_evidence.go` | CAUGHT | --- FAIL: TestA207TheRepeatPredicateNeedsEveryTerm (0.27s) |
| M5 | 압축 행에서 decision_id 도 NULL | `internal/journal/exit_state.go` | CAUGHT | --- FAIL: TestA207ARepeatedRefusedArmIsWrittenCompact (0.29s) |
| M6 | 무장 행 한정 조건 제거(판정 전용 · legacy 행도 술어를 물음) | `internal/journal/exit_state.go` | CAUGHT | --- FAIL: TestA207OnlyArmRowsAskTheRepeatPredicate (0.24s) |
| M7 | 리더가 표지 없는 JSON 없는 행도 압축으로 읽음 | `internal/journal/exit_snapshot.go` | CAUGHT | --- FAIL: TestA207ReadersReadACompactRowAsARepeat (0.30s) |
| M7b | 압축 모양 검사에서 JSON NULL 요구 제거 | `internal/journal/a207_compact_evidence.go` | CAUGHT | --- FAIL: TestA207AMarkedRowWithTheWrongShapeIsCorrupt (0.95s) |
| M8 | 해소 알림을 셈과 무관하게 냄 | `internal/app/engine/a207_unblock_notice.go` | CAUGHT | --- FAIL: TestA207AStreakDoesNotCrossALifecycleGeneration (0.00s) |
| M9 | 해소 알림을 주기 끝이 아니라 확정 갈래에서 바로 냄 | `internal/app/engine/a207_unblock_notice.go` | CAUGHT | --- FAIL: TestA207AStreakDoesNotCrossALifecycleGeneration (0.00s) |
| M15 | 해소 알림의 세대 비교 제거 | `internal/app/engine/a207_unblock_notice.go` | CAUGHT | --- FAIL: TestA207AStreakDoesNotCrossALifecycleGeneration (0.00s) |
| M15b | 연속 해제 셈의 세대 경계 제거 | `internal/app/engine/a207_unblock_notice.go` | CAUGHT | --- FAIL: TestA207AStreakDoesNotCrossALifecycleGeneration (0.00s) |
| M11 | 술어 조회 오류를 판정 오류로 돌려줌 | `internal/journal/exit_state.go` | CAUGHT | --- FAIL: TestA207ALookupFaultWritesTheFullRowAndStillArms (0.25s) |
| M12 | 원장이 해제를 거절해도 셈을 더함 | `internal/app/engine/exitloop.go` | CAUGHT | --- FAIL: TestA207AReleaseTheLedgerRefusesIsNotCounted (0.26s) |
| M13 | 주기 끝의 해소 알림 넘기기 제거 | `internal/app/engine/exit_order_log.go` | CAUGHT | --- FAIL: TestA207AStopRefusedForNCyclesIsSubmittedOnTheFirstClearCycle (0.39s) |
| M14 | 확정 갈래의 해소 대기열 넣기 제거 | `internal/app/engine/exitloop.go` | CAUGHT | --- FAIL: TestA207AStopRefusedForNCyclesIsSubmittedOnTheFirstClearCycle (0.36s) |
final: copy clean, HEAD 092b15d486fd05ec9bed5c1714f79a1e0ed87f93

- 16 변이 전부 CAUGHT, SURVIVED 0. M6b 는 지운 조건이라 2 회차에서 뺌. M15 · M15b 는 리뷰 r1 B-P2 수리(세대 경계)의 변이.
- M9 의 첫 실패가 2 회차에서는 내부 시험(대기열 길이)으로 바뀜 — 1 회차에서는 순서 시험 `TestA207UnblockNoticesFollowEveryPlaceOfTheCycle` 이 잡았음(같은 변이).
- 원장 뒤 변경: `092b15d48` 다음 커밋은 `appendExitEventTx` 의 **주석 두 줄**만 고침(리뷰 r2 P3) — 변이 대상 코드 무변경.

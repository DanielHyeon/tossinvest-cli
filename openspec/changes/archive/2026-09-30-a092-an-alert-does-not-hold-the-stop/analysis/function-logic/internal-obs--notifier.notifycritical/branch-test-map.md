# Branch Test Map: `Notifier.notifyCritical`

- Source: `internal/obs/notifier.go` (:195-250); **편집 뒤** 측정 — `analysis/harness/coverage-post-r26d-obs.json`(연결 워크트리 `15b64676`, `internal/obs` 시험 118개를 하나씩, 실패 0). 편집 전 번들은 `analysis/pre-edit/r26b/internal-obs--notifier.notifycritical/`에 보존.
- 재번호: 없음.

| Branch | AST anchor | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | if at 196:2 | 원장 없음 | `TestA092DropAndNoJournalLinesKeepTheirOwnEventKey`, `TestCriticalWithoutAJournalIsLoudRatherThanSilent` | 해당 없음(분기 불변) | 블록 196.22-200.19을 시험 3개가 실행, PASS |
| B2 | if at 200:3 | 로그 | `TestA092DropAndNoJournalLinesKeepTheirOwnEventKey`, `TestCriticalWithoutAJournalIsLoudRatherThanSilent` | 전용 변이 없음 — `TestA092DropAndNoJournalLinesKeepTheirOwnEventKey` 가 이 줄의 event 키 수를 셈(편집 전 FAIL · `red-r26b-obs.log`) | 블록 200.19-204.4을 시험 3개가 실행, PASS |
| B3 | if at 220:2 | claim 실패 → 승격 | `TestAClaimThatFailsAttemptsTheDurableBlock`, `TestAClaimThatFailsBlocksNewEntries` | 해당 없음(분기 불변) | 블록 220.16-240.3을 시험 4개가 실행, PASS |
| B4 | if at 242:2 | 미전달 → judge | `TestA092AFailedEscalationLatchKeepsTheAccountOut`, `TestA092AFailedEscalationLatchesUnconditionally` | 해당 없음(분기 불변) | 블록 242.19-248.3을 시험 23개가 실행, PASS |

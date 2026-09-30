# Branch Test Map: `Notifier.notifyCritical`

- Source: `internal/obs/notifier.go` (:191-246); **편집 뒤** 측정 — `analysis/harness/coverage-post-r25fix.json`(연결 워크트리 `55963f29`, `./internal/obs` 시험 99개를 하나씩). 편집 전 번들은 `analysis/pre-edit/unit3/`에 보존.
- 재번호: 없음.

| Branch | AST anchor | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | if at 195:2 | 원장 없음 | `TestCriticalWithoutAJournalIsLoudRatherThanSilent`, `TestTheTransitionLogLineIsCountable` | 해당 없음(분기 불변) | 블록 195.22-199.19을 시험 2개가 실행, PASS |
| B2 | if at 199:3 | 로그 | `TestCriticalWithoutAJournalIsLoudRatherThanSilent`, `TestTheTransitionLogLineIsCountable` | 해당 없음(분기 불변) | 블록 199.19-203.4을 시험 2개가 실행, PASS |
| B3 | if at 219:2 | claim 실패 → 승격 | `TestAClaimThatFailsAttemptsTheDurableBlock`, `TestAClaimThatFailsBlocksNewEntries` | 해당 없음(분기 불변) | 블록 219.16-239.3을 시험 4개가 실행, PASS |
| B4 | if at 241:2 | 미전달 → judge | `TestA092AFailedEscalationLatchesUnconditionally`, `TestA092AReleaseAfterTheEpochReadIsHonoured` | 변이 L02 · L06 · L07 · L08 | 블록 241.19-247.3을 시험 22개가 실행, PASS |

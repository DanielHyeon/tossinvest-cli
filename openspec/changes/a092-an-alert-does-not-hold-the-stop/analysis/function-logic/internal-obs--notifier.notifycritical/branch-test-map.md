# Branch Test Map: `Notifier.notifyCritical`

- Source: `internal/obs/notifier.go`; **편집 전** 측정 — `analysis/harness/coverage-pre-unit3.json`(연결 워크트리 `b3f14925`, `./internal/obs` 시험 92개를 하나씩).

| Branch | AST anchor | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | if at 177:2 | 원장 없음 | `TestCriticalWithoutAJournalIsLoudRatherThanSilent`, `TestTheTransitionLogLineIsCountable` | 편집 전(기준선) | 블록 177.22-181.19을 시험 2개가 실행, PASS |
| B2 | if at 181:3 | 로그 | `TestCriticalWithoutAJournalIsLoudRatherThanSilent`, `TestTheTransitionLogLineIsCountable` | 편집 전(기준선) | 블록 181.19-185.4을 시험 2개가 실행, PASS |
| B3 | if at 201:2 | claim 실패 → 승격 | `TestAClaimThatFailsAttemptsTheDurableBlock`, `TestAClaimThatFailsBlocksNewEntries` | 편집 전(기준선) | 블록 201.16-221.3을 시험 4개가 실행, PASS |
| B4 | if at 223:2 | 미전달 → 승격 | `TestACancelledSenderStillHandsTheLeaseBack`, `TestACriticalAlertStillEscalatesThroughTheSameNotifier` | 편집 전(기준선) | 블록 223.19-229.3을 시험 17개가 실행, PASS |

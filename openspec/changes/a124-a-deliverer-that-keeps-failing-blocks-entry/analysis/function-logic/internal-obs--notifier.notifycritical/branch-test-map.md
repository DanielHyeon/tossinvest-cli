# Branch Test Map: `Notifier.notifyCritical`

- Source: `internal/obs/notifier.go` (176-231); current (a124 는 편집하지 않음 — 대조 번들)
- 시험 칸은 측정값: `analysis/harness/branch_coverage.py` 가 `github.com/JungHoonGhae/tossinvest-cli/internal/obs` 의 시험 81 개를 하나씩 돌린 커버 프로필에서 그 분기 본문 블록을 실행한 시험.

| Branch | AST anchor | Scenario (source text) | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | if at 177:2 | `if n.Journal == nil {` | `TestCriticalWithoutAJournalIsLoudRatherThanSilent` · `TestTheTransitionLogLineIsCountable` | 대조(편집 없음) | 블록 177.22-181.19 을 시험 2 개가 실행, 전부 PASS |
| B2 | if at 181:3 | `if n.Log != nil {` | `TestCriticalWithoutAJournalIsLoudRatherThanSilent` · `TestTheTransitionLogLineIsCountable` | 대조(편집 없음) | 블록 181.19-185.4 을 시험 2 개가 실행, 전부 PASS |
| B3 | if at 201:2 | `if err != nil {` | `TestAClaimThatFailsAttemptsTheDurableBlock` · `TestAClaimThatFailsBlocksNewEntries` · `TestAFailedClaimStillReturnsItsError` (+1) | 대조(편집 없음) | 블록 201.16-221.3 을 시험 4 개가 실행, 전부 PASS |
| B4 | if at 223:2 | `if owed && !sent {` | `TestACancelledSenderStillHandsTheLeaseBack` · `TestACriticalAlertStillEscalatesThroughTheSameNotifier` · `TestADeadTransportIsStillFoundAfterASuccessfulDelivery` (+14) | 대조(편집 없음) | 블록 223.19-229.3 을 시험 17 개가 실행, 전부 PASS |

# Branch Test Map: `alertDeliverer.deliverOne`

- Source: `internal/app/engine/alertdelivery.go` (270-340); current
- 시험 칸은 측정값: `analysis/harness/branch_coverage.py` 가 `github.com/JungHoonGhae/tossinvest-cli/internal/app/engine` 의 시험 498 개를 하나씩 돌린 커버 프로필(-covermode=set)에서 그 분기 본문 블록을 실행한 시험. 「합집합」은 패키지 전체 한 판.

| Branch | AST anchor | Scenario (source text) | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | if at 272:2 | `if err != nil {` | `TestACancelledEngineIsNotALedgerFault` · `TestClaimFailuresCountAsRecordFailures` · `TestJudgingTransactionsDelayTheExitCycleOnlyWithinTheFixedMargin` (+1) | a124 RED — 편집 전에 없던 분기 | 블록 272.16-277.3 을 시험 4 개가 실행, 전부 PASS |
| B2 | switch at 278:2 | `switch claim.Disposition {` | `TestACancelledDeliveryJudgementFallsBackToAnUnconditionalLatch` · `TestAClaimThatFindsTheRowSettledEndsItsRun` · `TestAClearAfterTheEvidenceLeavesWhatAnOnTimeLatchWouldHave` (+61) | a124 RED — 편집 전에 없던 분기 | 블록 278.2-278.27 을 시험 64 개가 실행, 전부 PASS |
| B3 | case at 279:2 | `case journal.ClaimAcquired:` | `TestACancelledDeliveryJudgementFallsBackToAnUnconditionalLatch` · `TestAClaimThatFindsTheRowSettledEndsItsRun` · `TestAClearAfterTheEvidenceLeavesWhatAnOnTimeLatchWouldHave` (+59) | a124 RED — 편집 전에 없던 분기 | 블록 279.29-282.25 을 시험 62 개가 실행, 전부 PASS |
| B4 | case at 283:2 | `case journal.ClaimHeldElsewhere:` | `TestAFailedAcknowledgementKeepsTheRecordFailureRun` · `TestAFailedSettlementKeepsTheLeaseSoTheNextCycleDoesNotResend` · `TestAHeldRowIsReportedOncePerLeaseAndNotOncePerCycle` (+3) | a124 RED — 편집 전에 없던 분기 | 블록 283.34-292.46 을 시험 6 개가 실행, 전부 PASS |
| B5 | if at 292:3 | `if d.reportHeld(alert.ID, claim.ExpiresAt) {` | `TestAFailedAcknowledgementKeepsTheRecordFailureRun` · `TestAFailedSettlementKeepsTheLeaseSoTheNextCycleDoesNotResend` · `TestAHeldRowIsReportedOncePerLeaseAndNotOncePerCycle` (+3) | a124 RED — 편집 전에 없던 분기 | 블록 292.46-296.4 을 시험 6 개가 실행, 전부 PASS |
| B6 | case at 298:2 | `default:` | `TestAClaimThatFindsTheRowSettledEndsItsRun` · `TestAnUnrelatedAcknowledgementDoesNotDropTheJudgement` | a124 RED — 편집 전에 없던 분기 | 블록 298.10-303.9 을 시험 2 개가 실행, 전부 PASS |
| B7 | if at 305:2 | `if claim.Stole {` | `TestAPublishedButUnrecordedAlertLatchesAtOnceAndKeepsTheLease` · `TestAnExpiredLeaseIsTakenOverWithoutTheConditionHappeningAgain` | a124 RED — 편집 전에 없던 분기 | 블록 305.17-310.3 을 시험 2 개가 실행, 전부 PASS |
| B8 | if at 314:2 | `if d.Publisher == nil {` | `TestJudgingTransactionsDelayTheExitCycleOnlyWithinTheFixedMargin` · `TestTheExecutorLatchesAndEscalatesAtTheAttemptLimit` · `TestTheProductionExecutorLatchesWithoutTheSynchronousPath` | a124 RED — 편집 전에 없던 분기 | 블록 314.24-324.3 을 시험 3 개가 실행, 전부 PASS |
| B9 | else at 324:9 | `} else {` | 없음 | a124 RED — 편집 전에 없던 분기 | 블록 None 을 어느 시험도 실행하지 않음 |
| B10 | if at 331:3 | `if perr != nil {` | `TestAClaimThatFindsTheRowSettledEndsItsRun` · `TestAClearAfterTheEvidenceLeavesWhatAnOnTimeLatchWouldHave` · `TestAClearAfterTheLimitthRecordFailureStillEscalates` (+43) | a124 RED — 편집 전에 없던 분기 | 블록 331.18-333.4 을 시험 46 개가 실행, 전부 PASS |
| B11 | if at 335:2 | `if perr != nil {` | `TestAClaimThatFindsTheRowSettledEndsItsRun` · `TestAClearAfterTheEvidenceLeavesWhatAnOnTimeLatchWouldHave` · `TestAClearAfterTheLimitthRecordFailureStillEscalates` (+43) | a124 RED — 편집 전에 없던 분기 | 블록 335.17-338.3 을 시험 46 개가 실행, 전부 PASS |

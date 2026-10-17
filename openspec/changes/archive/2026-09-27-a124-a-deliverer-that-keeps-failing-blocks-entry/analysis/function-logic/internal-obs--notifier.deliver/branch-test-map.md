# Branch Test Map: `Notifier.deliver`

- Source: `internal/obs/notifier.go` (420-574); current (a124 는 편집하지 않음 — 대조 번들)
- 시험 칸은 측정값: `analysis/harness/branch_coverage.py` 가 `github.com/JungHoonGhae/tossinvest-cli/internal/obs` 의 시험 81 개를 하나씩 돌린 커버 프로필에서 그 분기 본문 블록을 실행한 시험.

| Branch | AST anchor | Scenario (source text) | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | if at 422:2 | `if attempts <= 0 {` | 없음 | 대조(편집 없음) | 블록 422.19-424.3 을 어느 시험도 실행하지 않음 |
| B2 | for at 428:2 | `for attempt := 1; attempt <= attempts; attempt++ {` | `TestACancelledSenderStillHandsTheLeaseBack` · `TestACriticalAlertStillEscalatesThroughTheSameNotifier` · `TestADeadTransportIsStillFoundAfterASuccessfulDelivery` (+36) | 대조(편집 없음) | 블록 428.51-429.25 을 시험 39 개가 실행, 전부 PASS |
| B3 | if at 429:3 | `if n.Publisher == nil {` | 없음 | 대조(편집 없음) | 블록 429.25-431.9 을 어느 시험도 실행하지 않음 |
| B4 | if at 434:3 | `if err == nil {` | `TestADeadTransportIsStillFoundAfterASuccessfulDelivery` · `TestADeliveredAlertChangesNoMode` · `TestALaterOccurrenceOfTheSameKeyIsNotSwallowed` (+21) | 대조(편집 없음) | 블록 434.17-436.22 을 시험 24 개가 실행, 전부 PASS |
| B5 | if at 436:4 | `if markErr == nil {` | `TestADeadTransportIsStillFoundAfterASuccessfulDelivery` · `TestADeliveredAlertChangesNoMode` · `TestALaterOccurrenceOfTheSameKeyIsNotSwallowed` (+19) | 대조(편집 없음) | 블록 436.22-437.28 을 시험 22 개가 실행, 전부 PASS |
| B6 | switch at 437:5 | `switch settled.Outcome {` | `TestADeadTransportIsStillFoundAfterASuccessfulDelivery` · `TestADeliveredAlertChangesNoMode` · `TestALaterOccurrenceOfTheSameKeyIsNotSwallowed` (+19) | 대조(편집 없음) | 블록 436.22-437.28 을 시험 22 개가 실행, 전부 PASS |
| B7 | case at 438:5 | `case journal.SettleApplied:` | `TestADeadTransportIsStillFoundAfterASuccessfulDelivery` · `TestADeliveredAlertChangesNoMode` · `TestALaterOccurrenceOfTheSameKeyIsNotSwallowed` (+19) | 대조(편집 없음) | 블록 438.32-439.24 을 시험 22 개가 실행, 전부 PASS |
| B8 | case at 440:5 | `case journal.SettleLeaseLost, journal.SettleAlreadySettled:` | 없음 | 대조(편집 없음) | 블록 440.64-451.23 을 어느 시험도 실행하지 않음 |
| B9 | case at 452:5 | `case journal.SettleNotFound:` | 없음 | 대조(편집 없음) | 블록 452.33-453.85 을 어느 시험도 실행하지 않음 |
| B10 | case at 454:5 | `default:` | 없음 | 대조(편집 없음) | 블록 454.13-462.23 을 어느 시험도 실행하지 않음 |
| B11 | if at 478:4 | `if n.Log != nil {` | `TestAPublishedButUnsettledRowKeepsItsLease` · `TestASendThatCannotBeRecordedLatchesTheGate` | 대조(편집 없음) | 블록 478.20-482.5 을 시험 2 개가 실행, 전부 PASS |
| B12 | if at 483:4 | `if n.Gate != nil {` | `TestAPublishedButUnsettledRowKeepsItsLease` · `TestASendThatCannotBeRecordedLatchesTheGate` | 대조(편집 없음) | 블록 483.21-485.5 을 시험 2 개가 실행, 전부 PASS |
| B13 | if at 495:3 | `if markErr != nil {` | `TestACancelledSenderStillHandsTheLeaseBack` | 대조(편집 없음) | 블록 495.21-496.20 을 시험 1 개가 실행, 전부 PASS |
| B14 | else at 499:10 | `} else if failed.Outcome != journal.SettleApplied {` | `TestARowThatVanishedIsNotReportedAsContention` · `TestASenderThatLosesTheLeaseStopsAtOnce` | 대조(편집 없음) | 블록 499.53-509.20 을 시험 2 개가 실행, 전부 PASS |
| B15 | if at 496:4 | `if n.Log != nil {` | `TestACancelledSenderStillHandsTheLeaseBack` | 대조(편집 없음) | 블록 496.20-498.5 을 시험 1 개가 실행, 전부 PASS |
| B16 | if at 499:10 | `} else if failed.Outcome != journal.SettleApplied {` | `TestARowThatVanishedIsNotReportedAsContention` · `TestASenderThatLosesTheLeaseStopsAtOnce` | 대조(편집 없음) | 블록 499.53-509.20 을 시험 2 개가 실행, 전부 PASS |
| B17 | if at 509:4 | `if n.Log != nil {` | `TestARowThatVanishedIsNotReportedAsContention` · `TestASenderThatLosesTheLeaseStopsAtOnce` | 대조(편집 없음) | 블록 509.20-513.5 을 시험 2 개가 실행, 전부 PASS |
| B18 | if at 519:4 | `if failed.Outcome == journal.SettleNotFound && n.Gate != nil {` | `TestARowThatVanishedIsNotReportedAsContention` | 대조(편집 없음) | 블록 519.65-522.5 을 시험 1 개가 실행, 전부 PASS |
| B19 | if at 525:3 | `if attempt < attempts {` | `TestACancelledSenderStillHandsTheLeaseBack` · `TestACriticalAlertStillEscalatesThroughTheSameNotifier` · `TestADeadTransportIsStillFoundAfterASuccessfulDelivery` (+11) | 대조(편집 없음) | 블록 525.25-526.20 을 시험 14 개가 실행, 전부 PASS |
| B20 | if at 526:4 | `if !n.wait(ctx) {` | `TestACancelledSenderStillHandsTheLeaseBack` | 대조(편집 없음) | 블록 526.20-527.10 을 시험 1 개가 실행, 전부 PASS |
| B21 | switch at 543:2 | `switch {` | `TestACancelledSenderStillHandsTheLeaseBack` · `TestACriticalAlertStillEscalatesThroughTheSameNotifier` · `TestADeadTransportIsStillFoundAfterASuccessfulDelivery` (+12) | 대조(편집 없음) | 블록 540.2-543.9 을 시험 15 개가 실행, 전부 PASS |
| B22 | case at 544:2 | `case relErr != nil:` | 없음 | 대조(편집 없음) | 블록 544.21-545.19 을 어느 시험도 실행하지 않음 |
| B23 | if at 545:3 | `if n.Log != nil {` | 없음 | 대조(편집 없음) | 블록 545.19-547.4 을 어느 시험도 실행하지 않음 |
| B24 | case at 548:2 | `case released.Outcome == journal.SettleApplied:` | `TestACancelledSenderStillHandsTheLeaseBack` · `TestACriticalAlertStillEscalatesThroughTheSameNotifier` · `TestADeadTransportIsStillFoundAfterASuccessfulDelivery` (+12) | 대조(편집 없음) | 블록 548.49-548.49 을 시험 15 개가 실행, 전부 PASS |
| B25 | case at 551:2 | `default:` | 없음 | 대조(편집 없음) | 블록 551.10-561.21 을 어느 시험도 실행하지 않음 |
| B26 | if at 565:2 | `if n.Log != nil {` | `TestACancelledSenderStillHandsTheLeaseBack` · `TestACriticalAlertStillEscalatesThroughTheSameNotifier` · `TestADeadTransportIsStillFoundAfterASuccessfulDelivery` (+11) | 대조(편집 없음) | 블록 565.18-569.3 을 시험 14 개가 실행, 전부 PASS |
| B27 | if at 570:2 | `if n.Gate != nil {` | `TestACancelledSenderStillHandsTheLeaseBack` · `TestACriticalAlertStillEscalatesThroughTheSameNotifier` · `TestADeadTransportIsStillFoundAfterASuccessfulDelivery` (+11) | 대조(편집 없음) | 블록 570.19-572.3 을 시험 14 개가 실행, 전부 PASS |

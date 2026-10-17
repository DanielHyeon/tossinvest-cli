# Branch Test Map: `Notifier.deliver`

- Source: `internal/obs/notifier.go` (420-574); **편집 전** 측정 — `analysis/harness/coverage-pre-unit3.json`(연결 워크트리 `b3f14925`, `./internal/obs` 시험 92개를 하나씩).

| Branch | AST anchor | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | if at 422:2 | `attempts <= 0` | (미실행) | 편집 전(기준선) | 측정 표본의 시험 0개 |
| B2 | for at 428:2 | 시도 루프 | `TestA092BothAnnouncersBuildTheSameEvent`, `TestA092EachTransitionIsItsOwnAnnouncement` | 편집 전(기준선) | 블록 428.51-429.25을 시험 42개가 실행, PASS |
| B3 | if at 429:3 | `n.Publisher == nil` | (미실행) | 편집 전(기준선) | 측정 표본의 시험 0개 |
| B4 | if at 434:3 | 발행 성공(`err == nil`) | `TestA092BothAnnouncersBuildTheSameEvent`, `TestA092EachTransitionIsItsOwnAnnouncement` | 편집 전(기준선) | 블록 434.17-436.22을 시험 27개가 실행, PASS |
| B5 | if at 436:4 | `markErr == nil` | `TestA092BothAnnouncersBuildTheSameEvent`, `TestA092EachTransitionIsItsOwnAnnouncement` | 편집 전(기준선) | 블록 436.22-437.28을 시험 25개가 실행, PASS |
| B6 | switch at 437:5 | `settled.Outcome` | `TestA092BothAnnouncersBuildTheSameEvent`, `TestA092EachTransitionIsItsOwnAnnouncement` | 편집 전(기준선) | 블록 436.22-437.28을 시험 25개가 실행, PASS |
| B7 | case at 438:5 | `SettleApplied` | `TestA092BothAnnouncersBuildTheSameEvent`, `TestA092EachTransitionIsItsOwnAnnouncement` | 편집 전(기준선) | 블록 438.32-439.24을 시험 25개가 실행, PASS |
| B8 | case at 440:5 | `LeaseLost · AlreadySettled` | (미실행) | 편집 전(기준선) | 측정 표본의 시험 0개 |
| B9 | case at 452:5 | `SettleNotFound` | (미실행) | 편집 전(기준선) | 측정 표본의 시험 0개 |
| B10 | case at 454:5 | 모르는 결과 | (미실행) | 편집 전(기준선) | 측정 표본의 시험 0개 |
| B11 | if at 478:4 | `n.Log != nil` | `TestAPublishedButUnsettledRowKeepsItsLease`, `TestASendThatCannotBeRecordedLatchesTheGate` | 편집 전(기준선) | 블록 478.20-482.5을 시험 2개가 실행, PASS |
| B12 | if at 483:4 | `n.Gate != nil` | `TestAPublishedButUnsettledRowKeepsItsLease`, `TestASendThatCannotBeRecordedLatchesTheGate` | 편집 전(기준선) | 블록 483.21-485.5을 시험 2개가 실행, PASS |
| B13 | if at 495:3 | `MarkAlertAttemptFailed` 오류 | `TestACancelledSenderStillHandsTheLeaseBack` | 편집 전(기준선) | 블록 495.21-496.20을 시험 1개가 실행, PASS |
| B14 | else at 499:10 | 오류 없음 분기 | `TestARowThatVanishedIsNotReportedAsContention`, `TestASenderThatLosesTheLeaseStopsAtOnce` | 편집 전(기준선) | 블록 499.53-509.20을 시험 2개가 실행, PASS |
| B15 | if at 496:4 | `n.Log != nil` | `TestACancelledSenderStillHandsTheLeaseBack` | 편집 전(기준선) | 블록 496.20-498.5을 시험 1개가 실행, PASS |
| B16 | if at 499:10 | `failed.Outcome != SettleApplied` | `TestARowThatVanishedIsNotReportedAsContention`, `TestASenderThatLosesTheLeaseStopsAtOnce` | 편집 전(기준선) | 블록 499.53-509.20을 시험 2개가 실행, PASS |
| B17 | if at 509:4 | `n.Log != nil` | `TestARowThatVanishedIsNotReportedAsContention`, `TestASenderThatLosesTheLeaseStopsAtOnce` | 편집 전(기준선) | 블록 509.20-513.5을 시험 2개가 실행, PASS |
| B18 | if at 519:4 | `failed.Outcome == SettleNotFound && n.Gate != nil` | `TestARowThatVanishedIsNotReportedAsContention` | 편집 전(기준선) | 블록 519.65-522.5을 시험 1개가 실행, PASS |
| B19 | if at 525:3 | `attempt < attempts` | `TestACancelledSenderStillHandsTheLeaseBack`, `TestACriticalAlertStillEscalatesThroughTheSameNotifier` | 편집 전(기준선) | 블록 525.25-526.20을 시험 14개가 실행, PASS |
| B20 | if at 526:4 | `!n.wait(ctx)` | `TestACancelledSenderStillHandsTheLeaseBack` | 편집 전(기준선) | 블록 526.20-527.10을 시험 1개가 실행, PASS |
| B21 | switch at 543:2 | 반납 결과 | `TestACancelledSenderStillHandsTheLeaseBack`, `TestACriticalAlertStillEscalatesThroughTheSameNotifier` | 편집 전(기준선) | 블록 540.2-543.9을 시험 15개가 실행, PASS |
| B22 | case at 544:2 | `relErr != nil` | (미실행) | 편집 전(기준선) | 측정 표본의 시험 0개 |
| B23 | if at 545:3 | `n.Log != nil` | (미실행) | 편집 전(기준선) | 측정 표본의 시험 0개 |
| B24 | case at 548:2 | `SettleApplied` | `TestACancelledSenderStillHandsTheLeaseBack`, `TestACriticalAlertStillEscalatesThroughTheSameNotifier` | 편집 전(기준선) | 블록 548.49-548.49을 시험 15개가 실행, PASS |
| B25 | case at 551:2 | 그 밖(선점) | (미실행) | 편집 전(기준선) | 측정 표본의 시험 0개 |
| B26 | if at 565:2 | `n.Log != nil` | `TestACancelledSenderStillHandsTheLeaseBack`, `TestACriticalAlertStillEscalatesThroughTheSameNotifier` | 편집 전(기준선) | 블록 565.18-569.3을 시험 14개가 실행, PASS |
| B27 | if at 570:2 | `n.Gate != nil` | `TestACancelledSenderStillHandsTheLeaseBack`, `TestACriticalAlertStillEscalatesThroughTheSameNotifier` | 편집 전(기준선) | 블록 570.19-572.3을 시험 14개가 실행, PASS |

# Branch Test Map: `alertDeliverer.deliverOne`

- Source: `internal/app/engine/alertdelivery.go`; **편집 전** 측정 — `analysis/harness/coverage-pre-unit5-deliverone.json`(연결 워크트리 `b01e0cd0`, 시험 17개).

| Branch | AST anchor | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | if at 272:2 | `if err != nil` | (미실행) | 편집 전(기준선) | 측정 표본의 시험 0개 |
| B2 | switch at 278:2 | `switch claim.Disposition` | `TestACancelledDeliveryJudgementFallsBackToAnUnconditionalLatch`, `TestADeliveryAndAnEmptyAcknowledgementAfterTheEvidenceDropOnlyTheBlock` | 편집 전(기준선) | 블록 278.2-278.27을 시험 8개가 실행, PASS |
| B3 | case at 279:2 | `case journal.ClaimAcquired:` | `TestACancelledDeliveryJudgementFallsBackToAnUnconditionalLatch`, `TestADeliveryAndAnEmptyAcknowledgementAfterTheEvidenceDropOnlyTheBlock` | 편집 전(기준선) | 블록 279.29-282.25을 시험 8개가 실행, PASS |
| B4 | case at 283:2 | `case journal.ClaimHeldElsewhere:` | (미실행) | 편집 전(기준선) | 측정 표본의 시험 0개 |
| B5 | if at 292:3 | `if d.reportHeld(alert.ID, claim.ExpiresAt)` | (미실행) | 편집 전(기준선) | 측정 표본의 시험 0개 |
| B6 | case at 298:2 | `default:` | (미실행) | 편집 전(기준선) | 측정 표본의 시험 0개 |
| B7 | if at 305:2 | `if claim.Stole` | (미실행) | 편집 전(기준선) | 측정 표본의 시험 0개 |
| B8 | if at 314:2 | `if d.Publisher == nil` | (미실행) | 편집 전(기준선) | 측정 표본의 시험 0개 |
| B9 | else at 324:9 | `} else` | (미실행) | 편집 전(기준선) | 측정 표본의 시험 0개 |
| B10 | if at 331:3 | `if perr != nil` | `TestADeliveryAndAnEmptyAcknowledgementAfterTheEvidenceDropOnlyTheBlock` | 편집 전(기준선) | 블록 331.18-333.4을 시험 1개가 실행, PASS |
| B11 | if at 335:2 | `if perr != nil` | `TestADeliveryAndAnEmptyAcknowledgementAfterTheEvidenceDropOnlyTheBlock` | 편집 전(기준선) | 블록 335.17-338.3을 시험 1개가 실행, PASS |

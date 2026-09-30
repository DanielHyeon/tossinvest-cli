# Branch Test Map: `alertDeliverer.recordFailedAttempt`

- Source: `internal/app/engine/alertdelivery.go`; **편집 전** 측정 — `analysis/harness/coverage-pre-r26-deliverer.json`(연결 워크트리 `e55102f0`, 시험 19개).

| Branch | AST anchor | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | if at 350:2 | `if err != nil` | (미실행) | 편집 전(기준선) | 측정 표본의 시험 0개 |
| B2 | if at 355:2 | `if err != nil` | (미실행) | 편집 전(기준선) | 측정 표본의 시험 0개 |
| B3 | switch at 361:2 | `switch res.Outcome` | `TestADeliveryAndAnEmptyAcknowledgementAfterTheEvidenceDropOnlyTheBlock` | 편집 전(기준선) | 블록 361.2-361.21을 시험 1개가 실행, PASS |
| B4 | case at 362:2 | `case journal.SettleApplied:` | `TestADeliveryAndAnEmptyAcknowledgementAfterTheEvidenceDropOnlyTheBlock` | 편집 전(기준선) | 블록 362.29-365.39을 시험 1개가 실행, PASS |
| B5 | if at 365:3 | `if res.Attempts < alertAttemptLimit` | `TestADeliveryAndAnEmptyAcknowledgementAfterTheEvidenceDropOnlyTheBlock` | 편집 전(기준선) | 블록 365.39-367.4을 시험 1개가 실행, PASS |
| B6 | case at 369:2 | `case journal.SettleAlreadySettled, journal.SettleLeaseLost:` | (미실행) | 편집 전(기준선) | 측정 표본의 시험 0개 |
| B7 | case at 371:2 | `default:` | (미실행) | 편집 전(기준선) | 측정 표본의 시험 0개 |

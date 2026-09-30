# Branch Test Map: `Notifier.escalate`

- Source: `internal/obs/notifier.go`; **편집 전** 측정 — `analysis/harness/coverage-pre-unit3.json`(연결 워크트리 `b3f14925`, `./internal/obs` 시험 92개를 하나씩).

| Branch | AST anchor | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | if at 379:2 | 승격 미포함 | `TestACancelledSenderStillHandsTheLeaseBack`, `TestADeadTransportIsStillFoundAfterASuccessfulDelivery` | 편집 전(기준선) | 블록 379.63-381.3을 시험 14개가 실행, PASS |
| B2 | switch at 384:2 | 결과 분기 | `TestA092RecordOnlyFailureLatchesAndEscalates`, `TestAClaimThatFailsAttemptsTheDurableBlock` | 편집 전(기준선) | 블록 382.2-384.9을 시험 8개가 실행, PASS |
| B3 | case at 385:2 | 승격 실패 | `TestA092RecordOnlyFailureLatchesAndEscalates`, `TestAClaimThatFailsAttemptsTheDurableBlock` | 편집 전(기준선) | 블록 385.34-390.41을 시험 4개가 실행, PASS |
| B4 | case at 391:2 | 승격 됨 | `TestACriticalAlertStillEscalatesThroughTheSameNotifier`, `TestAnUndeliverableCriticalAlertTightensTheOperatingMode` | 편집 전(기준선) | 블록 391.31-397.92을 시험 4개가 실행, PASS |

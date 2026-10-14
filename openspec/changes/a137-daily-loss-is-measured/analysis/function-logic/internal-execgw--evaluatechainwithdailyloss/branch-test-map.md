# Branch Test Map: `evaluateChainWithDailyLoss`

- Source: `internal/execgw/a137_daily_loss.go` (183-189, 편집 뒤). 편집 전 분기수 1 → 편집 뒤 1. RED 는 `analysis/red/`, GREEN 은 `analysis/verify/receipts.md`.

| Branch | Anchor | Scenario | Test | a137 |
|---|---|---|---|---|
| B1 | if 185 | 기존 동작 | `TestA137AMeasurementFailureNeverFallsBackToTheCaller`(상세 교체) · `TestA137TheKillSwitchReasonSurvivesAMeasurementFailure`(앞 단 사유 보존) | 무편집 분기 |

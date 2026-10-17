# Branch Test Map: `evaluateChainWithDailyLoss`

- Source: `internal/execgw/a137_daily_loss.go` (187-199, 편집 뒤). 편집 전 분기수 1 → 편집 뒤 2. RED 는 `analysis/red/`, GREEN 은 `analysis/verify/receipts.md`.

| Branch | Anchor | Scenario | Test | a137 |
|---|---|---|---|---|
| B1 | if 189 | a137 편집 | `TestA137AMeasurementFailureNeverFallsBackToTheCaller` · `TestA137TheKillSwitchReasonSurvivesAMeasurementFailure` | 편집 분기 |
| B2 | if 194 | a137 편집 | `TestA137AnInvalidEquityWithAMeasurementFailureStaysUnmarked` | 편집 분기 |

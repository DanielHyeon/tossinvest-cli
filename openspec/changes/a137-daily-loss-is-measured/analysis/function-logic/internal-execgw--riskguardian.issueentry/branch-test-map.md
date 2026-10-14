# Branch Test Map: `RiskGuardian.IssueEntry`

- Source: `internal/execgw/riskguardian.go` (418-575, 편집 뒤). 편집 전 분기수 14 → 편집 뒤 14. RED 는 `analysis/red/`, GREEN 은 `analysis/verify/receipts.md`.

| Branch | Anchor | Scenario | Test | a137 |
|---|---|---|---|---|
| B1 | if 419 | 기존 동작 | `go test ./internal/execgw -run RiskGuardian` | 무편집 분기 |
| B2 | if 423 | 기존 동작 | `go test ./internal/execgw -run RiskGuardian` | 무편집 분기 |
| B3 | if 426 | 기존 동작 | `go test ./internal/execgw -run RiskGuardian` | 무편집 분기 |
| B4 | if 444 | a137 편집 | `TestA137TheLedgerLossRefusesAZeroCallerAndEscalates` · `TestA137ALargerValidCallerValueWins` · `TestA137AnInvalidCallerValueIsNotReplaced` · `TestA137AMeasurementFailureNeverFallsBackToTheCaller` · `TestA137TheKillSwitchReasonSurvivesAMeasurementFailure` · `TestReachingTheDailyLossLimitTightensTheOperatingMode`(기존) · `TestA137AnInvalidEquityWithAMeasurementFailureStaysUnmarked` | 편집 분기 |
| B5 | if 453 | 기존 동작 | `go test ./internal/execgw -run RiskGuardian` | 무편집 분기 |
| B6 | if 488 | 기존 동작 | `go test ./internal/execgw -run RiskGuardian` | 무편집 분기 |
| B7 | if 492 | 기존 동작 | `go test ./internal/execgw -run RiskGuardian` | 무편집 분기 |
| B8 | if 497 | 기존 동작 | `go test ./internal/execgw -run RiskGuardian` | 무편집 분기 |
| B9 | if 529 | 기존 동작 | `go test ./internal/execgw -run RiskGuardian` | 무편집 분기 |
| B10 | else 531 | 기존 동작 | `go test ./internal/execgw -run RiskGuardian` | 무편집 분기 |
| B11 | if 535 | 기존 동작 | `go test ./internal/execgw -run RiskGuardian` | 무편집 분기 |
| B12 | if 540 | 기존 동작 | `go test ./internal/execgw -run RiskGuardian` | 무편집 분기 |
| B13 | if 546 | 기존 동작 | `go test ./internal/execgw -run RiskGuardian` | 무편집 분기 |
| B14 | if 553 | 기존 동작 | `go test ./internal/execgw -run RiskGuardian` | 무편집 분기 |

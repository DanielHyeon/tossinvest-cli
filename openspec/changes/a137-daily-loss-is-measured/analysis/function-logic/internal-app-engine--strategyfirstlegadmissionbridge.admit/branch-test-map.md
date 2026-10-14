# Branch Test Map: `strategyFirstLegAdmissionBridge.admit`

- Source: `internal/app/engine/strategy_first_leg_admission.go` (80-128, 편집 뒤). 편집 전 분기수 8 → 편집 뒤 9. RED 는 `analysis/red/`, GREEN 은 `analysis/verify/receipts.md`.

| Branch | Anchor | Scenario | Test | a137 |
|---|---|---|---|---|
| B1 | if 82 | 기존 동작 | `go test -tags tossos_testseams ./internal/app/engine -run StrategyFirstLegAdmission` | 무편집 분기 |
| B2 | if 85 | 기존 동작 | `go test -tags tossos_testseams ./internal/app/engine -run StrategyFirstLegAdmission` | 무편집 분기 |
| B3 | if 89 | 기존 동작 | `go test -tags tossos_testseams ./internal/app/engine -run StrategyFirstLegAdmission` | 무편집 분기 |
| B4 | if 94 | 기존 동작 | `go test -tags tossos_testseams ./internal/app/engine -run StrategyFirstLegAdmission` | 무편집 분기 |
| B5 | if 100 | a137 편집 | `TestA137APrecheckRefusalKeepsItsCause` · `TestA137ClassificationIsUnchangedForEveryPrecheckRefusal` · 기존 `TestStrategyFirstLegAdmissionClassifiesPairedFailures` | 편집 분기 |
| B6 | if 103 | a137 편집 | `TestA137ADailyLossRefusalEscalatesAndIsNotALaneFault` · `TestA137AFailedEscalationIsJoinedNotLost` · `TestA137ClassificationIsUnchangedForEveryPrecheckRefusal` | 편집 분기 |
| B7 | if 116 | 기존 동작 | `go test -tags tossos_testseams ./internal/app/engine -run StrategyFirstLegAdmission` | 무편집 분기 |
| B8 | if 117 | 기존 동작 | `go test -tags tossos_testseams ./internal/app/engine -run StrategyFirstLegAdmission` | 무편집 분기 |
| B9 | if 124 | 기존 동작 | `go test -tags tossos_testseams ./internal/app/engine -run StrategyFirstLegAdmission` | 무편집 분기 |

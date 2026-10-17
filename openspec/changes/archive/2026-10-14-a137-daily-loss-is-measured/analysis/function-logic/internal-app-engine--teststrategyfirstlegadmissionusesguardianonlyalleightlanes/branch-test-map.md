# Branch Test Map: `TestStrategyFirstLegAdmissionUsesGuardianOnlyAllEightLanes`

- Source: `internal/app/engine/strategy_first_leg_admission_test.go` (32-58, 편집 뒤). 편집 전 분기수 4 → 편집 뒤 4. RED 는 `analysis/red/`, GREEN 은 `analysis/verify/receipts.md`.

| Branch | Anchor | Scenario | Test | a137 |
|---|---|---|---|---|
| B1 | range 34 | 기존 동작 | `go test -tags tossos_testseams ./internal/app/engine -run StrategyFirstLegAdmission` | 무편집 분기 |
| B2 | if 44 | 기존 동작 | `go test -tags tossos_testseams ./internal/app/engine -run StrategyFirstLegAdmission` | 무편집 분기 |
| B3 | if 47 | 기존 동작 | `go test -tags tossos_testseams ./internal/app/engine -run StrategyFirstLegAdmission` | 무편집 분기 |
| B4 | if 55 | 기존 동작 | `go test -tags tossos_testseams ./internal/app/engine -run StrategyFirstLegAdmission` | 무편집 분기 |

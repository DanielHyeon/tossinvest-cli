# Branch Test Map: `newProductionRiskGuardian`

- Source: `internal/app/engine/guardian_wiring.go` (80-116)

| Branch | Anchor | Scenario | Test |
|---|---|---|---|
| B1 | if 88 | 분기 동작 | `go test ./internal/app/engine -run 'ProductionPolicy|A138|A139'` · `TestA139AnUnderivableBudgetRefusesTheProductionGuardianByName` · `TestA139ProductionPolicyDerivesAThirdOfTheDailyLoss` |
| B2 | if 93 | 분기 동작 | `go test ./internal/app/engine -run 'ProductionPolicy|A138|A139'` · `TestA139AnUnderivableBudgetRefusesTheProductionGuardianByName` · `TestA139ProductionPolicyDerivesAThirdOfTheDailyLoss` |
| B3 | if 97 | 분기 동작 | `go test ./internal/app/engine -run 'ProductionPolicy|A138|A139'` · `TestA139AnUnderivableBudgetRefusesTheProductionGuardianByName` · `TestA139ProductionPolicyDerivesAThirdOfTheDailyLoss` |
| B4 | if 112 | 분기 동작 | `go test ./internal/app/engine -run 'ProductionPolicy|A138|A139'` · `TestA139AnUnderivableBudgetRefusesTheProductionGuardianByName` · `TestA139ProductionPolicyDerivesAThirdOfTheDailyLoss` |

# Branch Test Map: `riskPolicyFromAutomationGate`

- Source: `internal/app/engine/guardian_wiring.go` (37-78)

| Branch | Anchor | Scenario | Test |
|---|---|---|---|
| B1 | if 64 | 분기 동작 | `go test ./internal/app/engine -run 'RiskPolicy|A136|A138|A139'` · `TestA139ProductionPolicyDerivesAThirdOfTheDailyLoss` · `TestA139AnUnderivableBudgetRefusesTheProductionGuardianByName` |
| B2 | else 66 | 분기 동작 | `go test ./internal/app/engine -run 'RiskPolicy|A136|A138|A139'` · `TestA139ProductionPolicyDerivesAThirdOfTheDailyLoss` · `TestA139AnUnderivableBudgetRefusesTheProductionGuardianByName` |

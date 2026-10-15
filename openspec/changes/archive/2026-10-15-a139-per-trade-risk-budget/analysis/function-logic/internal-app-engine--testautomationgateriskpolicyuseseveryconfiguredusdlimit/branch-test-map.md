# Branch Test Map: `TestAutomationGateRiskPolicyUsesEveryConfiguredUSDLimit`

- Source: `internal/app/engine/guardian_wiring_internal_test.go` (15-48)

| Branch | Anchor | Scenario | Test |
|---|---|---|---|
| B1 | if 37 | 분기 동작 | `go test ./internal/app/engine -run RiskPolicy` |
| B2 | if 42 | 분기 동작 | `go test ./internal/app/engine -run RiskPolicy` |
| B3 | if 45 | 분기 동작 | `go test ./internal/app/engine -run RiskPolicy` |

# Branch Test Map: `riskPolicyFromAutomationGate`

- Source: `internal/app/engine/guardian_wiring.go` (34-69)

| Branch | Anchor | Scenario | Test |
|---|---|---|---|
| B1 | 직선(분기 없음) | 정상 경로 | `go test ./internal/app/engine -run 'RiskPolicy|A136|A138'` · `TestA138ProductionPolicyResolvesTheOpenPositionMaximum` |

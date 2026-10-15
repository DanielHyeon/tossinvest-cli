# Branch Test Map: `newProductionRiskGuardian`

- Source: `internal/app/engine/guardian_wiring.go` (71-103)

| Branch | Anchor | Scenario | Test |
|---|---|---|---|
| B1 | if 79 | 분기 동작 | `go test ./internal/app/engine -run 'ProductionPolicy|A136|A138'` · `TestA138ProductionPolicyResolvesTheOpenPositionMaximum` |
| B2 | if 84 | 분기 동작 | `go test ./internal/app/engine -run 'ProductionPolicy|A136|A138'` · `TestA138ProductionPolicyResolvesTheOpenPositionMaximum` |
| B3 | if 99 | 분기 동작 | `go test ./internal/app/engine -run 'ProductionPolicy|A136|A138'` · `TestA138ProductionPolicyResolvesTheOpenPositionMaximum` |

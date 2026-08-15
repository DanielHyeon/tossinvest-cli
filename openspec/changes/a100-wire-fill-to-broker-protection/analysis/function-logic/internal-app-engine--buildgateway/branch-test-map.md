# Branch Test Map: `buildGateway`

Source: `internal/app/engine/gateway.go` (234-355).

| Branch | Scenario | Test |
|---|---|---|
| B1 | projection wiring refusal | `TestGatewayClauseChecksTheWiringAndNotThePointer` |
| B2 | tracker restore refusal | `TestStartupConstructsTheGateway` |
| B3 | alert latch restore refusal | `TestStartupConstructsTheGateway` |
| B4 | readiness refusal | `TestTheProfileReportsItsProtectionReadiness` |
| B5 | execution gateway refusal | `TestStartupConstructsTheGateway` |

# Branch Test Map: `buildGateway`

Source: `internal/app/engine/gateway.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | if at `internal/app/engine/gateway.go:201` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B2 | if at `internal/app/engine/gateway.go:223` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B3 | if at `internal/app/engine/gateway.go:260` | Focused canonical/temporal ownership regressions plus full and race suites. |

Coverage includes external exclusion, identifier reuse, partial-scope rejection, v19 migration, confirmed temporal ownership, reservation recovery, and operator-release refusal/success.

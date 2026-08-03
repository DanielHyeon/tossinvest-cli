# Branch Test Map: `engineFillDetector`

Source: `cmd/tossctl/engine.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | if at `cmd/tossctl/engine.go:401` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B2 | if at `cmd/tossctl/engine.go:402` | Focused canonical/temporal ownership regressions plus full and race suites. |

Coverage includes external exclusion, identifier reuse, partial-scope rejection, v19 migration, confirmed temporal ownership, reservation recovery, and operator-release refusal/success.

# Branch Test Map: `TestStartupPrunesSpentNoncesOnce`

Source: `internal/app/engine/engine_test.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | if at `internal/app/engine/engine_test.go:451` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B2 | if at `internal/app/engine/engine_test.go:456` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B3 | if at `internal/app/engine/engine_test.go:460` | Focused canonical/temporal ownership regressions plus full and race suites. |

Coverage includes external exclusion, identifier reuse, partial-scope rejection, v19 migration, confirmed temporal ownership, reservation recovery, and operator-release refusal/success.

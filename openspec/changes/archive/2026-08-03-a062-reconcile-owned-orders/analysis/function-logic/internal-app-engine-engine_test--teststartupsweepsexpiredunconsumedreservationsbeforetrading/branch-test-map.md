# Branch Test Map: `TestStartupSweepsExpiredUnconsumedReservationsBeforeTrading`

Source: `internal/app/engine/engine_test.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | if at `internal/app/engine/engine_test.go:474` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B2 | if at `internal/app/engine/engine_test.go:478` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B3 | if at `internal/app/engine/engine_test.go:481` | Focused canonical/temporal ownership regressions plus full and race suites. |

Coverage includes external exclusion, identifier reuse, partial-scope rejection, v19 migration, confirmed temporal ownership, reservation recovery, and operator-release refusal/success.

# Branch Test Map: `TestReservationsAwaitingOperatorUsesOnlyExactScopedFailClosedSnapshot`

Source: `internal/journal/reservation_sweep_test.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | if at `internal/journal/reservation_sweep_test.go:304` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B2 | if at `internal/journal/reservation_sweep_test.go:307` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B3 | if at `internal/journal/reservation_sweep_test.go:317` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B4 | if at `internal/journal/reservation_sweep_test.go:329` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B5 | if at `internal/journal/reservation_sweep_test.go:332` | Focused canonical/temporal ownership regressions plus full and race suites. |

Coverage includes external exclusion, identifier reuse, partial-scope rejection, v19 migration, confirmed temporal ownership, reservation recovery, and operator-release refusal/success.

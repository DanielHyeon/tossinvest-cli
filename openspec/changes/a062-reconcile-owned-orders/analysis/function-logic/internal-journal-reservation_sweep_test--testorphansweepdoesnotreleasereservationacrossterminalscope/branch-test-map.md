# Branch Test Map: `TestOrphanSweepDoesNotReleaseReservationAcrossTerminalScope`

Source: `internal/journal/reservation_sweep_test.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | range at `internal/journal/reservation_sweep_test.go:353` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B2 | if at `internal/journal/reservation_sweep_test.go:362` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B3 | if at `internal/journal/reservation_sweep_test.go:365` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B4 | if at `internal/journal/reservation_sweep_test.go:368` | Focused canonical/temporal ownership regressions plus full and race suites. |

Coverage includes external exclusion, identifier reuse, partial-scope rejection, v19 migration, confirmed temporal ownership, reservation recovery, and operator-release refusal/success.

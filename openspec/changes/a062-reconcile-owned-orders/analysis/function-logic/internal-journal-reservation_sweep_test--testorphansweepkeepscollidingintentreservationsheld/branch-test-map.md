# Branch Test Map: `TestOrphanSweepKeepsCollidingIntentReservationsHeld`

Source: `internal/journal/reservation_sweep_test.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | if at `internal/journal/reservation_sweep_test.go:385` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B2 | if at `internal/journal/reservation_sweep_test.go:388` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B3 | range at `internal/journal/reservation_sweep_test.go:391` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B4 | if at `internal/journal/reservation_sweep_test.go:392` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B5 | if at `internal/journal/reservation_sweep_test.go:397` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B6 | if at `internal/journal/reservation_sweep_test.go:400` | Focused canonical/temporal ownership regressions plus full and race suites. |

Coverage includes external exclusion, identifier reuse, partial-scope rejection, v19 migration, confirmed temporal ownership, reservation recovery, and operator-release refusal/success.

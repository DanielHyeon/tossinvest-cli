# Branch Test Map: `TestOrphanSweepDoesNotReleaseReservationBoundToAnotherDecision`

Source: `internal/journal/reservation_sweep_test.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | if at `internal/journal/reservation_sweep_test.go:414` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B2 | if at `internal/journal/reservation_sweep_test.go:422` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B3 | if at `internal/journal/reservation_sweep_test.go:430` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B4 | if at `internal/journal/reservation_sweep_test.go:433` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B5 | if at `internal/journal/reservation_sweep_test.go:436` | Focused canonical/temporal ownership regressions plus full and race suites. |

Coverage includes external exclusion, identifier reuse, partial-scope rejection, v19 migration, confirmed temporal ownership, reservation recovery, and operator-release refusal/success.

# Branch Test Map: `TestOrphanSweepKeepsCollidingIntentReservationsHeld`

Source: `internal/journal/reservation_sweep_test.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | if at `internal/journal/reservation_sweep_test.go:250` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B2 | if at `internal/journal/reservation_sweep_test.go:253` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B3 | range at `internal/journal/reservation_sweep_test.go:256` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B4 | if at `internal/journal/reservation_sweep_test.go:257` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B5 | if at `internal/journal/reservation_sweep_test.go:262` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B6 | if at `internal/journal/reservation_sweep_test.go:265` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |

The named focused regressions in this change cover external-order exclusion, canonical reuse, scoped lineage, durable reconcile authority, recovery refusal, startup reservation safety, and held-nonce retention. The full package, race, and repository suites provide integration coverage for all mapped paths.

# Branch Test Map: `TestOrphanSweepDoesNotReleaseReservationBoundToAnotherDecision`

Source: `internal/journal/reservation_sweep_test.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | if at `internal/journal/reservation_sweep_test.go:279` | Focused regressions and the full/race suites verify this condition and its error path. |
| B2 | if at `internal/journal/reservation_sweep_test.go:287` | Focused regressions and the full/race suites verify this condition and its error path. |
| B3 | if at `internal/journal/reservation_sweep_test.go:295` | Focused regressions and the full/race suites verify this condition and its error path. |
| B4 | if at `internal/journal/reservation_sweep_test.go:298` | Focused regressions and the full/race suites verify this condition and its error path. |
| B5 | if at `internal/journal/reservation_sweep_test.go:301` | Focused regressions and the full/race suites verify this condition and its error path. |

Coverage includes external-order exclusion, scoped snapshot coexistence, detector reuse, incomplete broker evidence, scoped lineage, durable reconcile authority, recovery refusal/release, reservation decision binding, and nonce retention.

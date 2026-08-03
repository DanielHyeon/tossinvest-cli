# Branch Test Map: `TestCollidingTerminalFillKeepsEveryReservationHeld`

Source: `internal/journal/reservation_release_test.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | if at `internal/journal/reservation_release_test.go:197` | Focused regressions and the full/race suites verify this condition and its error path. |
| B2 | if at `internal/journal/reservation_release_test.go:205` | Focused regressions and the full/race suites verify this condition and its error path. |
| B3 | if at `internal/journal/reservation_release_test.go:216` | Focused regressions and the full/race suites verify this condition and its error path. |
| B4 | if at `internal/journal/reservation_release_test.go:219` | Focused regressions and the full/race suites verify this condition and its error path. |
| B5 | range at `internal/journal/reservation_release_test.go:222` | Focused regressions and the full/race suites verify this condition and its error path. |
| B6 | if at `internal/journal/reservation_release_test.go:223` | Focused regressions and the full/race suites verify this condition and its error path. |
| B7 | if at `internal/journal/reservation_release_test.go:228` | Focused regressions and the full/race suites verify this condition and its error path. |
| B8 | if at `internal/journal/reservation_release_test.go:231` | Focused regressions and the full/race suites verify this condition and its error path. |

Coverage includes external-order exclusion, scoped snapshot coexistence, detector reuse, incomplete broker evidence, scoped lineage, durable reconcile authority, recovery refusal/release, reservation decision binding, and nonce retention.

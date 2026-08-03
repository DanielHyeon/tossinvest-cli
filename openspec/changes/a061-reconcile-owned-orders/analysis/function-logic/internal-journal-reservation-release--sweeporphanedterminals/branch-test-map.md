# Branch Test Map: `sweepOrphanedTerminals`

Source: `internal/journal/reservation_release.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | if at `internal/journal/reservation_release.go:518` | Focused regressions and the full/race suites verify this condition and its error path. |
| B2 | if at `internal/journal/reservation_release.go:559` | Focused regressions and the full/race suites verify this condition and its error path. |
| B3 | for at `internal/journal/reservation_release.go:563` | Focused regressions and the full/race suites verify this condition and its error path. |
| B4 | if at `internal/journal/reservation_release.go:565` | Focused regressions and the full/race suites verify this condition and its error path. |
| B5 | if at `internal/journal/reservation_release.go:573` | Focused regressions and the full/race suites verify this condition and its error path. |
| B6 | range at `internal/journal/reservation_release.go:582` | Focused regressions and the full/race suites verify this condition and its error path. |
| B7 | if at `internal/journal/reservation_release.go:583` | Focused regressions and the full/race suites verify this condition and its error path. |
| B8 | if at `internal/journal/reservation_release.go:588` | Focused regressions and the full/race suites verify this condition and its error path. |
| B9 | if at `internal/journal/reservation_release.go:597` | Focused regressions and the full/race suites verify this condition and its error path. |

Coverage includes external-order exclusion, scoped snapshot coexistence, detector reuse, incomplete broker evidence, scoped lineage, durable reconcile authority, recovery refusal/release, reservation decision binding, and nonce retention.

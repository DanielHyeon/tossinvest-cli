# Branch Test Map: `Tracker.Refresh`

Source: `internal/reconcile/mismatch.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | if at `internal/reconcile/mismatch.go:623` | Focused regressions and the full/race suites verify this condition and its error path. |
| B2 | if at `internal/reconcile/mismatch.go:631` | Focused regressions and the full/race suites verify this condition and its error path. |
| B3 | range at `internal/reconcile/mismatch.go:637` | Focused regressions and the full/race suites verify this condition and its error path. |
| B4 | if at `internal/reconcile/mismatch.go:638` | Focused regressions and the full/race suites verify this condition and its error path. |
| B5 | range at `internal/reconcile/mismatch.go:642` | Focused regressions and the full/race suites verify this condition and its error path. |
| B6 | if at `internal/reconcile/mismatch.go:643` | Focused regressions and the full/race suites verify this condition and its error path. |
| B7 | if at `internal/reconcile/mismatch.go:651` | Focused regressions and the full/race suites verify this condition and its error path. |

Coverage includes external-order exclusion, scoped snapshot coexistence, detector reuse, incomplete broker evidence, scoped lineage, durable reconcile authority, recovery refusal/release, reservation decision binding, and nonce retention.

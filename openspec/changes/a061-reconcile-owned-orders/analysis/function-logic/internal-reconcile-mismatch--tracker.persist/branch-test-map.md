# Branch Test Map: `Tracker.persist`

Source: `internal/reconcile/mismatch.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | if at `internal/reconcile/mismatch.go:514` | Focused regressions and the full/race suites verify this condition and its error path. |
| B2 | range at `internal/reconcile/mismatch.go:524` | Focused regressions and the full/race suites verify this condition and its error path. |
| B3 | if at `internal/reconcile/mismatch.go:532` | Focused regressions and the full/race suites verify this condition and its error path. |
| B4 | if at `internal/reconcile/mismatch.go:536` | Focused regressions and the full/race suites verify this condition and its error path. |
| B5 | range at `internal/reconcile/mismatch.go:545` | Focused regressions and the full/race suites verify this condition and its error path. |
| B6 | if at `internal/reconcile/mismatch.go:553` | Focused regressions and the full/race suites verify this condition and its error path. |
| B7 | if at `internal/reconcile/mismatch.go:557` | Focused regressions and the full/race suites verify this condition and its error path. |

Coverage includes external-order exclusion, scoped snapshot coexistence, detector reuse, incomplete broker evidence, scoped lineage, durable reconcile authority, recovery refusal/release, reservation decision binding, and nonce retention.

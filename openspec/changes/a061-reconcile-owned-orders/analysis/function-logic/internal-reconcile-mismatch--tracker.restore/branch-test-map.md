# Branch Test Map: `Tracker.Restore`

Source: `internal/reconcile/mismatch.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | if at `internal/reconcile/mismatch.go:578` | Focused regressions and the full/race suites verify this condition and its error path. |
| B2 | if at `internal/reconcile/mismatch.go:582` | Focused regressions and the full/race suites verify this condition and its error path. |
| B3 | range at `internal/reconcile/mismatch.go:589` | Focused regressions and the full/race suites verify this condition and its error path. |
| B4 | if at `internal/reconcile/mismatch.go:590` | Focused regressions and the full/race suites verify this condition and its error path. |
| B5 | if at `internal/reconcile/mismatch.go:607` | Focused regressions and the full/race suites verify this condition and its error path. |
| B6 | if at `internal/reconcile/mismatch.go:610` | Focused regressions and the full/race suites verify this condition and its error path. |

Coverage includes external-order exclusion, scoped snapshot coexistence, detector reuse, incomplete broker evidence, scoped lineage, durable reconcile authority, recovery refusal/release, reservation decision binding, and nonce retention.

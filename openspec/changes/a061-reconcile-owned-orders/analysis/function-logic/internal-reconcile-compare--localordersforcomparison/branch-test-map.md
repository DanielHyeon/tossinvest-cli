# Branch Test Map: `localOrdersForComparison`

Source: `internal/reconcile/compare.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | if at `internal/reconcile/compare.go:480` | Focused regressions and the full/race suites verify this condition and its error path. |
| B2 | range at `internal/reconcile/compare.go:482` | Focused regressions and the full/race suites verify this condition and its error path. |
| B3 | range at `internal/reconcile/compare.go:497` | Focused regressions and the full/race suites verify this condition and its error path. |
| B4 | if at `internal/reconcile/compare.go:498` | Focused regressions and the full/race suites verify this condition and its error path. |

Coverage includes external-order exclusion, scoped snapshot coexistence, detector reuse, incomplete broker evidence, scoped lineage, durable reconcile authority, recovery refusal/release, reservation decision binding, and nonce retention.

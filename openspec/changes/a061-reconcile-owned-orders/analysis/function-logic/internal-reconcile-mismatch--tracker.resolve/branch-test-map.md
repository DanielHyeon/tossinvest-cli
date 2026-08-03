# Branch Test Map: `Tracker.Resolve`

Source: `internal/reconcile/mismatch.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | if at `internal/reconcile/mismatch.go:757` | Focused regressions and the full/race suites verify this condition and its error path. |
| B2 | if at `internal/reconcile/mismatch.go:760` | Focused regressions and the full/race suites verify this condition and its error path. |
| B3 | if at `internal/reconcile/mismatch.go:764` | Focused regressions and the full/race suites verify this condition and its error path. |
| B4 | range at `internal/reconcile/mismatch.go:767` | Focused regressions and the full/race suites verify this condition and its error path. |
| B5 | if at `internal/reconcile/mismatch.go:776` | Focused regressions and the full/race suites verify this condition and its error path. |

Coverage includes external-order exclusion, scoped snapshot coexistence, detector reuse, incomplete broker evidence, scoped lineage, durable reconcile authority, recovery refusal/release, reservation decision binding, and nonce retention.

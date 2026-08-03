# Branch Test Map: `TestObserveKeepsMemoryAndGateBlockedWhenDurableReleaseFails`

Source: `internal/reconcile/restore_test.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | range at `internal/reconcile/restore_test.go:425` | Focused regressions and the full/race suites verify this condition and its error path. |
| B2 | if at `internal/reconcile/restore_test.go:437` | Focused regressions and the full/race suites verify this condition and its error path. |
| B3 | if at `internal/reconcile/restore_test.go:445` | Focused regressions and the full/race suites verify this condition and its error path. |
| B4 | if at `internal/reconcile/restore_test.go:455` | Focused regressions and the full/race suites verify this condition and its error path. |
| B5 | if at `internal/reconcile/restore_test.go:458` | Focused regressions and the full/race suites verify this condition and its error path. |
| B6 | if at `internal/reconcile/restore_test.go:461` | Focused regressions and the full/race suites verify this condition and its error path. |
| B7 | if at `internal/reconcile/restore_test.go:464` | Focused regressions and the full/race suites verify this condition and its error path. |
| B8 | if at `internal/reconcile/restore_test.go:467` | Focused regressions and the full/race suites verify this condition and its error path. |

Coverage includes external-order exclusion, scoped snapshot coexistence, detector reuse, incomplete broker evidence, scoped lineage, durable reconcile authority, recovery refusal/release, reservation decision binding, and nonce retention.

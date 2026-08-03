# Branch Test Map: `releaseControllingStore.ReleaseReconciles`

Source: `internal/reconcile/restore_test.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | if at `internal/reconcile/restore_test.go:105` | Focused regressions and the full/race suites verify this condition and its error path. |
| B2 | if at `internal/reconcile/restore_test.go:108` | Focused regressions and the full/race suites verify this condition and its error path. |

Coverage includes external-order exclusion, scoped snapshot coexistence, detector reuse, incomplete broker evidence, scoped lineage, durable reconcile authority, recovery refusal/release, reservation decision binding, and nonce retention.

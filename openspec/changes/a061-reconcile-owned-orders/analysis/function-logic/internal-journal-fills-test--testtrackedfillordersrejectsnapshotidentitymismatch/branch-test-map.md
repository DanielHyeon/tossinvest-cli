# Branch Test Map: `TestTrackedFillOrdersRejectSnapshotIdentityMismatch`

Source: `internal/journal/fills_test.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | if at `internal/journal/fills_test.go:437` | Focused regressions and the full/race suites verify this condition and its error path. |
| B2 | if at `internal/journal/fills_test.go:442` | Focused regressions and the full/race suites verify this condition and its error path. |
| B3 | if at `internal/journal/fills_test.go:446` | Focused regressions and the full/race suites verify this condition and its error path. |
| B4 | if at `internal/journal/fills_test.go:449` | Focused regressions and the full/race suites verify this condition and its error path. |

Coverage includes external-order exclusion, scoped snapshot coexistence, detector reuse, incomplete broker evidence, scoped lineage, durable reconcile authority, recovery refusal/release, reservation decision binding, and nonce retention.

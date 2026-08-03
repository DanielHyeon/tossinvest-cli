# Branch Test Map: `TestTrackedFillOrders`

Source: `internal/journal/fills_test.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | if at `internal/journal/fills_test.go:385` | Focused regressions and the full/race suites verify this condition and its error path. |
| B2 | if at `internal/journal/fills_test.go:392` | Focused regressions and the full/race suites verify this condition and its error path. |
| B3 | if at `internal/journal/fills_test.go:397` | Focused regressions and the full/race suites verify this condition and its error path. |
| B4 | if at `internal/journal/fills_test.go:400` | Focused regressions and the full/race suites verify this condition and its error path. |
| B5 | if at `internal/journal/fills_test.go:403` | Focused regressions and the full/race suites verify this condition and its error path. |

Coverage includes external-order exclusion, scoped snapshot coexistence, detector reuse, incomplete broker evidence, scoped lineage, durable reconcile authority, recovery refusal/release, reservation decision binding, and nonce retention.

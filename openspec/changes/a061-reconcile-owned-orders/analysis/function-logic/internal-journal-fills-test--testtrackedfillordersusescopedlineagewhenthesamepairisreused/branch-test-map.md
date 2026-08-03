# Branch Test Map: `TestTrackedFillOrdersUseScopedLineageWhenTheSamePairIsReused`

Source: `internal/journal/fills_test.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | if at `internal/journal/fills_test.go:1146` | Focused regressions and the full/race suites verify this condition and its error path. |
| B2 | if at `internal/journal/fills_test.go:1151` | Focused regressions and the full/race suites verify this condition and its error path. |
| B3 | range at `internal/journal/fills_test.go:1155` | Focused regressions and the full/race suites verify this condition and its error path. |
| B4 | if at `internal/journal/fills_test.go:1156` | Focused regressions and the full/race suites verify this condition and its error path. |
| B5 | if at `internal/journal/fills_test.go:1158` | Focused regressions and the full/race suites verify this condition and its error path. |
| B6 | if at `internal/journal/fills_test.go:1163` | Focused regressions and the full/race suites verify this condition and its error path. |

Coverage includes external-order exclusion, scoped snapshot coexistence, detector reuse, incomplete broker evidence, scoped lineage, durable reconcile authority, recovery refusal/release, reservation decision binding, and nonce retention.

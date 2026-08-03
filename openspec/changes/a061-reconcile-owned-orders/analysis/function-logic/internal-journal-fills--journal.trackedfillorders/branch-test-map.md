# Branch Test Map: `Journal.TrackedFillOrders`

Source: `internal/journal/fills.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | if at `internal/journal/fills.go:1057` | Focused regressions and the full/race suites verify this condition and its error path. |
| B2 | if at `internal/journal/fills.go:1060` | Focused regressions and the full/race suites verify this condition and its error path. |
| B3 | if at `internal/journal/fills.go:1061` | Focused regressions and the full/race suites verify this condition and its error path. |
| B4 | if at `internal/journal/fills.go:1170` | Focused regressions and the full/race suites verify this condition and its error path. |
| B5 | for at `internal/journal/fills.go:1176` | Focused regressions and the full/race suites verify this condition and its error path. |
| B6 | if at `internal/journal/fills.go:1178` | Focused regressions and the full/race suites verify this condition and its error path. |
| B7 | if at `internal/journal/fills.go:1184` | Focused regressions and the full/race suites verify this condition and its error path. |
| B8 | range at `internal/journal/fills.go:1192` | Focused regressions and the full/race suites verify this condition and its error path. |
| B9 | if at `internal/journal/fills.go:1200` | Focused regressions and the full/race suites verify this condition and its error path. |
| B10 | if at `internal/journal/fills.go:1203` | Focused regressions and the full/race suites verify this condition and its error path. |

Coverage includes external-order exclusion, scoped snapshot coexistence, detector reuse, incomplete broker evidence, scoped lineage, durable reconcile authority, recovery refusal/release, reservation decision binding, and nonce retention.

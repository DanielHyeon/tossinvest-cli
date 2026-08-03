# Branch Test Map: `Journal.LiveOrdersForSymbol`

Source: `internal/journal/fills.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | if at `internal/journal/fills.go:1411` | Focused regressions and the full/race suites verify this condition and its error path. |
| B2 | for at `internal/journal/fills.go:1417` | Focused regressions and the full/race suites verify this condition and its error path. |
| B3 | if at `internal/journal/fills.go:1419` | Focused regressions and the full/race suites verify this condition and its error path. |
| B4 | if at `internal/journal/fills.go:1426` | Focused regressions and the full/race suites verify this condition and its error path. |
| B5 | range at `internal/journal/fills.go:1430` | Focused regressions and the full/race suites verify this condition and its error path. |
| B6 | if at `internal/journal/fills.go:1438` | Focused regressions and the full/race suites verify this condition and its error path. |

Coverage includes external-order exclusion, scoped snapshot coexistence, detector reuse, incomplete broker evidence, scoped lineage, durable reconcile authority, recovery refusal/release, reservation decision binding, and nonce retention.

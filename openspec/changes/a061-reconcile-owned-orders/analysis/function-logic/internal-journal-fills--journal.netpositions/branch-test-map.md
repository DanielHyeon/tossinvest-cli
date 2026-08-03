# Branch Test Map: `Journal.NetPositions`

Source: `internal/journal/fills.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | if at `internal/journal/fills.go:1016` | Focused regressions and the full/race suites verify this condition and its error path. |
| B2 | for at `internal/journal/fills.go:1022` | Focused regressions and the full/race suites verify this condition and its error path. |
| B3 | if at `internal/journal/fills.go:1024` | Focused regressions and the full/race suites verify this condition and its error path. |
| B4 | if at `internal/journal/fills.go:1028` | Focused regressions and the full/race suites verify this condition and its error path. |
| B5 | if at `internal/journal/fills.go:1031` | Focused regressions and the full/race suites verify this condition and its error path. |
| B6 | if at `internal/journal/fills.go:1036` | Focused regressions and the full/race suites verify this condition and its error path. |
| B7 | range at `internal/journal/fills.go:1041` | Focused regressions and the full/race suites verify this condition and its error path. |

Coverage includes external-order exclusion, scoped snapshot coexistence, detector reuse, incomplete broker evidence, scoped lineage, durable reconcile authority, recovery refusal/release, reservation decision binding, and nonce retention.

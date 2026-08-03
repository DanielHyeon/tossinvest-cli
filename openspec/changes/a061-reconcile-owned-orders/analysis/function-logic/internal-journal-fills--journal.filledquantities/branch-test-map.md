# Branch Test Map: `Journal.FilledQuantities`

Source: `internal/journal/fills.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | if at `internal/journal/fills.go:952` | Focused regressions and the full/race suites verify this condition and its error path. |
| B2 | for at `internal/journal/fills.go:958` | Focused regressions and the full/race suites verify this condition and its error path. |
| B3 | if at `internal/journal/fills.go:960` | Focused regressions and the full/race suites verify this condition and its error path. |
| B4 | if at `internal/journal/fills.go:964` | Focused regressions and the full/race suites verify this condition and its error path. |
| B5 | if at `internal/journal/fills.go:969` | Focused regressions and the full/race suites verify this condition and its error path. |
| B6 | range at `internal/journal/fills.go:973` | Focused regressions and the full/race suites verify this condition and its error path. |

Coverage includes external-order exclusion, scoped snapshot coexistence, detector reuse, incomplete broker evidence, scoped lineage, durable reconcile authority, recovery refusal/release, reservation decision binding, and nonce retention.

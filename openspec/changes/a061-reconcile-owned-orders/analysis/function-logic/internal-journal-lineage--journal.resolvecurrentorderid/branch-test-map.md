# Branch Test Map: `Journal.ResolveCurrentOrderID`

Source: `internal/journal/lineage.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | if at `internal/journal/lineage.go:200` | Focused regressions and the full/race suites verify this condition and its error path. |
| B2 | for at `internal/journal/lineage.go:204` | Focused regressions and the full/race suites verify this condition and its error path. |
| B3 | if at `internal/journal/lineage.go:206` | Focused regressions and the full/race suites verify this condition and its error path. |
| B4 | range at `internal/journal/lineage.go:210` | Focused regressions and the full/race suites verify this condition and its error path. |
| B5 | if at `internal/journal/lineage.go:211` | Focused regressions and the full/race suites verify this condition and its error path. |
| B6 | if at `internal/journal/lineage.go:216` | Focused regressions and the full/race suites verify this condition and its error path. |
| B7 | if at `internal/journal/lineage.go:219` | Focused regressions and the full/race suites verify this condition and its error path. |

Coverage includes external-order exclusion, scoped snapshot coexistence, detector reuse, incomplete broker evidence, scoped lineage, durable reconcile authority, recovery refusal/release, reservation decision binding, and nonce retention.

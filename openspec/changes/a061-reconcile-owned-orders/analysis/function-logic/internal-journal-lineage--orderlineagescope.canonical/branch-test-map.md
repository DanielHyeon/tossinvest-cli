# Branch Test Map: `OrderLineageScope.canonical`

Source: `internal/journal/lineage.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | switch at `internal/journal/lineage.go:72` | Focused regressions and the full/race suites verify this condition and its error path. |
| B2 | case at `internal/journal/lineage.go:73` | Focused regressions and the full/race suites verify this condition and its error path. |
| B3 | case at `internal/journal/lineage.go:75` | Focused regressions and the full/race suites verify this condition and its error path. |
| B4 | case at `internal/journal/lineage.go:77` | Focused regressions and the full/race suites verify this condition and its error path. |
| B5 | case at `internal/journal/lineage.go:79` | Focused regressions and the full/race suites verify this condition and its error path. |
| B6 | case at `internal/journal/lineage.go:81` | Focused regressions and the full/race suites verify this condition and its error path. |

Coverage includes external-order exclusion, scoped snapshot coexistence, detector reuse, incomplete broker evidence, scoped lineage, durable reconcile authority, recovery refusal/release, reservation decision binding, and nonce retention.

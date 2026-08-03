# Branch Test Map: `Journal.scopedLineageChildren`

Source: `internal/journal/lineage.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | if at `internal/journal/lineage.go:348` | Focused regressions and the full/race suites verify this condition and its error path. |
| B2 | if at `internal/journal/lineage.go:352` | Focused regressions and the full/race suites verify this condition and its error path. |
| B3 | range at `internal/journal/lineage.go:362` | Focused regressions and the full/race suites verify this condition and its error path. |
| B4 | if at `internal/journal/lineage.go:363` | Focused regressions and the full/race suites verify this condition and its error path. |
| B5 | if at `internal/journal/lineage.go:368` | Focused regressions and the full/race suites verify this condition and its error path. |

Coverage includes external-order exclusion, scoped snapshot coexistence, detector reuse, incomplete broker evidence, scoped lineage, durable reconcile authority, recovery refusal/release, reservation decision binding, and nonce retention.

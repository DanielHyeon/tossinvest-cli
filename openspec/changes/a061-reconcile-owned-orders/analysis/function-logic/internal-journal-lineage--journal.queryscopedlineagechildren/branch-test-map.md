# Branch Test Map: `Journal.queryScopedLineageChildren`

Source: `internal/journal/lineage.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | if at `internal/journal/lineage.go:423` | Focused regressions and the full/race suites verify this condition and its error path. |
| B2 | for at `internal/journal/lineage.go:429` | Focused regressions and the full/race suites verify this condition and its error path. |
| B3 | if at `internal/journal/lineage.go:431` | Focused regressions and the full/race suites verify this condition and its error path. |
| B4 | if at `internal/journal/lineage.go:436` | Focused regressions and the full/race suites verify this condition and its error path. |

Coverage includes external-order exclusion, scoped snapshot coexistence, detector reuse, incomplete broker evidence, scoped lineage, durable reconcile authority, recovery refusal/release, reservation decision binding, and nonce retention.

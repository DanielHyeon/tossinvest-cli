# Branch Test Map: `Journal.ResolveCurrentOrderIDScoped`

Source: `internal/journal/lineage.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | if at `internal/journal/lineage.go:313` | Focused regressions and the full/race suites verify this condition and its error path. |
| B2 | if at `internal/journal/lineage.go:317` | Focused regressions and the full/race suites verify this condition and its error path. |
| B3 | for at `internal/journal/lineage.go:323` | Focused regressions and the full/race suites verify this condition and its error path. |
| B4 | if at `internal/journal/lineage.go:325` | Focused regressions and the full/race suites verify this condition and its error path. |
| B5 | switch at `internal/journal/lineage.go:328` | Focused regressions and the full/race suites verify this condition and its error path. |
| B6 | case at `internal/journal/lineage.go:329` | Focused regressions and the full/race suites verify this condition and its error path. |
| B7 | case at `internal/journal/lineage.go:331` | Focused regressions and the full/race suites verify this condition and its error path. |
| B8 | case at `internal/journal/lineage.go:333` | Focused regressions and the full/race suites verify this condition and its error path. |
| B9 | if at `internal/journal/lineage.go:338` | Focused regressions and the full/race suites verify this condition and its error path. |

Coverage includes external-order exclusion, scoped snapshot coexistence, detector reuse, incomplete broker evidence, scoped lineage, durable reconcile authority, recovery refusal/release, reservation decision binding, and nonce retention.

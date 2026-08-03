# Branch Test Map: `TestOrderIDLatestDayPartitionKeepsBothMarketsInTheSameAccount`

Source: `internal/journal/fills_test.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | if at `internal/journal/fills_test.go:906` | Focused regressions and the full/race suites verify this condition and its error path. |
| B2 | if at `internal/journal/fills_test.go:909` | Focused regressions and the full/race suites verify this condition and its error path. |
| B3 | if at `internal/journal/fills_test.go:912` | Focused regressions and the full/race suites verify this condition and its error path. |
| B4 | if at `internal/journal/fills_test.go:915` | Focused regressions and the full/race suites verify this condition and its error path. |
| B5 | if at `internal/journal/fills_test.go:924` | Focused regressions and the full/race suites verify this condition and its error path. |
| B6 | range at `internal/journal/fills_test.go:928` | Focused regressions and the full/race suites verify this condition and its error path. |
| B7 | if at `internal/journal/fills_test.go:931` | Focused regressions and the full/race suites verify this condition and its error path. |
| B8 | if at `internal/journal/fills_test.go:935` | Focused regressions and the full/race suites verify this condition and its error path. |
| B9 | if at `internal/journal/fills_test.go:938` | Focused regressions and the full/race suites verify this condition and its error path. |

Coverage includes external-order exclusion, scoped snapshot coexistence, detector reuse, incomplete broker evidence, scoped lineage, durable reconcile authority, recovery refusal/release, reservation decision binding, and nonce retention.

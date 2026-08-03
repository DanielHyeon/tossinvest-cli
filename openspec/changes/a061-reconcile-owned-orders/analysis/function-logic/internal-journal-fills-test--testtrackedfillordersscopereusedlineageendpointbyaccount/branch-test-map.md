# Branch Test Map: `TestTrackedFillOrdersScopeReusedLineageEndpointByAccount`

Source: `internal/journal/fills_test.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | if at `internal/journal/fills_test.go:1057` | Focused regressions and the full/race suites verify this condition and its error path. |
| B2 | if at `internal/journal/fills_test.go:1060` | Focused regressions and the full/race suites verify this condition and its error path. |
| B3 | if at `internal/journal/fills_test.go:1076` | Focused regressions and the full/race suites verify this condition and its error path. |
| B4 | if at `internal/journal/fills_test.go:1079` | Focused regressions and the full/race suites verify this condition and its error path. |
| B5 | if at `internal/journal/fills_test.go:1082` | Focused regressions and the full/race suites verify this condition and its error path. |
| B6 | if at `internal/journal/fills_test.go:1085` | Focused regressions and the full/race suites verify this condition and its error path. |
| B7 | if at `internal/journal/fills_test.go:1090` | Focused regressions and the full/race suites verify this condition and its error path. |
| B8 | range at `internal/journal/fills_test.go:1094` | Focused regressions and the full/race suites verify this condition and its error path. |
| B9 | if at `internal/journal/fills_test.go:1098` | Focused regressions and the full/race suites verify this condition and its error path. |
| B10 | if at `internal/journal/fills_test.go:1101` | Focused regressions and the full/race suites verify this condition and its error path. |

Coverage includes external-order exclusion, scoped snapshot coexistence, detector reuse, incomplete broker evidence, scoped lineage, durable reconcile authority, recovery refusal/release, reservation decision binding, and nonce retention.

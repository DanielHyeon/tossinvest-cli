# Branch Test Map: `TestTrackedAndLiveOrdersAllowReusedIDInDifferentSymbolAndSideScopes`

Source: `internal/journal/fills_test.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | if at `internal/journal/fills_test.go:763` | Focused regressions and the full/race suites verify this condition and its error path. |
| B2 | if at `internal/journal/fills_test.go:766` | Focused regressions and the full/race suites verify this condition and its error path. |
| B3 | if at `internal/journal/fills_test.go:769` | Focused regressions and the full/race suites verify this condition and its error path. |
| B4 | if at `internal/journal/fills_test.go:772` | Focused regressions and the full/race suites verify this condition and its error path. |
| B5 | if at `internal/journal/fills_test.go:779` | Focused regressions and the full/race suites verify this condition and its error path. |
| B6 | if at `internal/journal/fills_test.go:784` | Focused regressions and the full/race suites verify this condition and its error path. |
| B7 | if at `internal/journal/fills_test.go:789` | Focused regressions and the full/race suites verify this condition and its error path. |
| B8 | if at `internal/journal/fills_test.go:792` | Focused regressions and the full/race suites verify this condition and its error path. |
| B9 | range at `internal/journal/fills_test.go:795` | Focused regressions and the full/race suites verify this condition and its error path. |
| B10 | if at `internal/journal/fills_test.go:797` | Focused regressions and the full/race suites verify this condition and its error path. |
| B11 | if at `internal/journal/fills_test.go:800` | Focused regressions and the full/race suites verify this condition and its error path. |

Coverage includes external-order exclusion, scoped snapshot coexistence, detector reuse, incomplete broker evidence, scoped lineage, durable reconcile authority, recovery refusal/release, reservation decision binding, and nonce retention.

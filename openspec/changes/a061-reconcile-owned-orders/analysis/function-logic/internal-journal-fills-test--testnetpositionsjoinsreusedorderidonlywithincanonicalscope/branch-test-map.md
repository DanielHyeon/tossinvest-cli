# Branch Test Map: `TestNetPositionsJoinsReusedOrderIDOnlyWithinCanonicalScope`

Source: `internal/journal/fills_test.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | if at `internal/journal/fills_test.go:717` | Focused regressions and the full/race suites verify this condition and its error path. |
| B2 | if at `internal/journal/fills_test.go:720` | Focused regressions and the full/race suites verify this condition and its error path. |
| B3 | if at `internal/journal/fills_test.go:723` | Focused regressions and the full/race suites verify this condition and its error path. |
| B4 | if at `internal/journal/fills_test.go:726` | Focused regressions and the full/race suites verify this condition and its error path. |
| B5 | if at `internal/journal/fills_test.go:733` | Focused regressions and the full/race suites verify this condition and its error path. |
| B6 | if at `internal/journal/fills_test.go:738` | Focused regressions and the full/race suites verify this condition and its error path. |
| B7 | if at `internal/journal/fills_test.go:743` | Focused regressions and the full/race suites verify this condition and its error path. |
| B8 | if at `internal/journal/fills_test.go:746` | Focused regressions and the full/race suites verify this condition and its error path. |

Coverage includes external-order exclusion, scoped snapshot coexistence, detector reuse, incomplete broker evidence, scoped lineage, durable reconcile authority, recovery refusal/release, reservation decision binding, and nonce retention.

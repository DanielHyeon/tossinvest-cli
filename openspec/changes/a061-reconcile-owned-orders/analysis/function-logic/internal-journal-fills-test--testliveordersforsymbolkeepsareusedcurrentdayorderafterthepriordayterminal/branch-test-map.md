# Branch Test Map: `TestLiveOrdersForSymbolKeepsAReusedCurrentDayOrderAfterThePriorDayTerminal`

Source: `internal/journal/fills_test.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | if at `internal/journal/fills_test.go:818` | Focused regressions and the full/race suites verify this condition and its error path. |
| B2 | if at `internal/journal/fills_test.go:821` | Focused regressions and the full/race suites verify this condition and its error path. |
| B3 | if at `internal/journal/fills_test.go:824` | Focused regressions and the full/race suites verify this condition and its error path. |
| B4 | if at `internal/journal/fills_test.go:827` | Focused regressions and the full/race suites verify this condition and its error path. |
| B5 | if at `internal/journal/fills_test.go:835` | Focused regressions and the full/race suites verify this condition and its error path. |
| B6 | if at `internal/journal/fills_test.go:841` | Focused regressions and the full/race suites verify this condition and its error path. |
| B7 | if at `internal/journal/fills_test.go:844` | Focused regressions and the full/race suites verify this condition and its error path. |

Coverage includes external-order exclusion, scoped snapshot coexistence, detector reuse, incomplete broker evidence, scoped lineage, durable reconcile authority, recovery refusal/release, reservation decision binding, and nonce retention.

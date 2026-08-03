# Branch Test Map: `TestScopedTerminalStateDoesNotHideAnotherAccountsLiveReusedOrder`

Source: `internal/journal/fills_test.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | if at `internal/journal/fills_test.go:607` | Focused regressions and the full/race suites verify this condition and its error path. |
| B2 | if at `internal/journal/fills_test.go:610` | Focused regressions and the full/race suites verify this condition and its error path. |
| B3 | if at `internal/journal/fills_test.go:613` | Focused regressions and the full/race suites verify this condition and its error path. |
| B4 | if at `internal/journal/fills_test.go:616` | Focused regressions and the full/race suites verify this condition and its error path. |
| B5 | if at `internal/journal/fills_test.go:622` | Focused regressions and the full/race suites verify this condition and its error path. |
| B6 | if at `internal/journal/fills_test.go:627` | Focused regressions and the full/race suites verify this condition and its error path. |
| B7 | if at `internal/journal/fills_test.go:632` | Focused regressions and the full/race suites verify this condition and its error path. |
| B8 | if at `internal/journal/fills_test.go:635` | Focused regressions and the full/race suites verify this condition and its error path. |
| B9 | if at `internal/journal/fills_test.go:639` | Focused regressions and the full/race suites verify this condition and its error path. |
| B10 | if at `internal/journal/fills_test.go:642` | Focused regressions and the full/race suites verify this condition and its error path. |
| B11 | if at `internal/journal/fills_test.go:646` | Focused regressions and the full/race suites verify this condition and its error path. |
| B12 | if at `internal/journal/fills_test.go:649` | Focused regressions and the full/race suites verify this condition and its error path. |

Coverage includes external-order exclusion, scoped snapshot coexistence, detector reuse, incomplete broker evidence, scoped lineage, durable reconcile authority, recovery refusal/release, reservation decision binding, and nonce retention.

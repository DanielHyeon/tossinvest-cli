# Branch Test Map: `exitContextForFill`

Source: `internal/journal/exit_state.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | if at `internal/journal/exit_state.go:875` | Focused regressions and the full/race suites verify this condition and its error path. |
| B2 | if at `internal/journal/exit_state.go:878` | Focused regressions and the full/race suites verify this condition and its error path. |
| B3 | if at `internal/journal/exit_state.go:890` | Focused regressions and the full/race suites verify this condition and its error path. |
| B4 | if at `internal/journal/exit_state.go:895` | Focused regressions and the full/race suites verify this condition and its error path. |
| B5 | if at `internal/journal/exit_state.go:898` | Focused regressions and the full/race suites verify this condition and its error path. |

Coverage includes external-order exclusion, scoped snapshot coexistence, detector reuse, incomplete broker evidence, scoped lineage, durable reconcile authority, recovery refusal/release, reservation decision binding, and nonce retention.

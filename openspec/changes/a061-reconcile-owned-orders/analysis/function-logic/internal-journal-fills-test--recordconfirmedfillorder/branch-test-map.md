# Branch Test Map: `recordConfirmedFillOrder`

Source: `internal/journal/fills_test.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | if at `internal/journal/fills_test.go:45` | Focused regressions and the full/race suites verify this condition and its error path. |
| B2 | if at `internal/journal/fills_test.go:48` | Focused regressions and the full/race suites verify this condition and its error path. |
| B3 | if at `internal/journal/fills_test.go:51` | Focused regressions and the full/race suites verify this condition and its error path. |
| B4 | if at `internal/journal/fills_test.go:54` | Focused regressions and the full/race suites verify this condition and its error path. |

Coverage includes external-order exclusion, scoped snapshot coexistence, detector reuse, incomplete broker evidence, scoped lineage, durable reconcile authority, recovery refusal/release, reservation decision binding, and nonce retention.

# Branch Test Map: `TestFailedEnterIsRetriedUntilTheBlockIsDurable`

Source: `internal/reconcile/restore_test.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | if at `internal/reconcile/restore_test.go:637` | Focused regressions and the full/race suites verify this condition and its error path. |
| B2 | if at `internal/reconcile/restore_test.go:640` | Focused regressions and the full/race suites verify this condition and its error path. |
| B3 | if at `internal/reconcile/restore_test.go:643` | Focused regressions and the full/race suites verify this condition and its error path. |
| B4 | if at `internal/reconcile/restore_test.go:648` | Focused regressions and the full/race suites verify this condition and its error path. |
| B5 | if at `internal/reconcile/restore_test.go:651` | Focused regressions and the full/race suites verify this condition and its error path. |
| B6 | if at `internal/reconcile/restore_test.go:655` | Focused regressions and the full/race suites verify this condition and its error path. |
| B7 | if at `internal/reconcile/restore_test.go:658` | Focused regressions and the full/race suites verify this condition and its error path. |
| B8 | if at `internal/reconcile/restore_test.go:661` | Focused regressions and the full/race suites verify this condition and its error path. |

Coverage includes external-order exclusion, scoped snapshot coexistence, detector reuse, incomplete broker evidence, scoped lineage, durable reconcile authority, recovery refusal/release, reservation decision binding, and nonce retention.

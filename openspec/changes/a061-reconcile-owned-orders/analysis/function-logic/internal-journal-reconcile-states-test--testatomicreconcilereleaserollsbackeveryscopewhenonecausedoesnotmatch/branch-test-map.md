# Branch Test Map: `TestAtomicReconcileReleaseRollsBackEveryScopeWhenOneCauseDoesNotMatch`

Source: `internal/journal/reconcile_states_test.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | range at `internal/journal/reconcile_states_test.go:296` | Focused regressions and the full/race suites verify this condition and its error path. |
| B2 | if at `internal/journal/reconcile_states_test.go:300` | Focused regressions and the full/race suites verify this condition and its error path. |
| B3 | if at `internal/journal/reconcile_states_test.go:310` | Focused regressions and the full/race suites verify this condition and its error path. |
| B4 | if at `internal/journal/reconcile_states_test.go:314` | Focused regressions and the full/race suites verify this condition and its error path. |
| B5 | if at `internal/journal/reconcile_states_test.go:317` | Focused regressions and the full/race suites verify this condition and its error path. |
| B6 | range at `internal/journal/reconcile_states_test.go:320` | Focused regressions and the full/race suites verify this condition and its error path. |
| B7 | if at `internal/journal/reconcile_states_test.go:321` | Focused regressions and the full/race suites verify this condition and its error path. |

Coverage includes external-order exclusion, scoped snapshot coexistence, detector reuse, incomplete broker evidence, scoped lineage, durable reconcile authority, recovery refusal/release, reservation decision binding, and nonce retention.

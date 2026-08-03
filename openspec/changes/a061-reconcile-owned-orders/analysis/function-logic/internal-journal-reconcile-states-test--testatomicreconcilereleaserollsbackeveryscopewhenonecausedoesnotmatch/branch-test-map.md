# Branch Test Map: `TestAtomicReconcileReleaseRollsBackEveryScopeWhenOneCauseDoesNotMatch`

Source: `internal/journal/reconcile_states_test.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | range at `internal/journal/reconcile_states_test.go:296` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B2 | if at `internal/journal/reconcile_states_test.go:300` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B3 | if at `internal/journal/reconcile_states_test.go:310` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B4 | if at `internal/journal/reconcile_states_test.go:314` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B5 | if at `internal/journal/reconcile_states_test.go:317` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B6 | range at `internal/journal/reconcile_states_test.go:320` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B7 | if at `internal/journal/reconcile_states_test.go:321` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |

The named focused regressions in this change cover external-order exclusion, canonical reuse, scoped lineage, durable reconcile authority, recovery refusal, startup reservation safety, and held-nonce retention. The full package, race, and repository suites provide integration coverage for all mapped paths.

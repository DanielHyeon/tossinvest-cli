# Branch Test Map: `TestAtomicReconcileReleaseRollsBackEveryScopeWhenOneCauseDoesNotMatch`

Source: `internal/journal/reconcile_states_test.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | range at `internal/journal/reconcile_states_test.go:296` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B2 | if at `internal/journal/reconcile_states_test.go:300` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B3 | if at `internal/journal/reconcile_states_test.go:310` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B4 | if at `internal/journal/reconcile_states_test.go:314` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B5 | if at `internal/journal/reconcile_states_test.go:317` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B6 | range at `internal/journal/reconcile_states_test.go:320` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B7 | if at `internal/journal/reconcile_states_test.go:321` | Focused canonical/temporal ownership regressions plus full and race suites. |

Coverage includes external exclusion, identifier reuse, partial-scope rejection, v19 migration, confirmed temporal ownership, reservation recovery, and operator-release refusal/success.

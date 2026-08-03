# Branch Test Map: `TestObserveKeepsMemoryAndGateBlockedWhenDurableReleaseFails`

Source: `internal/reconcile/restore_test.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | range at `internal/reconcile/restore_test.go:425` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B2 | if at `internal/reconcile/restore_test.go:437` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B3 | if at `internal/reconcile/restore_test.go:445` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B4 | if at `internal/reconcile/restore_test.go:455` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B5 | if at `internal/reconcile/restore_test.go:458` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B6 | if at `internal/reconcile/restore_test.go:461` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B7 | if at `internal/reconcile/restore_test.go:464` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B8 | if at `internal/reconcile/restore_test.go:467` | Focused canonical/temporal ownership regressions plus full and race suites. |

Coverage includes external exclusion, identifier reuse, partial-scope rejection, v19 migration, confirmed temporal ownership, reservation recovery, and operator-release refusal/success.

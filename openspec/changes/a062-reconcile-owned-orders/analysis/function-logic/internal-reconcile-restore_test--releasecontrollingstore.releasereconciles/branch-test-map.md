# Branch Test Map: `releaseControllingStore.ReleaseReconciles`

Source: `internal/reconcile/restore_test.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | if at `internal/reconcile/restore_test.go:105` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B2 | if at `internal/reconcile/restore_test.go:108` | Focused canonical/temporal ownership regressions plus full and race suites. |

Coverage includes external exclusion, identifier reuse, partial-scope rejection, v19 migration, confirmed temporal ownership, reservation recovery, and operator-release refusal/success.

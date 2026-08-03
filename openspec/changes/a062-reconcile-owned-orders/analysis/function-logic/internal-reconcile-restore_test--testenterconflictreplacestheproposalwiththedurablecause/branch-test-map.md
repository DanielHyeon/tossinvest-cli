# Branch Test Map: `TestEnterConflictReplacesTheProposalWithTheDurableCause`

Source: `internal/reconcile/restore_test.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | if at `internal/reconcile/restore_test.go:611` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B2 | if at `internal/reconcile/restore_test.go:614` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B3 | if at `internal/reconcile/restore_test.go:618` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B4 | if at `internal/reconcile/restore_test.go:621` | Focused canonical/temporal ownership regressions plus full and race suites. |

Coverage includes external exclusion, identifier reuse, partial-scope rejection, v19 migration, confirmed temporal ownership, reservation recovery, and operator-release refusal/success.

# Branch Test Map: `TestRefreshCannotOverwriteABlockPersistedByObserve`

Source: `internal/reconcile/restore_test.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | select at `internal/reconcile/restore_test.go:291` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B2 | if at `internal/reconcile/restore_test.go:299` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B3 | if at `internal/reconcile/restore_test.go:302` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B4 | if at `internal/reconcile/restore_test.go:305` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B5 | if at `internal/reconcile/restore_test.go:308` | Focused canonical/temporal ownership regressions plus full and race suites. |

Coverage includes external exclusion, identifier reuse, partial-scope rejection, v19 migration, confirmed temporal ownership, reservation recovery, and operator-release refusal/success.

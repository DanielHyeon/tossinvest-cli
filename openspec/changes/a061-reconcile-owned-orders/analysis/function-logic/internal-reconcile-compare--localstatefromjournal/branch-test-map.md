# Branch Test Map: `LocalStateFromJournal`

Source: `internal/reconcile/compare.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | if at `internal/reconcile/compare.go:199` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B2 | if at `internal/reconcile/compare.go:203` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B3 | if at `internal/reconcile/compare.go:207` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B4 | range at `internal/reconcile/compare.go:217` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B5 | if at `internal/reconcile/compare.go:225` | Focused canonical/temporal ownership regressions plus full and race suites. |

Coverage includes external exclusion, identifier reuse, partial-scope rejection, v19 migration, confirmed temporal ownership, reservation recovery, and operator-release refusal/success.

# Branch Test Map: `OrderIdentity.less`

Source: `internal/reconcile/compare.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | range at `internal/reconcile/compare.go:82` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B2 | if at `internal/reconcile/compare.go:83` | Focused canonical/temporal ownership regressions plus full and race suites. |

Coverage includes external exclusion, identifier reuse, partial-scope rejection, v19 migration, confirmed temporal ownership, reservation recovery, and operator-release refusal/success.

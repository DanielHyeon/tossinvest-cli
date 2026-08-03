# Branch Test Map: `Block.Key`

Source: `internal/reconcile/mismatch.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | switch at `internal/reconcile/mismatch.go:223` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B2 | case at `internal/reconcile/mismatch.go:224` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B3 | case at `internal/reconcile/mismatch.go:226` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B4 | case at `internal/reconcile/mismatch.go:228` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B5 | case at `internal/reconcile/mismatch.go:230` | Focused canonical/temporal ownership regressions plus full and race suites. |

Coverage includes external exclusion, identifier reuse, partial-scope rejection, v19 migration, confirmed temporal ownership, reservation recovery, and operator-release refusal/success.

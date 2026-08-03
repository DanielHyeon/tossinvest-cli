# Branch Test Map: `Tracker.Restore`

Source: `internal/reconcile/mismatch.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | if at `internal/reconcile/mismatch.go:578` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B2 | if at `internal/reconcile/mismatch.go:582` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B3 | range at `internal/reconcile/mismatch.go:589` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B4 | if at `internal/reconcile/mismatch.go:590` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B5 | if at `internal/reconcile/mismatch.go:607` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B6 | if at `internal/reconcile/mismatch.go:610` | Focused canonical/temporal ownership regressions plus full and race suites. |

Coverage includes external exclusion, identifier reuse, partial-scope rejection, v19 migration, confirmed temporal ownership, reservation recovery, and operator-release refusal/success.

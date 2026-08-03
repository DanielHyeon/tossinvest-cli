# Branch Test Map: `Tracker.Refresh`

Source: `internal/reconcile/mismatch.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | if at `internal/reconcile/mismatch.go:623` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B2 | if at `internal/reconcile/mismatch.go:631` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B3 | range at `internal/reconcile/mismatch.go:637` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B4 | if at `internal/reconcile/mismatch.go:638` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B5 | range at `internal/reconcile/mismatch.go:642` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B6 | if at `internal/reconcile/mismatch.go:643` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B7 | if at `internal/reconcile/mismatch.go:651` | Focused canonical/temporal ownership regressions plus full and race suites. |

Coverage includes external exclusion, identifier reuse, partial-scope rejection, v19 migration, confirmed temporal ownership, reservation recovery, and operator-release refusal/success.

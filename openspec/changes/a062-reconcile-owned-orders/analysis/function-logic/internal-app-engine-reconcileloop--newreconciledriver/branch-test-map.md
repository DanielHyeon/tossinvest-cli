# Branch Test Map: `NewReconcileDriver`

Source: `internal/app/engine/reconcileloop.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | switch at `internal/app/engine/reconcileloop.go:280` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B2 | case at `internal/app/engine/reconcileloop.go:281` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B3 | case at `internal/app/engine/reconcileloop.go:283` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B4 | case at `internal/app/engine/reconcileloop.go:285` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B5 | case at `internal/app/engine/reconcileloop.go:287` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B6 | case at `internal/app/engine/reconcileloop.go:289` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B7 | case at `internal/app/engine/reconcileloop.go:292` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B8 | if at `internal/app/engine/reconcileloop.go:296` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B9 | if at `internal/app/engine/reconcileloop.go:297` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B10 | if at `internal/app/engine/reconcileloop.go:309` | Focused canonical/temporal ownership regressions plus full and race suites. |

Coverage includes external exclusion, identifier reuse, partial-scope rejection, v19 migration, confirmed temporal ownership, reservation recovery, and operator-release refusal/success.

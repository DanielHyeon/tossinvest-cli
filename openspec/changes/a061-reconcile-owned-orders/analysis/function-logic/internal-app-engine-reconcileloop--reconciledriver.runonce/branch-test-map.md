# Branch Test Map: `ReconcileDriver.RunOnce`

Source: `internal/app/engine/reconcileloop.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | if at `internal/app/engine/reconcileloop.go:387` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B2 | if at `internal/app/engine/reconcileloop.go:393` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B3 | if at `internal/app/engine/reconcileloop.go:404` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B4 | if at `internal/app/engine/reconcileloop.go:410` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B5 | if at `internal/app/engine/reconcileloop.go:416` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B6 | if at `internal/app/engine/reconcileloop.go:417` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B7 | if at `internal/app/engine/reconcileloop.go:425` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B8 | if at `internal/app/engine/reconcileloop.go:426` | Focused canonical/temporal ownership regressions plus full and race suites. |

Coverage includes external exclusion, identifier reuse, partial-scope rejection, v19 migration, confirmed temporal ownership, reservation recovery, and operator-release refusal/success.

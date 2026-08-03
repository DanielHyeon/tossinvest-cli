# Branch Test Map: `Journal.ReleaseReconcile`

Source: `internal/journal/reconcile_states.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | switch at `internal/journal/reconcile_states.go:251` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B2 | case at `internal/journal/reconcile_states.go:252` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B3 | case at `internal/journal/reconcile_states.go:255` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B4 | case at `internal/journal/reconcile_states.go:258` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B5 | if at `internal/journal/reconcile_states.go:267` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B6 | if at `internal/journal/reconcile_states.go:274` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B7 | if at `internal/journal/reconcile_states.go:277` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B8 | if at `internal/journal/reconcile_states.go:280` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B9 | if at `internal/journal/reconcile_states.go:284` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B10 | if at `internal/journal/reconcile_states.go:290` | Focused canonical/temporal ownership regressions plus full and race suites. |

Coverage includes external exclusion, identifier reuse, partial-scope rejection, v19 migration, confirmed temporal ownership, reservation recovery, and operator-release refusal/success.

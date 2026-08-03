# Branch Test Map: `TestAdoptionIsSilentUnderReconcile`

Source: `internal/app/engine/reconcileloop_test.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | if at `internal/app/engine/reconcileloop_test.go:359` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B2 | if at `internal/app/engine/reconcileloop_test.go:366` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B3 | if at `internal/app/engine/reconcileloop_test.go:370` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B4 | if at `internal/app/engine/reconcileloop_test.go:374` | Focused canonical/temporal ownership regressions plus full and race suites. |

Coverage includes external exclusion, identifier reuse, partial-scope rejection, v19 migration, confirmed temporal ownership, reservation recovery, and operator-release refusal/success.

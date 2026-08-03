# Branch Test Map: `TestTrackerReleaseFailureStopsBeforePriceAndAdoption`

Source: `internal/app/engine/reconcileloop_test.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | if at `internal/app/engine/reconcileloop_test.go:403` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B2 | if at `internal/app/engine/reconcileloop_test.go:409` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B3 | if at `internal/app/engine/reconcileloop_test.go:416` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B4 | if at `internal/app/engine/reconcileloop_test.go:419` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B5 | if at `internal/app/engine/reconcileloop_test.go:422` | Focused canonical/temporal ownership regressions plus full and race suites. |

Coverage includes external exclusion, identifier reuse, partial-scope rejection, v19 migration, confirmed temporal ownership, reservation recovery, and operator-release refusal/success.

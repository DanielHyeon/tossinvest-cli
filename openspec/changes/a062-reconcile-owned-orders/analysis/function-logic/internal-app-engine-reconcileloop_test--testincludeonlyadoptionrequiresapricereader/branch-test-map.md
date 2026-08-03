# Branch Test Map: `TestIncludeOnlyAdoptionRequiresAPriceReader`

Source: `internal/app/engine/reconcileloop_test.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | if at `internal/app/engine/reconcileloop_test.go:435` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B2 | if at `internal/app/engine/reconcileloop_test.go:446` | Focused canonical/temporal ownership regressions plus full and race suites. |

Coverage includes external exclusion, identifier reuse, partial-scope rejection, v19 migration, confirmed temporal ownership, reservation recovery, and operator-release refusal/success.

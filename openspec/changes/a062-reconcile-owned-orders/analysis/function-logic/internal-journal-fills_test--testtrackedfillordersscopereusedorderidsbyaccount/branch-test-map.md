# Branch Test Map: `TestTrackedFillOrdersScopeReusedOrderIDsByAccount`

Source: `internal/journal/fills_test.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | if at `internal/journal/fills_test.go:544` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B2 | if at `internal/journal/fills_test.go:547` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B3 | if at `internal/journal/fills_test.go:550` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B4 | if at `internal/journal/fills_test.go:553` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B5 | range at `internal/journal/fills_test.go:557` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B6 | if at `internal/journal/fills_test.go:559` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B7 | if at `internal/journal/fills_test.go:562` | Focused canonical/temporal ownership regressions plus full and race suites. |

Coverage includes external exclusion, identifier reuse, partial-scope rejection, v19 migration, confirmed temporal ownership, reservation recovery, and operator-release refusal/success.

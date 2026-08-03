# Branch Test Map: `TestTrackedFillOrdersExcludeExternalSnapshots`

Source: `internal/journal/fills_test.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | if at `internal/journal/fills_test.go:493` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B2 | if at `internal/journal/fills_test.go:496` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B3 | if at `internal/journal/fills_test.go:501` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B4 | if at `internal/journal/fills_test.go:504` | Focused canonical/temporal ownership regressions plus full and race suites. |

Coverage includes external exclusion, identifier reuse, partial-scope rejection, v19 migration, confirmed temporal ownership, reservation recovery, and operator-release refusal/success.

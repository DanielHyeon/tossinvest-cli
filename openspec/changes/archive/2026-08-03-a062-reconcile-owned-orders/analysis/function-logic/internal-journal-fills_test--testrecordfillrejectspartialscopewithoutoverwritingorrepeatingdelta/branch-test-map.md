# Branch Test Map: `TestRecordFillRejectsPartialScopeWithoutOverwritingOrRepeatingDelta`

Source: `internal/journal/fills_test.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | if at `internal/journal/fills_test.go:698` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B2 | for at `internal/journal/fills_test.go:704` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B3 | if at `internal/journal/fills_test.go:705` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B4 | if at `internal/journal/fills_test.go:710` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B5 | if at `internal/journal/fills_test.go:713` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B6 | if at `internal/journal/fills_test.go:717` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B7 | if at `internal/journal/fills_test.go:721` | Focused canonical/temporal ownership regressions plus full and race suites. |

Coverage includes external exclusion, identifier reuse, partial-scope rejection, v19 migration, confirmed temporal ownership, reservation recovery, and operator-release refusal/success.

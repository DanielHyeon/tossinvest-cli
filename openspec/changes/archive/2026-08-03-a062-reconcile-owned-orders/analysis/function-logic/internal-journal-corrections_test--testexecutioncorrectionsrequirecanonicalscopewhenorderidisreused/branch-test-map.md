# Branch Test Map: `TestExecutionCorrectionsRequireCanonicalScopeWhenOrderIDIsReused`

Source: `internal/journal/corrections_test.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | if at `internal/journal/corrections_test.go:134` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B2 | if at `internal/journal/corrections_test.go:138` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B3 | if at `internal/journal/corrections_test.go:144` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B4 | if at `internal/journal/corrections_test.go:148` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B5 | if at `internal/journal/corrections_test.go:152` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B6 | if at `internal/journal/corrections_test.go:156` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B7 | if at `internal/journal/corrections_test.go:159` | Focused canonical/temporal ownership regressions plus full and race suites. |

Coverage includes external exclusion, identifier reuse, partial-scope rejection, v19 migration, confirmed temporal ownership, reservation recovery, and operator-release refusal/success.

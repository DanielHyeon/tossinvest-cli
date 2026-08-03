# Branch Test Map: `TestExecutionCorrectionsScopedRequiresPreexistingUniqueConfirmedOwner`

Source: `internal/journal/corrections_test.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | range at `internal/journal/corrections_test.go:165` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B2 | if at `internal/journal/corrections_test.go:178` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B3 | for at `internal/journal/corrections_test.go:185` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B4 | if at `internal/journal/corrections_test.go:190` | Focused canonical/temporal ownership regressions plus full and race suites. |

Coverage includes external exclusion, identifier reuse, partial-scope rejection, v19 migration, confirmed temporal ownership, reservation recovery, and operator-release refusal/success.

# Branch Test Map: `TestReusedOrderIDDoesNotAttachPriorDayFillToNewPositionProvenance`

Source: `internal/journal/provenance_test.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | if at `internal/journal/provenance_test.go:345` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B2 | if at `internal/journal/provenance_test.go:352` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B3 | if at `internal/journal/provenance_test.go:360` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B4 | if at `internal/journal/provenance_test.go:365` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B5 | range at `internal/journal/provenance_test.go:370` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B6 | if at `internal/journal/provenance_test.go:371` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B7 | if at `internal/journal/provenance_test.go:375` | Focused canonical/temporal ownership regressions plus full and race suites. |

Coverage includes external exclusion, identifier reuse, partial-scope rejection, v19 migration, confirmed temporal ownership, reservation recovery, and operator-release refusal/success.

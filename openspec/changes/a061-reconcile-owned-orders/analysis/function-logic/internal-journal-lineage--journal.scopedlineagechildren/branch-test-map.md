# Branch Test Map: `Journal.scopedLineageChildren`

Source: `internal/journal/lineage.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | if at `internal/journal/lineage.go:348` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B2 | if at `internal/journal/lineage.go:352` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B3 | range at `internal/journal/lineage.go:362` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B4 | if at `internal/journal/lineage.go:363` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B5 | if at `internal/journal/lineage.go:368` | Focused canonical/temporal ownership regressions plus full and race suites. |

Coverage includes external exclusion, identifier reuse, partial-scope rejection, v19 migration, confirmed temporal ownership, reservation recovery, and operator-release refusal/success.

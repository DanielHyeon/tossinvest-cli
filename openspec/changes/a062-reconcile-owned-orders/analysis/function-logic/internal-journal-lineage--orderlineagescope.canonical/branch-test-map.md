# Branch Test Map: `OrderLineageScope.canonical`

Source: `internal/journal/lineage.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | switch at `internal/journal/lineage.go:72` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B2 | case at `internal/journal/lineage.go:73` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B3 | case at `internal/journal/lineage.go:75` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B4 | case at `internal/journal/lineage.go:77` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B5 | case at `internal/journal/lineage.go:79` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B6 | case at `internal/journal/lineage.go:81` | Focused canonical/temporal ownership regressions plus full and race suites. |

Coverage includes external exclusion, identifier reuse, partial-scope rejection, v19 migration, confirmed temporal ownership, reservation recovery, and operator-release refusal/success.

# Branch Test Map: `Journal.ResolveCurrentOrderIDScoped`

Source: `internal/journal/lineage.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | if at `internal/journal/lineage.go:313` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B2 | if at `internal/journal/lineage.go:317` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B3 | for at `internal/journal/lineage.go:323` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B4 | if at `internal/journal/lineage.go:325` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B5 | switch at `internal/journal/lineage.go:328` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B6 | case at `internal/journal/lineage.go:329` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B7 | case at `internal/journal/lineage.go:331` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B8 | case at `internal/journal/lineage.go:333` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B9 | if at `internal/journal/lineage.go:338` | Focused canonical/temporal ownership regressions plus full and race suites. |

Coverage includes external exclusion, identifier reuse, partial-scope rejection, v19 migration, confirmed temporal ownership, reservation recovery, and operator-release refusal/success.

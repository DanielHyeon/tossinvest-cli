# Branch Test Map: `Journal.ResolveCurrentOrderID`

Source: `internal/journal/lineage.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | if at `internal/journal/lineage.go:200` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B2 | for at `internal/journal/lineage.go:204` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B3 | if at `internal/journal/lineage.go:206` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B4 | range at `internal/journal/lineage.go:210` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B5 | if at `internal/journal/lineage.go:211` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B6 | if at `internal/journal/lineage.go:216` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B7 | if at `internal/journal/lineage.go:219` | Focused canonical/temporal ownership regressions plus full and race suites. |

Coverage includes external exclusion, identifier reuse, partial-scope rejection, v19 migration, confirmed temporal ownership, reservation recovery, and operator-release refusal/success.

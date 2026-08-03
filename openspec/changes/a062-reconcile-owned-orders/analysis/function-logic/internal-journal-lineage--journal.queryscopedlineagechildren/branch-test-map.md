# Branch Test Map: `Journal.queryScopedLineageChildren`

Source: `internal/journal/lineage.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | if at `internal/journal/lineage.go:423` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B2 | for at `internal/journal/lineage.go:429` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B3 | if at `internal/journal/lineage.go:431` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B4 | if at `internal/journal/lineage.go:436` | Focused canonical/temporal ownership regressions plus full and race suites. |

Coverage includes external exclusion, identifier reuse, partial-scope rejection, v19 migration, confirmed temporal ownership, reservation recovery, and operator-release refusal/success.

# Branch Test Map: `Journal.recordScopedLineageConflict`

Source: `internal/journal/lineage.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | if at `internal/journal/lineage.go:455` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B2 | if at `internal/journal/lineage.go:458` | Focused canonical/temporal ownership regressions plus full and race suites. |

Coverage includes external exclusion, identifier reuse, partial-scope rejection, v19 migration, confirmed temporal ownership, reservation recovery, and operator-release refusal/success.

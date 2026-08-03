# Branch Test Map: `journalTimeStrictlyAfter`

Source: `internal/journal/fills.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | if at `internal/journal/fills.go:617` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B2 | range at `internal/journal/fills.go:620` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B3 | if at `internal/journal/fills.go:621` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B4 | if at `internal/journal/fills.go:625` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B5 | if at `internal/journal/fills.go:628` | Focused canonical/temporal ownership regressions plus full and race suites. |

Coverage includes external exclusion, identifier reuse, partial-scope rejection, v19 migration, confirmed temporal ownership, reservation recovery, and operator-release refusal/success.

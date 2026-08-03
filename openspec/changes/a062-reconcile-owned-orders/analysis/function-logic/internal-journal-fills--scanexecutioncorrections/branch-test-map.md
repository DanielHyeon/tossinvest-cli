# Branch Test Map: `scanExecutionCorrections`

Source: `internal/journal/fills.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | for at `internal/journal/fills.go:1345` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B2 | if at `internal/journal/fills.go:1350` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B3 | if at `internal/journal/fills.go:1359` | Focused canonical/temporal ownership regressions plus full and race suites. |

Coverage includes external exclusion, identifier reuse, partial-scope rejection, v19 migration, confirmed temporal ownership, reservation recovery, and operator-release refusal/success.

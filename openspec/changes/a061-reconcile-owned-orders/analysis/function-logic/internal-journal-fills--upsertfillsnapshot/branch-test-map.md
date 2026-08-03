# Branch Test Map: `upsertFillSnapshot`

Source: `internal/journal/fills.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | if at `internal/journal/fills.go:708` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B2 | if at `internal/journal/fills.go:709` | Focused canonical/temporal ownership regressions plus full and race suites. |

Coverage includes external exclusion, identifier reuse, partial-scope rejection, v19 migration, confirmed temporal ownership, reservation recovery, and operator-release refusal/success.

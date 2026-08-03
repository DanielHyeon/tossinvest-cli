# Branch Test Map: `TestTrackedFillOrdersDoNotBindLegacyLineageSnapshotAcrossReusedDays`

Source: `internal/journal/fills_test.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | if at `internal/journal/fills_test.go:1427` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B2 | if at `internal/journal/fills_test.go:1431` | Focused canonical/temporal ownership regressions plus full and race suites. |

Coverage includes external exclusion, identifier reuse, partial-scope rejection, v19 migration, confirmed temporal ownership, reservation recovery, and operator-release refusal/success.

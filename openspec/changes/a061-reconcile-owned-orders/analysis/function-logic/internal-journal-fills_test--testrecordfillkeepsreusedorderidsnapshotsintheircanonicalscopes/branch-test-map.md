# Branch Test Map: `TestRecordFillKeepsReusedOrderIDSnapshotsInTheirCanonicalScopes`

Source: `internal/journal/fills_test.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | if at `internal/journal/fills_test.go:737` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B2 | if at `internal/journal/fills_test.go:740` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B3 | if at `internal/journal/fills_test.go:747` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B4 | if at `internal/journal/fills_test.go:750` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B5 | if at `internal/journal/fills_test.go:764` | Focused canonical/temporal ownership regressions plus full and race suites. |

Coverage includes external exclusion, identifier reuse, partial-scope rejection, v19 migration, confirmed temporal ownership, reservation recovery, and operator-release refusal/success.

# Branch Test Map: `TestLegacySnapshotCannotTerminateOrOwnAFutureReusedOrder`

Source: `internal/journal/fills_test.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | if at `internal/journal/fills_test.go:1071` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B2 | if at `internal/journal/fills_test.go:1081` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B3 | if at `internal/journal/fills_test.go:1084` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B4 | if at `internal/journal/fills_test.go:1088` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B5 | if at `internal/journal/fills_test.go:1091` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B6 | if at `internal/journal/fills_test.go:1095` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B7 | if at `internal/journal/fills_test.go:1098` | Focused canonical/temporal ownership regressions plus full and race suites. |

Coverage includes external exclusion, identifier reuse, partial-scope rejection, v19 migration, confirmed temporal ownership, reservation recovery, and operator-release refusal/success.

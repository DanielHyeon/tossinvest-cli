# Branch Test Map: `TestScopedExternalSnapshotCannotTerminateOrOwnAFutureOrder`

Source: `internal/journal/fills_test.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | if at `internal/journal/fills_test.go:1106` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B2 | if at `internal/journal/fills_test.go:1116` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B3 | if at `internal/journal/fills_test.go:1119` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B4 | if at `internal/journal/fills_test.go:1123` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B5 | if at `internal/journal/fills_test.go:1126` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B6 | if at `internal/journal/fills_test.go:1130` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B7 | if at `internal/journal/fills_test.go:1133` | Focused canonical/temporal ownership regressions plus full and race suites. |

Coverage includes external exclusion, identifier reuse, partial-scope rejection, v19 migration, confirmed temporal ownership, reservation recovery, and operator-release refusal/success.

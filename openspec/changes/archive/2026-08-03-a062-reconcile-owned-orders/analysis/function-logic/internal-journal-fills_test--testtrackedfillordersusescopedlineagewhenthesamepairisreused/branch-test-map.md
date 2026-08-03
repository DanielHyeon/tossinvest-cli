# Branch Test Map: `TestTrackedFillOrdersUseScopedLineageWhenTheSamePairIsReused`

Source: `internal/journal/fills_test.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | if at `internal/journal/fills_test.go:1390` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B2 | if at `internal/journal/fills_test.go:1395` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B3 | range at `internal/journal/fills_test.go:1399` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B4 | if at `internal/journal/fills_test.go:1400` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B5 | if at `internal/journal/fills_test.go:1402` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B6 | if at `internal/journal/fills_test.go:1407` | Focused canonical/temporal ownership regressions plus full and race suites. |

Coverage includes external exclusion, identifier reuse, partial-scope rejection, v19 migration, confirmed temporal ownership, reservation recovery, and operator-release refusal/success.

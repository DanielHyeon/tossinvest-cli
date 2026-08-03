# Branch Test Map: `TestTrackedAndLiveOrdersAllowReusedIDInDifferentSymbolAndSideScopes`

Source: `internal/journal/fills_test.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | if at `internal/journal/fills_test.go:937` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B2 | if at `internal/journal/fills_test.go:940` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B3 | if at `internal/journal/fills_test.go:943` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B4 | if at `internal/journal/fills_test.go:946` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B5 | if at `internal/journal/fills_test.go:953` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B6 | if at `internal/journal/fills_test.go:958` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B7 | if at `internal/journal/fills_test.go:963` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B8 | if at `internal/journal/fills_test.go:966` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B9 | range at `internal/journal/fills_test.go:969` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B10 | if at `internal/journal/fills_test.go:971` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B11 | if at `internal/journal/fills_test.go:974` | Focused canonical/temporal ownership regressions plus full and race suites. |

Coverage includes external exclusion, identifier reuse, partial-scope rejection, v19 migration, confirmed temporal ownership, reservation recovery, and operator-release refusal/success.

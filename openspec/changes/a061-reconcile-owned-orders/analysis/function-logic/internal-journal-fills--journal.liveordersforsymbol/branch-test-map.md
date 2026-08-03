# Branch Test Map: `Journal.LiveOrdersForSymbol`

Source: `internal/journal/fills.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | if at `internal/journal/fills.go:1832` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B2 | if at `internal/journal/fills.go:1875` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B3 | for at `internal/journal/fills.go:1881` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B4 | if at `internal/journal/fills.go:1883` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B5 | if at `internal/journal/fills.go:1890` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B6 | range at `internal/journal/fills.go:1894` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B7 | if at `internal/journal/fills.go:1902` | Focused canonical/temporal ownership regressions plus full and race suites. |

Coverage includes external exclusion, identifier reuse, partial-scope rejection, v19 migration, confirmed temporal ownership, reservation recovery, and operator-release refusal/success.

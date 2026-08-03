# Branch Test Map: `snapshotTradingDay`

Source: `internal/filldetect/payload.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | if at `internal/filldetect/payload.go:146` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B2 | if at `internal/filldetect/payload.go:150` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B3 | range at `internal/filldetect/payload.go:153` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B4 | if at `internal/filldetect/payload.go:154` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B5 | if at `internal/filldetect/payload.go:156` | Focused canonical/temporal ownership regressions plus full and race suites. |

Coverage includes external exclusion, identifier reuse, partial-scope rejection, v19 migration, confirmed temporal ownership, reservation recovery, and operator-release refusal/success.

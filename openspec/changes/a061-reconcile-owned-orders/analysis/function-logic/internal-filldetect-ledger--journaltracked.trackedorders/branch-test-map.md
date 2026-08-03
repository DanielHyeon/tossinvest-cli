# Branch Test Map: `JournalTracked.TrackedOrders`

Source: `internal/filldetect/ledger.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | if at `internal/filldetect/ledger.go:126` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B2 | if at `internal/filldetect/ledger.go:129` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B3 | if at `internal/filldetect/ledger.go:133` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B4 | range at `internal/filldetect/ledger.go:137` | Focused canonical/temporal ownership regressions plus full and race suites. |

Coverage includes external exclusion, identifier reuse, partial-scope rejection, v19 migration, confirmed temporal ownership, reservation recovery, and operator-release refusal/success.

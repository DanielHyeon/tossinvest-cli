# Branch Test Map: `alertsForOrder`

Source: `internal/journal/reservation_release.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | if at `internal/journal/reservation_release.go:222` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B2 | if at `internal/journal/reservation_release.go:248` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B3 | for at `internal/journal/reservation_release.go:254` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B4 | if at `internal/journal/reservation_release.go:256` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B5 | if at `internal/journal/reservation_release.go:266` | Focused canonical/temporal ownership regressions plus full and race suites. |

Coverage includes external exclusion, identifier reuse, partial-scope rejection, v19 migration, confirmed temporal ownership, reservation recovery, and operator-release refusal/success.

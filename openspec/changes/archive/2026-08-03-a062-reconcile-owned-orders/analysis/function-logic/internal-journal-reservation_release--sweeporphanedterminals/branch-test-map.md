# Branch Test Map: `sweepOrphanedTerminals`

Source: `internal/journal/reservation_release.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | if at `internal/journal/reservation_release.go:540` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B2 | if at `internal/journal/reservation_release.go:583` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B3 | for at `internal/journal/reservation_release.go:587` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B4 | if at `internal/journal/reservation_release.go:589` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B5 | if at `internal/journal/reservation_release.go:597` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B6 | range at `internal/journal/reservation_release.go:606` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B7 | if at `internal/journal/reservation_release.go:607` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B8 | if at `internal/journal/reservation_release.go:612` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B9 | if at `internal/journal/reservation_release.go:621` | Focused canonical/temporal ownership regressions plus full and race suites. |

Coverage includes external exclusion, identifier reuse, partial-scope rejection, v19 migration, confirmed temporal ownership, reservation recovery, and operator-release refusal/success.

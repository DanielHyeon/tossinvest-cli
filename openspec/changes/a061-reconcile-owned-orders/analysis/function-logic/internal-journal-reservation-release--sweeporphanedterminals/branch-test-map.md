# Branch Test Map: `sweepOrphanedTerminals`

Source: `internal/journal/reservation_release.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | if at `internal/journal/reservation_release.go:518` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B2 | if at `internal/journal/reservation_release.go:559` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B3 | for at `internal/journal/reservation_release.go:563` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B4 | if at `internal/journal/reservation_release.go:565` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B5 | if at `internal/journal/reservation_release.go:573` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B6 | range at `internal/journal/reservation_release.go:582` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B7 | if at `internal/journal/reservation_release.go:583` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B8 | if at `internal/journal/reservation_release.go:588` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B9 | if at `internal/journal/reservation_release.go:597` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |

The named focused regressions in this change cover external-order exclusion, canonical reuse, scoped lineage, durable reconcile authority, recovery refusal, startup reservation safety, and held-nonce retention. The full package, race, and repository suites provide integration coverage for all mapped paths.

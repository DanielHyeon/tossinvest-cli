# Branch Test Map: `Tracker.Refresh`

Source: `internal/reconcile/mismatch.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | if at `internal/reconcile/mismatch.go:623` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B2 | if at `internal/reconcile/mismatch.go:631` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B3 | range at `internal/reconcile/mismatch.go:637` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B4 | if at `internal/reconcile/mismatch.go:638` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B5 | range at `internal/reconcile/mismatch.go:642` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B6 | if at `internal/reconcile/mismatch.go:643` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B7 | if at `internal/reconcile/mismatch.go:651` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |

The named focused regressions in this change cover external-order exclusion, canonical reuse, scoped lineage, durable reconcile authority, recovery refusal, startup reservation safety, and held-nonce retention. The full package, race, and repository suites provide integration coverage for all mapped paths.

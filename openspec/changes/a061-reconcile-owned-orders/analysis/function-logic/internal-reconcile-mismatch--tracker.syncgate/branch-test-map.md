# Branch Test Map: `Tracker.syncGate`

Source: `internal/reconcile/mismatch.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | if at `internal/reconcile/mismatch.go:879` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B2 | range at `internal/reconcile/mismatch.go:883` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B3 | if at `internal/reconcile/mismatch.go:884` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B4 | range at `internal/reconcile/mismatch.go:894` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B5 | if at `internal/reconcile/mismatch.go:895` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B6 | range at `internal/reconcile/mismatch.go:901` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B7 | if at `internal/reconcile/mismatch.go:902` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B8 | if at `internal/reconcile/mismatch.go:908` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B9 | range at `internal/reconcile/mismatch.go:913` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B10 | if at `internal/reconcile/mismatch.go:917` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B11 | else at `internal/reconcile/mismatch.go:919` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |

The named focused regressions in this change cover external-order exclusion, canonical reuse, scoped lineage, durable reconcile authority, recovery refusal, startup reservation safety, and held-nonce retention. The full package, race, and repository suites provide integration coverage for all mapped paths.

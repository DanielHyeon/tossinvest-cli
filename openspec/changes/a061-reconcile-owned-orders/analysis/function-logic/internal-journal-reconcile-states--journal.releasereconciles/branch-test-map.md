# Branch Test Map: `Journal.ReleaseReconciles`

Source: `internal/journal/reconcile_states.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | if at `internal/journal/reconcile_states.go:309` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B2 | range at `internal/journal/reconcile_states.go:321` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B3 | switch at `internal/journal/reconcile_states.go:329` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B4 | case at `internal/journal/reconcile_states.go:330` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B5 | case at `internal/journal/reconcile_states.go:332` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B6 | case at `internal/journal/reconcile_states.go:334` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B7 | case at `internal/journal/reconcile_states.go:336` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B8 | if at `internal/journal/reconcile_states.go:340` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B9 | if at `internal/journal/reconcile_states.go:348` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B10 | range at `internal/journal/reconcile_states.go:354` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B11 | if at `internal/journal/reconcile_states.go:357` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B12 | if at `internal/journal/reconcile_states.go:361` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B13 | if at `internal/journal/reconcile_states.go:364` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B14 | range at `internal/journal/reconcile_states.go:374` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B15 | if at `internal/journal/reconcile_states.go:380` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B16 | if at `internal/journal/reconcile_states.go:383` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B17 | if at `internal/journal/reconcile_states.go:384` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B18 | if at `internal/journal/reconcile_states.go:393` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |

The named focused regressions in this change cover external-order exclusion, canonical reuse, scoped lineage, durable reconcile authority, recovery refusal, startup reservation safety, and held-nonce retention. The full package, race, and repository suites provide integration coverage for all mapped paths.

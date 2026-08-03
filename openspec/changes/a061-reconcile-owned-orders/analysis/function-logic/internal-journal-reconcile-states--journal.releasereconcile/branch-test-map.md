# Branch Test Map: `Journal.ReleaseReconcile`

Source: `internal/journal/reconcile_states.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | switch at `internal/journal/reconcile_states.go:251` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B2 | case at `internal/journal/reconcile_states.go:252` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B3 | case at `internal/journal/reconcile_states.go:255` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B4 | case at `internal/journal/reconcile_states.go:258` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B5 | if at `internal/journal/reconcile_states.go:267` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B6 | if at `internal/journal/reconcile_states.go:274` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B7 | if at `internal/journal/reconcile_states.go:277` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B8 | if at `internal/journal/reconcile_states.go:280` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B9 | if at `internal/journal/reconcile_states.go:284` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B10 | if at `internal/journal/reconcile_states.go:290` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |

The named focused regressions in this change cover external-order exclusion, canonical reuse, scoped lineage, durable reconcile authority, recovery refusal, startup reservation safety, and held-nonce retention. The full package, race, and repository suites provide integration coverage for all mapped paths.

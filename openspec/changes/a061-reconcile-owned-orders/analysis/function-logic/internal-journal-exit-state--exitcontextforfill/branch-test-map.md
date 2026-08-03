# Branch Test Map: `exitContextForFill`

Source: `internal/journal/exit_state.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | if at `internal/journal/exit_state.go:875` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B2 | if at `internal/journal/exit_state.go:878` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B3 | if at `internal/journal/exit_state.go:890` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B4 | if at `internal/journal/exit_state.go:895` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B5 | if at `internal/journal/exit_state.go:898` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |

The named focused regressions in this change cover external-order exclusion, canonical reuse, scoped lineage, durable reconcile authority, recovery refusal, startup reservation safety, and held-nonce retention. The full package, race, and repository suites provide integration coverage for all mapped paths.

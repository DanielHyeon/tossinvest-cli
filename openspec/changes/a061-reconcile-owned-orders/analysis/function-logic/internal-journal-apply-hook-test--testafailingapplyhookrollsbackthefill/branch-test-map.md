# Branch Test Map: `TestAFailingApplyHookRollsBackTheFill`

Source: `internal/journal/apply_hook_test.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | if at `internal/journal/apply_hook_test.go:151` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B2 | if at `internal/journal/apply_hook_test.go:154` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B3 | if at `internal/journal/apply_hook_test.go:169` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B4 | if at `internal/journal/apply_hook_test.go:175` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B5 | if at `internal/journal/apply_hook_test.go:179` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B6 | if at `internal/journal/apply_hook_test.go:182` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B7 | if at `internal/journal/apply_hook_test.go:187` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B8 | if at `internal/journal/apply_hook_test.go:191` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B9 | if at `internal/journal/apply_hook_test.go:197` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B10 | if at `internal/journal/apply_hook_test.go:200` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |

The named focused regressions in this change cover external-order exclusion, canonical reuse, scoped lineage, durable reconcile authority, recovery refusal, startup reservation safety, and held-nonce retention. The full package, race, and repository suites provide integration coverage for all mapped paths.

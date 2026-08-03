# Branch Test Map: `TestRestoreProjectsEveryActiveCauseForTheConfiguredAccount`

Source: `internal/reconcile/restore_test.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | range at `internal/reconcile/restore_test.go:373` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B2 | if at `internal/reconcile/restore_test.go:374` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B3 | if at `internal/reconcile/restore_test.go:381` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B4 | if at `internal/reconcile/restore_test.go:386` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B5 | range at `internal/reconcile/restore_test.go:390` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B6 | range at `internal/reconcile/restore_test.go:393` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B7 | if at `internal/reconcile/restore_test.go:400` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B8 | if at `internal/reconcile/restore_test.go:404` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B9 | range at `internal/reconcile/restore_test.go:408` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B10 | if at `internal/reconcile/restore_test.go:414` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B11 | if at `internal/reconcile/restore_test.go:419` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |

The named focused regressions in this change cover external-order exclusion, canonical reuse, scoped lineage, durable reconcile authority, recovery refusal, startup reservation safety, and held-nonce retention. The full package, race, and repository suites provide integration coverage for all mapped paths.

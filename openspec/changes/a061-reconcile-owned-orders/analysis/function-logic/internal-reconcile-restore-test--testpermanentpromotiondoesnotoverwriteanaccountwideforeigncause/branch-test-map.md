# Branch Test Map: `TestPermanentPromotionDoesNotOverwriteAnAccountWideForeignCause`

Source: `internal/reconcile/restore_test.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | if at `internal/reconcile/restore_test.go:565` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B2 | if at `internal/reconcile/restore_test.go:572` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B3 | for at `internal/reconcile/restore_test.go:576` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B4 | if at `internal/reconcile/restore_test.go:577` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B5 | range at `internal/reconcile/restore_test.go:583` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B6 | if at `internal/reconcile/restore_test.go:584` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B7 | if at `internal/reconcile/restore_test.go:589` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B8 | if at `internal/reconcile/restore_test.go:592` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |

The named focused regressions in this change cover external-order exclusion, canonical reuse, scoped lineage, durable reconcile authority, recovery refusal, startup reservation safety, and held-nonce retention. The full package, race, and repository suites provide integration coverage for all mapped paths.

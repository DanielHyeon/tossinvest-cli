# Branch Test Map: `NewReconcileDriver`

Source: `internal/app/engine/reconcileloop.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | switch at `internal/app/engine/reconcileloop.go:280` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B2 | case at `internal/app/engine/reconcileloop.go:281` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B3 | case at `internal/app/engine/reconcileloop.go:283` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B4 | case at `internal/app/engine/reconcileloop.go:285` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B5 | case at `internal/app/engine/reconcileloop.go:287` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B6 | case at `internal/app/engine/reconcileloop.go:289` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B7 | case at `internal/app/engine/reconcileloop.go:292` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B8 | if at `internal/app/engine/reconcileloop.go:296` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B9 | if at `internal/app/engine/reconcileloop.go:297` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B10 | if at `internal/app/engine/reconcileloop.go:309` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |

The named focused regressions in this change cover external-order exclusion, canonical reuse, scoped lineage, durable reconcile authority, recovery refusal, startup reservation safety, and held-nonce retention. The full package, race, and repository suites provide integration coverage for all mapped paths.

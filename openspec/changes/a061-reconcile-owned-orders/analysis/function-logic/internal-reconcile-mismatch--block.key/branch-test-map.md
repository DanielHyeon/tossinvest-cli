# Branch Test Map: `Block.Key`

Source: `internal/reconcile/mismatch.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | switch at `internal/reconcile/mismatch.go:223` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B2 | case at `internal/reconcile/mismatch.go:224` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B3 | case at `internal/reconcile/mismatch.go:226` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B4 | case at `internal/reconcile/mismatch.go:228` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B5 | case at `internal/reconcile/mismatch.go:230` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |

The named focused regressions in this change cover external-order exclusion, canonical reuse, scoped lineage, durable reconcile authority, recovery refusal, startup reservation safety, and held-nonce retention. The full package, race, and repository suites provide integration coverage for all mapped paths.

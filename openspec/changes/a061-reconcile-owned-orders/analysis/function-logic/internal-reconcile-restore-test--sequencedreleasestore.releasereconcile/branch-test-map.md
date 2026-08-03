# Branch Test Map: `sequencedReleaseStore.ReleaseReconcile`

Source: `internal/reconcile/restore_test.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | if at `internal/reconcile/restore_test.go:67` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |

The named focused regressions in this change cover external-order exclusion, canonical reuse, scoped lineage, durable reconcile authority, recovery refusal, startup reservation safety, and held-nonce retention. The full package, race, and repository suites provide integration coverage for all mapped paths.

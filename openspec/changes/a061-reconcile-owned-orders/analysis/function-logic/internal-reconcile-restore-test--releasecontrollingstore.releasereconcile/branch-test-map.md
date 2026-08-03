# Branch Test Map: `releaseControllingStore.ReleaseReconcile`

Source: `internal/reconcile/restore_test.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | Happy path at `internal/reconcile/restore_test.go:96` | Focused package regression plus the full suite verifies the branchless contract. |

The named focused regressions in this change cover external-order exclusion, canonical reuse, scoped lineage, durable reconcile authority, recovery refusal, startup reservation safety, and held-nonce retention. The full package, race, and repository suites provide integration coverage for all mapped paths.

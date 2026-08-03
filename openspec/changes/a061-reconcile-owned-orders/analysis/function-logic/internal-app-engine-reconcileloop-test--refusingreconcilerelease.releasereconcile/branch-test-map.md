# Branch Test Map: `refusingReconcileRelease.ReleaseReconcile`

Source: `internal/app/engine/reconcileloop_test.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | Happy path at `internal/app/engine/reconcileloop_test.go:100` | Focused package regression plus the full suite verifies the branchless contract. |

The named focused regressions in this change cover external-order exclusion, canonical reuse, scoped lineage, durable reconcile authority, recovery refusal, startup reservation safety, and held-nonce retention. The full package, race, and repository suites provide integration coverage for all mapped paths.

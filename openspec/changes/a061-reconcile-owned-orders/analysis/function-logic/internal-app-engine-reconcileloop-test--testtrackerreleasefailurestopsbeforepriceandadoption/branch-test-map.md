# Branch Test Map: `TestTrackerReleaseFailureStopsBeforePriceAndAdoption`

Source: `internal/app/engine/reconcileloop_test.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | if at `internal/app/engine/reconcileloop_test.go:403` | Focused regressions and the full/race suites verify this condition and its error path. |
| B2 | if at `internal/app/engine/reconcileloop_test.go:409` | Focused regressions and the full/race suites verify this condition and its error path. |
| B3 | if at `internal/app/engine/reconcileloop_test.go:416` | Focused regressions and the full/race suites verify this condition and its error path. |
| B4 | if at `internal/app/engine/reconcileloop_test.go:419` | Focused regressions and the full/race suites verify this condition and its error path. |
| B5 | if at `internal/app/engine/reconcileloop_test.go:422` | Focused regressions and the full/race suites verify this condition and its error path. |

Coverage includes external-order exclusion, scoped snapshot coexistence, detector reuse, incomplete broker evidence, scoped lineage, durable reconcile authority, recovery refusal/release, reservation decision binding, and nonce retention.

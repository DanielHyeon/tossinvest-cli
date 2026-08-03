# Branch Test Map: `TestIncludeOnlyAdoptionRequiresAPriceReader`

Source: `internal/app/engine/reconcileloop_test.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | if at `internal/app/engine/reconcileloop_test.go:435` | Focused regressions and the full/race suites verify this condition and its error path. |
| B2 | if at `internal/app/engine/reconcileloop_test.go:446` | Focused regressions and the full/race suites verify this condition and its error path. |

Coverage includes external-order exclusion, scoped snapshot coexistence, detector reuse, incomplete broker evidence, scoped lineage, durable reconcile authority, recovery refusal/release, reservation decision binding, and nonce retention.

# Branch Test Map: `ReconcileDriver.RunOnce`

Source: `internal/app/engine/reconcileloop.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | if at `internal/app/engine/reconcileloop.go:387` | Focused regressions and the full/race suites verify this condition and its error path. |
| B2 | if at `internal/app/engine/reconcileloop.go:393` | Focused regressions and the full/race suites verify this condition and its error path. |
| B3 | if at `internal/app/engine/reconcileloop.go:404` | Focused regressions and the full/race suites verify this condition and its error path. |
| B4 | if at `internal/app/engine/reconcileloop.go:410` | Focused regressions and the full/race suites verify this condition and its error path. |
| B5 | if at `internal/app/engine/reconcileloop.go:416` | Focused regressions and the full/race suites verify this condition and its error path. |
| B6 | if at `internal/app/engine/reconcileloop.go:417` | Focused regressions and the full/race suites verify this condition and its error path. |
| B7 | if at `internal/app/engine/reconcileloop.go:425` | Focused regressions and the full/race suites verify this condition and its error path. |
| B8 | if at `internal/app/engine/reconcileloop.go:426` | Focused regressions and the full/race suites verify this condition and its error path. |

Coverage includes external-order exclusion, scoped snapshot coexistence, detector reuse, incomplete broker evidence, scoped lineage, durable reconcile authority, recovery refusal/release, reservation decision binding, and nonce retention.

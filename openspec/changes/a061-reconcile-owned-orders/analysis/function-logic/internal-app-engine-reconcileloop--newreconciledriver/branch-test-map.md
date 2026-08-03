# Branch Test Map: `NewReconcileDriver`

Source: `internal/app/engine/reconcileloop.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | switch at `internal/app/engine/reconcileloop.go:280` | Focused regressions and the full/race suites verify this condition and its error path. |
| B2 | case at `internal/app/engine/reconcileloop.go:281` | Focused regressions and the full/race suites verify this condition and its error path. |
| B3 | case at `internal/app/engine/reconcileloop.go:283` | Focused regressions and the full/race suites verify this condition and its error path. |
| B4 | case at `internal/app/engine/reconcileloop.go:285` | Focused regressions and the full/race suites verify this condition and its error path. |
| B5 | case at `internal/app/engine/reconcileloop.go:287` | Focused regressions and the full/race suites verify this condition and its error path. |
| B6 | case at `internal/app/engine/reconcileloop.go:289` | Focused regressions and the full/race suites verify this condition and its error path. |
| B7 | case at `internal/app/engine/reconcileloop.go:292` | Focused regressions and the full/race suites verify this condition and its error path. |
| B8 | if at `internal/app/engine/reconcileloop.go:296` | Focused regressions and the full/race suites verify this condition and its error path. |
| B9 | if at `internal/app/engine/reconcileloop.go:297` | Focused regressions and the full/race suites verify this condition and its error path. |
| B10 | if at `internal/app/engine/reconcileloop.go:309` | Focused regressions and the full/race suites verify this condition and its error path. |

Coverage includes external-order exclusion, scoped snapshot coexistence, detector reuse, incomplete broker evidence, scoped lineage, durable reconcile authority, recovery refusal/release, reservation decision binding, and nonce retention.

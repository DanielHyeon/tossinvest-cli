# Branch Test Map: `Journal.ReleaseReconcile`

Source: `internal/journal/reconcile_states.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | switch at `internal/journal/reconcile_states.go:251` | Focused regressions and the full/race suites verify this condition and its error path. |
| B2 | case at `internal/journal/reconcile_states.go:252` | Focused regressions and the full/race suites verify this condition and its error path. |
| B3 | case at `internal/journal/reconcile_states.go:255` | Focused regressions and the full/race suites verify this condition and its error path. |
| B4 | case at `internal/journal/reconcile_states.go:258` | Focused regressions and the full/race suites verify this condition and its error path. |
| B5 | if at `internal/journal/reconcile_states.go:267` | Focused regressions and the full/race suites verify this condition and its error path. |
| B6 | if at `internal/journal/reconcile_states.go:274` | Focused regressions and the full/race suites verify this condition and its error path. |
| B7 | if at `internal/journal/reconcile_states.go:277` | Focused regressions and the full/race suites verify this condition and its error path. |
| B8 | if at `internal/journal/reconcile_states.go:280` | Focused regressions and the full/race suites verify this condition and its error path. |
| B9 | if at `internal/journal/reconcile_states.go:284` | Focused regressions and the full/race suites verify this condition and its error path. |
| B10 | if at `internal/journal/reconcile_states.go:290` | Focused regressions and the full/race suites verify this condition and its error path. |

Coverage includes external-order exclusion, scoped snapshot coexistence, detector reuse, incomplete broker evidence, scoped lineage, durable reconcile authority, recovery refusal/release, reservation decision binding, and nonce retention.

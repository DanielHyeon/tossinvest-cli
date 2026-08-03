# Branch Test Map: `Journal.ReleaseReconciles`

Source: `internal/journal/reconcile_states.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | if at `internal/journal/reconcile_states.go:309` | Focused regressions and the full/race suites verify this condition and its error path. |
| B2 | range at `internal/journal/reconcile_states.go:321` | Focused regressions and the full/race suites verify this condition and its error path. |
| B3 | switch at `internal/journal/reconcile_states.go:329` | Focused regressions and the full/race suites verify this condition and its error path. |
| B4 | case at `internal/journal/reconcile_states.go:330` | Focused regressions and the full/race suites verify this condition and its error path. |
| B5 | case at `internal/journal/reconcile_states.go:332` | Focused regressions and the full/race suites verify this condition and its error path. |
| B6 | case at `internal/journal/reconcile_states.go:334` | Focused regressions and the full/race suites verify this condition and its error path. |
| B7 | case at `internal/journal/reconcile_states.go:336` | Focused regressions and the full/race suites verify this condition and its error path. |
| B8 | if at `internal/journal/reconcile_states.go:340` | Focused regressions and the full/race suites verify this condition and its error path. |
| B9 | if at `internal/journal/reconcile_states.go:348` | Focused regressions and the full/race suites verify this condition and its error path. |
| B10 | range at `internal/journal/reconcile_states.go:354` | Focused regressions and the full/race suites verify this condition and its error path. |
| B11 | if at `internal/journal/reconcile_states.go:357` | Focused regressions and the full/race suites verify this condition and its error path. |
| B12 | if at `internal/journal/reconcile_states.go:361` | Focused regressions and the full/race suites verify this condition and its error path. |
| B13 | if at `internal/journal/reconcile_states.go:364` | Focused regressions and the full/race suites verify this condition and its error path. |
| B14 | range at `internal/journal/reconcile_states.go:374` | Focused regressions and the full/race suites verify this condition and its error path. |
| B15 | if at `internal/journal/reconcile_states.go:380` | Focused regressions and the full/race suites verify this condition and its error path. |
| B16 | if at `internal/journal/reconcile_states.go:383` | Focused regressions and the full/race suites verify this condition and its error path. |
| B17 | if at `internal/journal/reconcile_states.go:384` | Focused regressions and the full/race suites verify this condition and its error path. |
| B18 | if at `internal/journal/reconcile_states.go:393` | Focused regressions and the full/race suites verify this condition and its error path. |

Coverage includes external-order exclusion, scoped snapshot coexistence, detector reuse, incomplete broker evidence, scoped lineage, durable reconcile authority, recovery refusal/release, reservation decision binding, and nonce retention.

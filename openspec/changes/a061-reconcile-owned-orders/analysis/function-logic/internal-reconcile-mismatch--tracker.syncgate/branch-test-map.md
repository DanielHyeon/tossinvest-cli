# Branch Test Map: `Tracker.syncGate`

Source: `internal/reconcile/mismatch.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | if at `internal/reconcile/mismatch.go:879` | Focused regressions and the full/race suites verify this condition and its error path. |
| B2 | range at `internal/reconcile/mismatch.go:883` | Focused regressions and the full/race suites verify this condition and its error path. |
| B3 | if at `internal/reconcile/mismatch.go:884` | Focused regressions and the full/race suites verify this condition and its error path. |
| B4 | range at `internal/reconcile/mismatch.go:894` | Focused regressions and the full/race suites verify this condition and its error path. |
| B5 | if at `internal/reconcile/mismatch.go:895` | Focused regressions and the full/race suites verify this condition and its error path. |
| B6 | range at `internal/reconcile/mismatch.go:901` | Focused regressions and the full/race suites verify this condition and its error path. |
| B7 | if at `internal/reconcile/mismatch.go:902` | Focused regressions and the full/race suites verify this condition and its error path. |
| B8 | if at `internal/reconcile/mismatch.go:908` | Focused regressions and the full/race suites verify this condition and its error path. |
| B9 | range at `internal/reconcile/mismatch.go:913` | Focused regressions and the full/race suites verify this condition and its error path. |
| B10 | if at `internal/reconcile/mismatch.go:917` | Focused regressions and the full/race suites verify this condition and its error path. |
| B11 | else at `internal/reconcile/mismatch.go:919` | Focused regressions and the full/race suites verify this condition and its error path. |

Coverage includes external-order exclusion, scoped snapshot coexistence, detector reuse, incomplete broker evidence, scoped lineage, durable reconcile authority, recovery refusal/release, reservation decision binding, and nonce retention.

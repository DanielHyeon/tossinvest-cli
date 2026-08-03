# Branch Test Map: `Tracker.Observe`

Source: `internal/reconcile/mismatch.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | if at `internal/reconcile/mismatch.go:366` | Focused regressions and the full/race suites verify this condition and its error path. |
| B2 | if at `internal/reconcile/mismatch.go:373` | Focused regressions and the full/race suites verify this condition and its error path. |
| B3 | else at `internal/reconcile/mismatch.go:396` | Focused regressions and the full/race suites verify this condition and its error path. |
| B4 | range at `internal/reconcile/mismatch.go:380` | Focused regressions and the full/race suites verify this condition and its error path. |
| B5 | if at `internal/reconcile/mismatch.go:381` | Focused regressions and the full/race suites verify this condition and its error path. |
| B6 | if at `internal/reconcile/mismatch.go:388` | Focused regressions and the full/race suites verify this condition and its error path. |
| B7 | range at `internal/reconcile/mismatch.go:398` | Focused regressions and the full/race suites verify this condition and its error path. |
| B8 | if at `internal/reconcile/mismatch.go:399` | Focused regressions and the full/race suites verify this condition and its error path. |
| B9 | if at `internal/reconcile/mismatch.go:405` | Focused regressions and the full/race suites verify this condition and its error path. |
| B10 | if at `internal/reconcile/mismatch.go:418` | Focused regressions and the full/race suites verify this condition and its error path. |
| B11 | range at `internal/reconcile/mismatch.go:440` | Focused regressions and the full/race suites verify this condition and its error path. |
| B12 | range at `internal/reconcile/mismatch.go:443` | Focused regressions and the full/race suites verify this condition and its error path. |
| B13 | if at `internal/reconcile/mismatch.go:444` | Focused regressions and the full/race suites verify this condition and its error path. |
| B14 | range at `internal/reconcile/mismatch.go:449` | Focused regressions and the full/race suites verify this condition and its error path. |
| B15 | range at `internal/reconcile/mismatch.go:453` | Focused regressions and the full/race suites verify this condition and its error path. |
| B16 | range at `internal/reconcile/mismatch.go:456` | Focused regressions and the full/race suites verify this condition and its error path. |
| B17 | if at `internal/reconcile/mismatch.go:462` | Focused regressions and the full/race suites verify this condition and its error path. |
| B18 | else at `internal/reconcile/mismatch.go:475` | Focused regressions and the full/race suites verify this condition and its error path. |
| B19 | range at `internal/reconcile/mismatch.go:467` | Focused regressions and the full/race suites verify this condition and its error path. |
| B20 | if at `internal/reconcile/mismatch.go:468` | Focused regressions and the full/race suites verify this condition and its error path. |

Coverage includes external-order exclusion, scoped snapshot coexistence, detector reuse, incomplete broker evidence, scoped lineage, durable reconcile authority, recovery refusal/release, reservation decision binding, and nonce retention.

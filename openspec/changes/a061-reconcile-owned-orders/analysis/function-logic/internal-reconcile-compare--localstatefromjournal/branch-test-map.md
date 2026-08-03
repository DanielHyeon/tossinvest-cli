# Branch Test Map: `LocalStateFromJournal`

Source: `internal/reconcile/compare.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | if at `internal/reconcile/compare.go:199` | Focused regressions and the full/race suites verify this condition and its error path. |
| B2 | if at `internal/reconcile/compare.go:203` | Focused regressions and the full/race suites verify this condition and its error path. |
| B3 | if at `internal/reconcile/compare.go:207` | Focused regressions and the full/race suites verify this condition and its error path. |
| B4 | range at `internal/reconcile/compare.go:217` | Focused regressions and the full/race suites verify this condition and its error path. |
| B5 | if at `internal/reconcile/compare.go:225` | Focused regressions and the full/race suites verify this condition and its error path. |

Coverage includes external-order exclusion, scoped snapshot coexistence, detector reuse, incomplete broker evidence, scoped lineage, durable reconcile authority, recovery refusal/release, reservation decision binding, and nonce retention.

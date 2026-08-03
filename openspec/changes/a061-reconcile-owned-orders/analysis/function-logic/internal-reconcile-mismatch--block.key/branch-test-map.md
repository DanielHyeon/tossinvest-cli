# Branch Test Map: `Block.Key`

Source: `internal/reconcile/mismatch.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | switch at `internal/reconcile/mismatch.go:223` | Focused regressions and the full/race suites verify this condition and its error path. |
| B2 | case at `internal/reconcile/mismatch.go:224` | Focused regressions and the full/race suites verify this condition and its error path. |
| B3 | case at `internal/reconcile/mismatch.go:226` | Focused regressions and the full/race suites verify this condition and its error path. |
| B4 | case at `internal/reconcile/mismatch.go:228` | Focused regressions and the full/race suites verify this condition and its error path. |
| B5 | case at `internal/reconcile/mismatch.go:230` | Focused regressions and the full/race suites verify this condition and its error path. |

Coverage includes external-order exclusion, scoped snapshot coexistence, detector reuse, incomplete broker evidence, scoped lineage, durable reconcile authority, recovery refusal/release, reservation decision binding, and nonce retention.

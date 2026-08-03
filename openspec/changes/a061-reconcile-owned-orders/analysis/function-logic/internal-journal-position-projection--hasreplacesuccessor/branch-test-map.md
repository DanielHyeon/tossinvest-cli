# Branch Test Map: `hasReplaceSuccessor`

Source: `internal/journal/position_projection.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | if at `internal/journal/position_projection.go:243` | Focused regressions and the full/race suites verify this condition and its error path. |
| B2 | if at `internal/journal/position_projection.go:248` | Focused regressions and the full/race suites verify this condition and its error path. |

Coverage includes external-order exclusion, scoped snapshot coexistence, detector reuse, incomplete broker evidence, scoped lineage, durable reconcile authority, recovery refusal/release, reservation decision binding, and nonce retention.

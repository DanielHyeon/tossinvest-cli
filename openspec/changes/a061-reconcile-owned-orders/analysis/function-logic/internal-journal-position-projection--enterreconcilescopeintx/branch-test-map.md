# Branch Test Map: `enterReconcileScopeInTx`

Source: `internal/journal/position_projection.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | if at `internal/journal/position_projection.go:389` | Focused regressions and the full/race suites verify this condition and its error path. |
| B2 | if at `internal/journal/position_projection.go:396` | Focused regressions and the full/race suites verify this condition and its error path. |
| B3 | if at `internal/journal/position_projection.go:400` | Focused regressions and the full/race suites verify this condition and its error path. |
| B4 | if at `internal/journal/position_projection.go:407` | Focused regressions and the full/race suites verify this condition and its error path. |

Coverage includes external-order exclusion, scoped snapshot coexistence, detector reuse, incomplete broker evidence, scoped lineage, durable reconcile authority, recovery refusal/release, reservation decision binding, and nonce retention.

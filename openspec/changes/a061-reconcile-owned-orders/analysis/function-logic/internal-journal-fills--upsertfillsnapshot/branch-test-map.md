# Branch Test Map: `upsertFillSnapshot`

Source: `internal/journal/fills.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | if at `internal/journal/fills.go:562` | Focused regressions and the full/race suites verify this condition and its error path. |
| B2 | if at `internal/journal/fills.go:563` | Focused regressions and the full/race suites verify this condition and its error path. |

Coverage includes external-order exclusion, scoped snapshot coexistence, detector reuse, incomplete broker evidence, scoped lineage, durable reconcile authority, recovery refusal/release, reservation decision binding, and nonce retention.

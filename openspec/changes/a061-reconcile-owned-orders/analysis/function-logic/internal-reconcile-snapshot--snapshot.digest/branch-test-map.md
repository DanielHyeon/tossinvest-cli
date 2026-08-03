# Branch Test Map: `Snapshot.Digest`

Source: `internal/reconcile/snapshot.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | range at `internal/reconcile/snapshot.go:169` | Focused regressions and the full/race suites verify this condition and its error path. |
| B2 | range at `internal/reconcile/snapshot.go:180` | Focused regressions and the full/race suites verify this condition and its error path. |
| B3 | range at `internal/reconcile/snapshot.go:187` | Focused regressions and the full/race suites verify this condition and its error path. |

Coverage includes external-order exclusion, scoped snapshot coexistence, detector reuse, incomplete broker evidence, scoped lineage, durable reconcile authority, recovery refusal/release, reservation decision binding, and nonce retention.

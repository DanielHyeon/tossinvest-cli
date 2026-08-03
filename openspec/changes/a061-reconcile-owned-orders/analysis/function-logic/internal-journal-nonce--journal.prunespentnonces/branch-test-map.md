# Branch Test Map: `Journal.PruneSpentNonces`

Source: `internal/journal/nonce.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | if at `internal/journal/nonce.go:151` | Focused regressions and the full/race suites verify this condition and its error path. |
| B2 | if at `internal/journal/nonce.go:155` | Focused regressions and the full/race suites verify this condition and its error path. |
| B3 | if at `internal/journal/nonce.go:158` | Focused regressions and the full/race suites verify this condition and its error path. |
| B4 | if at `internal/journal/nonce.go:172` | Focused regressions and the full/race suites verify this condition and its error path. |
| B5 | if at `internal/journal/nonce.go:176` | Focused regressions and the full/race suites verify this condition and its error path. |

Coverage includes external-order exclusion, scoped snapshot coexistence, detector reuse, incomplete broker evidence, scoped lineage, durable reconcile authority, recovery refusal/release, reservation decision binding, and nonce retention.

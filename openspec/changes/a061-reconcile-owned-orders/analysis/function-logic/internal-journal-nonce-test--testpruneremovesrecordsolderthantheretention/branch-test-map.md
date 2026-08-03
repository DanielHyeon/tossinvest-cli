# Branch Test Map: `TestPruneRemovesRecordsOlderThanTheRetention`

Source: `internal/journal/nonce_test.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | if at `internal/journal/nonce_test.go:234` | Focused regressions and the full/race suites verify this condition and its error path. |
| B2 | if at `internal/journal/nonce_test.go:241` | Focused regressions and the full/race suites verify this condition and its error path. |
| B3 | if at `internal/journal/nonce_test.go:244` | Focused regressions and the full/race suites verify this condition and its error path. |
| B4 | if at `internal/journal/nonce_test.go:250` | Focused regressions and the full/race suites verify this condition and its error path. |
| B5 | if at `internal/journal/nonce_test.go:253` | Focused regressions and the full/race suites verify this condition and its error path. |

Coverage includes external-order exclusion, scoped snapshot coexistence, detector reuse, incomplete broker evidence, scoped lineage, durable reconcile authority, recovery refusal/release, reservation decision binding, and nonce retention.

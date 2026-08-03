# Branch Test Map: `TestPrunePreservesSpentNonceWhileItsReservationIsHeld`

Source: `internal/journal/nonce_test.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | if at `internal/journal/nonce_test.go:263` | Focused regressions and the full/race suites verify this condition and its error path. |
| B2 | if at `internal/journal/nonce_test.go:271` | Focused regressions and the full/race suites verify this condition and its error path. |
| B3 | if at `internal/journal/nonce_test.go:274` | Focused regressions and the full/race suites verify this condition and its error path. |

Coverage includes external-order exclusion, scoped snapshot coexistence, detector reuse, incomplete broker evidence, scoped lineage, durable reconcile authority, recovery refusal/release, reservation decision binding, and nonce retention.

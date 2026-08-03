# Branch Test Map: `TestDerivedTerminalFillReleasesTheHold`

Source: `internal/journal/reservation_release_test.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | if at `internal/journal/reservation_release_test.go:167` | Focused regressions and the full/race suites verify this condition and its error path. |
| B2 | if at `internal/journal/reservation_release_test.go:179` | Focused regressions and the full/race suites verify this condition and its error path. |
| B3 | if at `internal/journal/reservation_release_test.go:182` | Focused regressions and the full/race suites verify this condition and its error path. |
| B4 | if at `internal/journal/reservation_release_test.go:186` | Focused regressions and the full/race suites verify this condition and its error path. |

Coverage includes external-order exclusion, scoped snapshot coexistence, detector reuse, incomplete broker evidence, scoped lineage, durable reconcile authority, recovery refusal/release, reservation decision binding, and nonce retention.

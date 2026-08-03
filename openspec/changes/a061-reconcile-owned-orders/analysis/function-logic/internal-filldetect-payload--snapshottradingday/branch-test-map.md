# Branch Test Map: `snapshotTradingDay`

Source: `internal/filldetect/payload.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | if at `internal/filldetect/payload.go:146` | Focused regressions and the full/race suites verify this condition and its error path. |
| B2 | if at `internal/filldetect/payload.go:150` | Focused regressions and the full/race suites verify this condition and its error path. |
| B3 | range at `internal/filldetect/payload.go:153` | Focused regressions and the full/race suites verify this condition and its error path. |
| B4 | if at `internal/filldetect/payload.go:154` | Focused regressions and the full/race suites verify this condition and its error path. |
| B5 | if at `internal/filldetect/payload.go:156` | Focused regressions and the full/race suites verify this condition and its error path. |

Coverage includes external-order exclusion, scoped snapshot coexistence, detector reuse, incomplete broker evidence, scoped lineage, durable reconcile authority, recovery refusal/release, reservation decision binding, and nonce retention.

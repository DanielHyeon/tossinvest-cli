# Branch Test Map: `fillMarketFromCurrency`

Source: `internal/filldetect/payload.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | switch at `internal/filldetect/payload.go:134` | Focused regressions and the full/race suites verify this condition and its error path. |
| B2 | case at `internal/filldetect/payload.go:135` | Focused regressions and the full/race suites verify this condition and its error path. |
| B3 | case at `internal/filldetect/payload.go:137` | Focused regressions and the full/race suites verify this condition and its error path. |
| B4 | case at `internal/filldetect/payload.go:139` | Focused regressions and the full/race suites verify this condition and its error path. |

Coverage includes external-order exclusion, scoped snapshot coexistence, detector reuse, incomplete broker evidence, scoped lineage, durable reconcile authority, recovery refusal/release, reservation decision binding, and nonce retention.

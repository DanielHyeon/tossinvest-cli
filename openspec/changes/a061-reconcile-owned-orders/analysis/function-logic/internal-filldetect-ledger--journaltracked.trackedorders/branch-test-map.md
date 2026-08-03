# Branch Test Map: `JournalTracked.TrackedOrders`

Source: `internal/filldetect/ledger.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | if at `internal/filldetect/ledger.go:126` | Focused regressions and the full/race suites verify this condition and its error path. |
| B2 | if at `internal/filldetect/ledger.go:129` | Focused regressions and the full/race suites verify this condition and its error path. |
| B3 | if at `internal/filldetect/ledger.go:133` | Focused regressions and the full/race suites verify this condition and its error path. |
| B4 | range at `internal/filldetect/ledger.go:137` | Focused regressions and the full/race suites verify this condition and its error path. |

Coverage includes external-order exclusion, scoped snapshot coexistence, detector reuse, incomplete broker evidence, scoped lineage, durable reconcile authority, recovery refusal/release, reservation decision binding, and nonce retention.

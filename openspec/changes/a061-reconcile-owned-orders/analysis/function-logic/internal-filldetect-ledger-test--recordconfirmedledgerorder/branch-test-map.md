# Branch Test Map: `recordConfirmedLedgerOrder`

Source: `internal/filldetect/ledger_test.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | if at `internal/filldetect/ledger_test.go:64` | Focused regressions and the full/race suites verify this condition and its error path. |
| B2 | if at `internal/filldetect/ledger_test.go:67` | Focused regressions and the full/race suites verify this condition and its error path. |
| B3 | if at `internal/filldetect/ledger_test.go:70` | Focused regressions and the full/race suites verify this condition and its error path. |
| B4 | if at `internal/filldetect/ledger_test.go:73` | Focused regressions and the full/race suites verify this condition and its error path. |

Coverage includes external-order exclusion, scoped snapshot coexistence, detector reuse, incomplete broker evidence, scoped lineage, durable reconcile authority, recovery refusal/release, reservation decision binding, and nonce retention.

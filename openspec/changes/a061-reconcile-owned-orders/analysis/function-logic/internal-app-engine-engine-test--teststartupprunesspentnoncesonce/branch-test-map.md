# Branch Test Map: `TestStartupPrunesSpentNoncesOnce`

Source: `internal/app/engine/engine_test.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | if at `internal/app/engine/engine_test.go:451` | Focused regressions and the full/race suites verify this condition and its error path. |
| B2 | if at `internal/app/engine/engine_test.go:456` | Focused regressions and the full/race suites verify this condition and its error path. |
| B3 | if at `internal/app/engine/engine_test.go:460` | Focused regressions and the full/race suites verify this condition and its error path. |

Coverage includes external-order exclusion, scoped snapshot coexistence, detector reuse, incomplete broker evidence, scoped lineage, durable reconcile authority, recovery refusal/release, reservation decision binding, and nonce retention.

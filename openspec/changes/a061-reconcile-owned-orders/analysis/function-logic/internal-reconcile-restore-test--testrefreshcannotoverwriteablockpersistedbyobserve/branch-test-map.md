# Branch Test Map: `TestRefreshCannotOverwriteABlockPersistedByObserve`

Source: `internal/reconcile/restore_test.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | select at `internal/reconcile/restore_test.go:291` | Focused regressions and the full/race suites verify this condition and its error path. |
| B2 | if at `internal/reconcile/restore_test.go:299` | Focused regressions and the full/race suites verify this condition and its error path. |
| B3 | if at `internal/reconcile/restore_test.go:302` | Focused regressions and the full/race suites verify this condition and its error path. |
| B4 | if at `internal/reconcile/restore_test.go:305` | Focused regressions and the full/race suites verify this condition and its error path. |
| B5 | if at `internal/reconcile/restore_test.go:308` | Focused regressions and the full/race suites verify this condition and its error path. |

Coverage includes external-order exclusion, scoped snapshot coexistence, detector reuse, incomplete broker evidence, scoped lineage, durable reconcile authority, recovery refusal/release, reservation decision binding, and nonce retention.

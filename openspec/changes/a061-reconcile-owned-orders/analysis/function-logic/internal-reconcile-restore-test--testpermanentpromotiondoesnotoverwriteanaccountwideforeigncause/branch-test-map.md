# Branch Test Map: `TestPermanentPromotionDoesNotOverwriteAnAccountWideForeignCause`

Source: `internal/reconcile/restore_test.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | if at `internal/reconcile/restore_test.go:565` | Focused regressions and the full/race suites verify this condition and its error path. |
| B2 | if at `internal/reconcile/restore_test.go:572` | Focused regressions and the full/race suites verify this condition and its error path. |
| B3 | for at `internal/reconcile/restore_test.go:576` | Focused regressions and the full/race suites verify this condition and its error path. |
| B4 | if at `internal/reconcile/restore_test.go:577` | Focused regressions and the full/race suites verify this condition and its error path. |
| B5 | range at `internal/reconcile/restore_test.go:583` | Focused regressions and the full/race suites verify this condition and its error path. |
| B6 | if at `internal/reconcile/restore_test.go:584` | Focused regressions and the full/race suites verify this condition and its error path. |
| B7 | if at `internal/reconcile/restore_test.go:589` | Focused regressions and the full/race suites verify this condition and its error path. |
| B8 | if at `internal/reconcile/restore_test.go:592` | Focused regressions and the full/race suites verify this condition and its error path. |

Coverage includes external-order exclusion, scoped snapshot coexistence, detector reuse, incomplete broker evidence, scoped lineage, durable reconcile authority, recovery refusal/release, reservation decision binding, and nonce retention.

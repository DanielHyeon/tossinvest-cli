# Branch Test Map: `TestObservePublishesOnlyDurablePartialReleasesAndRetriesTheRemainder`

Source: `internal/reconcile/restore_test.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | range at `internal/reconcile/restore_test.go:478` | Focused regressions and the full/race suites verify this condition and its error path. |
| B2 | if at `internal/reconcile/restore_test.go:479` | Focused regressions and the full/race suites verify this condition and its error path. |
| B3 | if at `internal/reconcile/restore_test.go:488` | Focused regressions and the full/race suites verify this condition and its error path. |
| B4 | if at `internal/reconcile/restore_test.go:496` | Focused regressions and the full/race suites verify this condition and its error path. |
| B5 | if at `internal/reconcile/restore_test.go:499` | Focused regressions and the full/race suites verify this condition and its error path. |
| B6 | if at `internal/reconcile/restore_test.go:502` | Focused regressions and the full/race suites verify this condition and its error path. |
| B7 | if at `internal/reconcile/restore_test.go:506` | Focused regressions and the full/race suites verify this condition and its error path. |
| B8 | if at `internal/reconcile/restore_test.go:509` | Focused regressions and the full/race suites verify this condition and its error path. |
| B9 | if at `internal/reconcile/restore_test.go:512` | Focused regressions and the full/race suites verify this condition and its error path. |
| B10 | if at `internal/reconcile/restore_test.go:518` | Focused regressions and the full/race suites verify this condition and its error path. |
| B11 | if at `internal/reconcile/restore_test.go:521` | Focused regressions and the full/race suites verify this condition and its error path. |
| B12 | if at `internal/reconcile/restore_test.go:524` | Focused regressions and the full/race suites verify this condition and its error path. |

Coverage includes external-order exclusion, scoped snapshot coexistence, detector reuse, incomplete broker evidence, scoped lineage, durable reconcile authority, recovery refusal/release, reservation decision binding, and nonce retention.

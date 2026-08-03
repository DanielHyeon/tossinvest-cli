# Branch Test Map: `TestReleasingNothingIsNotAnError`

Source: `internal/journal/reconcile_states_test.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | if at `internal/journal/reconcile_states_test.go:285` | Focused regressions and the full/race suites verify this condition and its error path. |
| B2 | if at `internal/journal/reconcile_states_test.go:288` | Focused regressions and the full/race suites verify this condition and its error path. |

Coverage includes external-order exclusion, scoped snapshot coexistence, detector reuse, incomplete broker evidence, scoped lineage, durable reconcile authority, recovery refusal/release, reservation decision binding, and nonce retention.

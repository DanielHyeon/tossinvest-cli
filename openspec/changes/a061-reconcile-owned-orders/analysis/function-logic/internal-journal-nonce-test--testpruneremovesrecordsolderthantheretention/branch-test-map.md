# Branch Test Map: `TestPruneRemovesRecordsOlderThanTheRetention`

Source: `internal/journal/nonce_test.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | if at `internal/journal/nonce_test.go:234` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B2 | if at `internal/journal/nonce_test.go:241` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B3 | if at `internal/journal/nonce_test.go:244` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B4 | if at `internal/journal/nonce_test.go:250` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B5 | if at `internal/journal/nonce_test.go:253` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |

The named focused regressions in this change cover external-order exclusion, canonical reuse, scoped lineage, durable reconcile authority, recovery refusal, startup reservation safety, and held-nonce retention. The full package, race, and repository suites provide integration coverage for all mapped paths.

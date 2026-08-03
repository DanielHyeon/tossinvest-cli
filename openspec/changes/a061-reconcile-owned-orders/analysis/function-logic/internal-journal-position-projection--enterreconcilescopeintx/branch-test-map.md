# Branch Test Map: `enterReconcileScopeInTx`

Source: `internal/journal/position_projection.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | if at `internal/journal/position_projection.go:389` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B2 | if at `internal/journal/position_projection.go:396` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B3 | if at `internal/journal/position_projection.go:400` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B4 | if at `internal/journal/position_projection.go:407` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |

The named focused regressions in this change cover external-order exclusion, canonical reuse, scoped lineage, durable reconcile authority, recovery refusal, startup reservation safety, and held-nonce retention. The full package, race, and repository suites provide integration coverage for all mapped paths.

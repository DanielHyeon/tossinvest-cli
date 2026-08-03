# Branch Test Map: `OrderLineageScope.canonical`

Source: `internal/journal/lineage.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | switch at `internal/journal/lineage.go:72` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B2 | case at `internal/journal/lineage.go:73` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B3 | case at `internal/journal/lineage.go:75` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B4 | case at `internal/journal/lineage.go:77` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B5 | case at `internal/journal/lineage.go:79` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B6 | case at `internal/journal/lineage.go:81` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |

The named focused regressions in this change cover external-order exclusion, canonical reuse, scoped lineage, durable reconcile authority, recovery refusal, startup reservation safety, and held-nonce retention. The full package, race, and repository suites provide integration coverage for all mapped paths.

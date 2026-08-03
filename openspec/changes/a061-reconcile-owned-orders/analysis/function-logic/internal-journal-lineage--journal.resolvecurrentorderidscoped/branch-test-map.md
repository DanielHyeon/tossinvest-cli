# Branch Test Map: `Journal.ResolveCurrentOrderIDScoped`

Source: `internal/journal/lineage.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | if at `internal/journal/lineage.go:313` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B2 | if at `internal/journal/lineage.go:317` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B3 | for at `internal/journal/lineage.go:323` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B4 | if at `internal/journal/lineage.go:325` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B5 | switch at `internal/journal/lineage.go:328` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B6 | case at `internal/journal/lineage.go:329` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B7 | case at `internal/journal/lineage.go:331` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B8 | case at `internal/journal/lineage.go:333` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B9 | if at `internal/journal/lineage.go:338` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |

The named focused regressions in this change cover external-order exclusion, canonical reuse, scoped lineage, durable reconcile authority, recovery refusal, startup reservation safety, and held-nonce retention. The full package, race, and repository suites provide integration coverage for all mapped paths.

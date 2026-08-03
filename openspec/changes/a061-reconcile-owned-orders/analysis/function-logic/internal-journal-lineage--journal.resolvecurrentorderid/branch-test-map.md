# Branch Test Map: `Journal.ResolveCurrentOrderID`

Source: `internal/journal/lineage.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | if at `internal/journal/lineage.go:200` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B2 | for at `internal/journal/lineage.go:204` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B3 | if at `internal/journal/lineage.go:206` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B4 | range at `internal/journal/lineage.go:210` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B5 | if at `internal/journal/lineage.go:211` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B6 | if at `internal/journal/lineage.go:216` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B7 | if at `internal/journal/lineage.go:219` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |

The named focused regressions in this change cover external-order exclusion, canonical reuse, scoped lineage, durable reconcile authority, recovery refusal, startup reservation safety, and held-nonce retention. The full package, race, and repository suites provide integration coverage for all mapped paths.

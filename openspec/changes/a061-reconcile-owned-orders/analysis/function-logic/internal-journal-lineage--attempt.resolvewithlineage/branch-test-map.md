# Branch Test Map: `Attempt.resolveWithLineage`

Source: `internal/journal/lineage.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | if at `internal/journal/lineage.go:133` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B2 | if at `internal/journal/lineage.go:139` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B3 | if at `internal/journal/lineage.go:141` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B4 | if at `internal/journal/lineage.go:146` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B5 | if at `internal/journal/lineage.go:162` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B6 | if at `internal/journal/lineage.go:166` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B7 | if at `internal/journal/lineage.go:169` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B8 | if at `internal/journal/lineage.go:173` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B9 | if at `internal/journal/lineage.go:180` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B10 | if at `internal/journal/lineage.go:216` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B11 | if at `internal/journal/lineage.go:221` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B12 | if at `internal/journal/lineage.go:225` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B13 | if at `internal/journal/lineage.go:230` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |

The named focused regressions in this change cover external-order exclusion, canonical reuse, scoped lineage, durable reconcile authority, recovery refusal, startup reservation safety, and held-nonce retention. The full package, race, and repository suites provide integration coverage for all mapped paths.

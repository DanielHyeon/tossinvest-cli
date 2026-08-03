# Branch Test Map: `TestTrackedFillOrdersCarryLineage`

Source: `internal/journal/fills_test.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | if at `internal/journal/fills_test.go:733` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B2 | if at `internal/journal/fills_test.go:744` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B3 | if at `internal/journal/fills_test.go:747` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B4 | if at `internal/journal/fills_test.go:750` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B5 | if at `internal/journal/fills_test.go:753` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B6 | if at `internal/journal/fills_test.go:761` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B7 | range at `internal/journal/fills_test.go:765` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B8 | if at `internal/journal/fills_test.go:766` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B9 | if at `internal/journal/fills_test.go:768` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B10 | if at `internal/journal/fills_test.go:774` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B11 | if at `internal/journal/fills_test.go:781` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B12 | if at `internal/journal/fills_test.go:785` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B13 | range at `internal/journal/fills_test.go:788` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B14 | if at `internal/journal/fills_test.go:789` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |

The named focused regressions in this change cover external-order exclusion, canonical reuse, scoped lineage, durable reconcile authority, recovery refusal, startup reservation safety, and held-nonce retention. The full package, race, and repository suites provide integration coverage for all mapped paths.

# Branch Test Map: `TestTrackedFillOrdersScopeReusedLineageEndpointByAccount`

Source: `internal/journal/fills_test.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | if at `internal/journal/fills_test.go:800` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B2 | if at `internal/journal/fills_test.go:803` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B3 | if at `internal/journal/fills_test.go:819` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B4 | if at `internal/journal/fills_test.go:822` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B5 | if at `internal/journal/fills_test.go:825` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B6 | if at `internal/journal/fills_test.go:828` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B7 | if at `internal/journal/fills_test.go:833` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B8 | range at `internal/journal/fills_test.go:837` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B9 | if at `internal/journal/fills_test.go:841` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B10 | if at `internal/journal/fills_test.go:844` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |

The named focused regressions in this change cover external-order exclusion, canonical reuse, scoped lineage, durable reconcile authority, recovery refusal, startup reservation safety, and held-nonce retention. The full package, race, and repository suites provide integration coverage for all mapped paths.

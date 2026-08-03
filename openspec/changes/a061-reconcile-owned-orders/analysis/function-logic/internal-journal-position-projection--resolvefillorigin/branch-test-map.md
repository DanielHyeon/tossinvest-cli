# Branch Test Map: `resolveFillOrigin`

Source: `internal/journal/position_projection.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | if at `internal/journal/position_projection.go:141` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B2 | for at `internal/journal/position_projection.go:153` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B3 | if at `internal/journal/position_projection.go:155` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B4 | if at `internal/journal/position_projection.go:162` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B5 | if at `internal/journal/position_projection.go:167` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B6 | range at `internal/journal/position_projection.go:181` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B7 | if at `internal/journal/position_projection.go:182` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B8 | if at `internal/journal/position_projection.go:190` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B9 | if at `internal/journal/position_projection.go:193` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B10 | range at `internal/journal/position_projection.go:195` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B11 | if at `internal/journal/position_projection.go:199` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B12 | if at `internal/journal/position_projection.go:203` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B13 | if at `internal/journal/position_projection.go:216` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B14 | if at `internal/journal/position_projection.go:225` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |

The named focused regressions in this change cover external-order exclusion, canonical reuse, scoped lineage, durable reconcile authority, recovery refusal, startup reservation safety, and held-nonce retention. The full package, race, and repository suites provide integration coverage for all mapped paths.

# Branch Test Map: `EntryGate.CheckEntryFor`

Source: `internal/execgw/symbolgate.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | if at `internal/execgw/symbolgate.go:223` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B2 | if at `internal/execgw/symbolgate.go:224` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B3 | if at `internal/execgw/symbolgate.go:231` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B4 | if at `internal/execgw/symbolgate.go:234` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B5 | range at `internal/execgw/symbolgate.go:243` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B6 | if at `internal/execgw/symbolgate.go:244` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B7 | if at `internal/execgw/symbolgate.go:248` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B8 | range at `internal/execgw/symbolgate.go:252` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B9 | if at `internal/execgw/symbolgate.go:253` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B10 | if at `internal/execgw/symbolgate.go:256` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |

The named focused regressions in this change cover external-order exclusion, canonical reuse, scoped lineage, durable reconcile authority, recovery refusal, startup reservation safety, and held-nonce retention. The full package, race, and repository suites provide integration coverage for all mapped paths.

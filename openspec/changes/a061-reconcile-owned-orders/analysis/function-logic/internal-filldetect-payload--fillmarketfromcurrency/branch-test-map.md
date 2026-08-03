# Branch Test Map: `fillMarketFromCurrency`

Source: `internal/filldetect/payload.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | switch at `internal/filldetect/payload.go:134` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B2 | case at `internal/filldetect/payload.go:135` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B3 | case at `internal/filldetect/payload.go:137` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B4 | case at `internal/filldetect/payload.go:139` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |

The named focused regressions in this change cover external-order exclusion, canonical reuse, scoped lineage, durable reconcile authority, recovery refusal, startup reservation safety, and held-nonce retention. The full package, race, and repository suites provide integration coverage for all mapped paths.

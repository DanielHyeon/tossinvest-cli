# Branch Test Map: `engineFillDetector`

Source: `cmd/tossctl/engine.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | if at `cmd/tossctl/engine.go:401` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B2 | if at `cmd/tossctl/engine.go:402` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |

The named focused regressions in this change cover external-order exclusion, canonical reuse, scoped lineage, durable reconcile authority, recovery refusal, startup reservation safety, and held-nonce retention. The full package, race, and repository suites provide integration coverage for all mapped paths.

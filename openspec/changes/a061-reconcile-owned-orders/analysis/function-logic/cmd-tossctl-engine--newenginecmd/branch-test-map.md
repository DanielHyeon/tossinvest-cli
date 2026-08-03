# Branch Test Map: `newEngineCmd`

Source: `cmd/tossctl/engine.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | Happy path at `cmd/tossctl/engine.go:112` | Focused package regression plus the full suite verifies the branchless contract. |

The named focused regressions in this change cover external-order exclusion, canonical reuse, scoped lineage, durable reconcile authority, recovery refusal, startup reservation safety, and held-nonce retention. The full package, race, and repository suites provide integration coverage for all mapped paths.

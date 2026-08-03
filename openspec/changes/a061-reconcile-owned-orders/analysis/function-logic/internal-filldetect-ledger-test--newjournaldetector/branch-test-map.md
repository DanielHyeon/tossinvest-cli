# Branch Test Map: `newJournalDetector`

Source: `internal/filldetect/ledger_test.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | Happy path at `internal/filldetect/ledger_test.go:38` | Focused package regression plus the full suite verifies the branchless contract. |

The named focused regressions in this change cover external-order exclusion, canonical reuse, scoped lineage, durable reconcile authority, recovery refusal, startup reservation safety, and held-nonce retention. The full package, race, and repository suites provide integration coverage for all mapped paths.

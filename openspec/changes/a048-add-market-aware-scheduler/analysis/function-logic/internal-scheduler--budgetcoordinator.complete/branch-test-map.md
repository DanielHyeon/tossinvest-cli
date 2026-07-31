# Branch Test Map: `BudgetCoordinator.Complete`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | nil/zero capability refuses | capability lifecycle tests | sequential nonzero ID was accepted authority | yes |
| B2 | wrong endpoint key cannot find or mutate commitment | `TestCommitmentCapabilityIsOpaqueAndBoundToCoordinatorKeyClassAndGeneration` | numeric ID could target by key guess | yes |
| B3 | forged/cross-coordinator/cross-key/cross-generation binding refuses | `TestCommitmentCapabilityIsOpaqueAndBoundToCoordinatorKeyClassAndGeneration` | sequential commitment was predictable/forgeable | yes |
| B4 | missing/replayed/cross-class record refuses | capability binding and repeated-completion tests | completion deleted reservation and reopened capacity | yes |
| B5 | unavailable completion clock fails closed and leaves record in-flight | `TestCompletionClockFailureLeavesCommitmentInFlight` | no completion ordering boundary | yes |
| B6 | zero completion instant fails closed and leaves record in-flight | `TestCompletionClockFailureLeavesCommitmentInFlight` | zero timestamp could reconcile against any response | yes |
| B7 | monotonic completion sequence exhausted | `TestCompletionSequenceExhaustionFailsClosed` | wall-only chronology had no sequence guard | yes |
| success | marks completed with a monotonic causal sequence but capacity remains consumed for success/error/cancel | lifecycle tests plus held-response wall-rollback test | wall timestamp comparison allowed an earlier request to reconcile a later completion | yes |

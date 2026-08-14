# Branch Test Map: `ReconcileDriver.blocked`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | no tracker configured | existing driver harness coverage | preserve | preserve |
| Return 1 | an already-durable permanent block defers every adoption candidate and performs no price read | `TestUSIncludedSymbolWaitsUnderAccountWidePermanentReconcile` | preserve | preserve |
| Return 2 | an ordinary covering block defers only its scope; an uncovered candidate may proceed | `TestTheAdoptionTransactionProposesNothing`, reconcile-loop focused tests | preserve | preserve |

a110 does not edit this function. The map exists because the proposal uses its internal branch as evidence that false permanent promotion suppresses baseline creation.

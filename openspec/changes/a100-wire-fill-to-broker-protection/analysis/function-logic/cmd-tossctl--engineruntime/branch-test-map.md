# Branch Test Map: `engineRuntime`

Source: `cmd/tossctl/engine.go` (491-578), AST branches 6.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | nil-hint detector path remains constructor-safe | `TestEngineRuntimeB1IsStructurallyUnreachableWithTheHardcodedNilHintPath` | n/a: pre-existing branch | yes |
| B2 | missing reconcile constructor fails closed | `TestEngineRuntimeConstructionBranchesFailClosedAndAssembleExactSuccess` | n/a: pre-existing branch | yes |
| B3 | exit observer constructor failure fails closed | `TestEngineRuntimeConstructionBranchesFailClosedAndAssembleExactSuccess` | n/a: pre-existing branch | yes |
| B4 | recovery constructor failure fails closed | `TestEngineRuntimeConstructionBranchesFailClosedAndAssembleExactSuccess` | n/a: pre-existing branch | yes |
| B5 | strategy-entry constructor failure | no current named test; A100 RED required before edit | no | no |
| B6 | alert-deliverer constructor failure | no current named test; A100 RED required before edit | no | no |

Before task 3.9 changes this function, add RED tests for B5/B6 and for the
new worker’s verified-gate, recovery ordering, and cancellation contracts.

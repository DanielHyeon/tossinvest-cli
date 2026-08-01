# Branch Test Map: `engineRuntime`

| Branch | Scenario | Test | RED observed | GREEN observed |
| --- | --- | --- | --- | --- |
| B1 | `engineFillDetector` error returns before any other component is built; currently unreachable because this function hardcodes `hints=nil` and that detector path has no error return | `TestEngineRuntimeB1IsStructurallyUnreachableWithTheHardcodedNilHintPath` (structural reachability proof, not branch execution) | no injectable error path | structurally unreachable |
| B2 | `ReconcileDriver` construction error returns before exit/recovery/runtime assembly | `TestEngineRuntimeConstructionBranchesFailClosedAndAssembleExactSuccess/B2_reconcile_construction` | missing direct branch test | pass |
| B3 | `ExitObserver` construction error returns after reconcile construction and before recovery/runtime assembly | `TestEngineRuntimeConstructionBranchesFailClosedAndAssembleExactSuccess/B3_exit_observer_construction` | missing direct branch test | pass |
| B4 | `Recovery` construction error returns before `engine.NewRuntime` | `TestEngineRuntimeConstructionBranchesFailClosedAndAssembleExactSuccess/B4_recovery_construction` | missing direct branch test | pass |
| Scenario | all constructors succeed and the exact three-loop runtime is returned with recovery already holding the entry latch | `TestEngineRuntimeConstructionBranchesFailClosedAndAssembleExactSuccess`, `TestTheLoopSetIsTheSpecifiedThree` | missing direct success construction | pass |
| A047 | missing source/protection/provenance/scheduler keeps strategy loop absent while exit remains | dormant runtime/static tests | missing | pass |

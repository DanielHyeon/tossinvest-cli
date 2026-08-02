# Branch Test Map: `ReadOnly.PositionPolicies`

| Branch | AST control path | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | branchless happy path | valid invocation preserves the function contract | TestReadOnlyPositionPoliciesPreservesReleasedLifecycle; TestTheReadOnlyHandleHasNoWriteMethods | covered by a052 contract RED or pre-existing regression | verified by focused package suite |

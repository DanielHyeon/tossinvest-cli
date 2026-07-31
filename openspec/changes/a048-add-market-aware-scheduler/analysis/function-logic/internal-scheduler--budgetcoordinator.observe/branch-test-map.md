# Branch Test Map: `BudgetCoordinator.Observe`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | nil coordinator is a no-op | nil receiver coverage | existing | yes |
| B2 | manual observation delegates with no cycle authority | `TestManualObserveNeverGainsReconciliationAuthority`, `TestManualNewWindowCannotResetIssuedCapWhenCommitmentsAreEmpty` | empty commitments let manual evidence reset generation/caps | yes |

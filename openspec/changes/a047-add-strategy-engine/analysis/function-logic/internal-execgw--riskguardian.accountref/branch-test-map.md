# Branch Test Map: `RiskGuardian.AccountRef`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | branchless happy-path sentinel (AST has no conditional): returns the constructor-frozen account exactly | `TestAssembleEngineWiresOneProductionGuardianToTheEngineJournalAndExitObserver`, `TestRiskGuardianIssueStrategyEntryDirectSuccessDelegatesOnlyToAtomicJournalIssuance` | missing accessor | pass |

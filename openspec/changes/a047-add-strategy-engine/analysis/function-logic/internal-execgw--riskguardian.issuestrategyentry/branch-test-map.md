# Branch Test Map: `RiskGuardian.IssueStrategyEntry`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | invalid opaque decision returns before snapshot binding, sizing, journal or collection | `TestRiskGuardianIssueStrategyEntryRejectsInvalidOpaqueDecisionBeforeCollection` | missing direct method test | pass |
| B2 | expected policy version or limits digest differs from this Guardian | no authentic direct test: a valid opaque `Decision` cannot be minted while the frozen source manifest is unavailable | unchecked draft | activation-gated unverified |
| B3 | Guardian-owned `StrategyEntryQuantity` refuses the valid decision's entry/stop geometry | pure sizing is covered by `TestStrategyEntryQuantityUsesExactMinimumOfGuardianCaps` and refusal tests, but this method branch has no authentic valid `Decision` fixture | caller quantity draft | activation-gated unverified |
| B4 | `strategyDecisionLineage` returns an encoding error | no authentic direct test: current `DecisionRecord` is scalar-only, but the branch must be re-audited with a valid opaque decision if that schema changes | n/a | activation-gated unverified |
| Scenario | exact lineage is built and delegated to the private atomic `IssueEntry` strategy plan | journal atomic tests and adapter-spy tests cover the callees, but no direct authentic `IssueStrategyEntry` success exists | post-commit lookup window | activation-gated unverified |
| Security | all 60 DecisionRecord fields persist in payload/hash | `TestStrategyDecisionLineagePayloadPreservesCompleteDecisionRecord` | projection could omit new fields | pass |

# Branch Test Map: `RiskGuardian.IssueStrategyEntry`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | invalid opaque decision mints nothing | strategy adapter/Guardian invalid tests | bypass existed in draft | pass |
| B2 | policy/limits binding mismatch | Guardian strategy binding test | unchecked draft | pass |
| B3 | Guardian exact min sizing | `TestStrategyEntryQuantityUsesExactMinimumOfGuardianCaps` | caller quantity draft | pass |
| B4 | atomic decision+reservation+lineage+DISPATCH_START | strategy issuance/journal rollback tests | split commits in draft | pass |
| B5 | risk intent quantity equals lineage and adapter plan | strategy adapter integration test | post-commit lookup window | pass |
| Security | all 60 DecisionRecord fields persist in payload/hash | `TestStrategyDecisionLineagePayloadPreservesCompleteDecisionRecord` | projection could omit new fields | pass |

# Branch Test Map: `completeLineage`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | Happy path (branchless AST): returns the one canonical complete lineage fixture with every exact identity populated | `TestTradeValidationRejectsInvalidStoredAmountsAndIdentity`, `TestUnknownCostKeepsGrossButMarksCostAdjustedMarkoutNotMeasured` | not separately captured for this evidence refresh | package-targeted regression PASS before integration; rerun by gate |

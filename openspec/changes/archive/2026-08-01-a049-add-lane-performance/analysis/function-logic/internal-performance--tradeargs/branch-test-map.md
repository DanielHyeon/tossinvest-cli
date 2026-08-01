# Branch Test Map: `tradeArgs`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | Happy path (branchless AST): returns immutable trade arguments in schema order and converts unknown cost to SQL NULL | `TestStoreSchemaIsSeparateAppendOnlyAndVersioned`, `TestCollectExactReplayIsIdempotentAcrossRestartAndConcurrency`, `TestCollectDivergentReplayFailsClosed`, `TestCollectPersistsUnknownCostAsSQLNullWithoutInventingZero` | not separately captured for this evidence refresh | package-targeted regression PASS before integration; rerun by gate |

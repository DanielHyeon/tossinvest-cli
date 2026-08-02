# Branch Test Map: `adoptionQuoteKey`

| Branch | AST control path | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | branchless happy path | market is lower/trim normalized and symbol is upper/trim normalized into one collision-safe key | `TestUSIncludedSymbolFoldsAdoptsAndOpensExitT0`; `TestObserveCandidatesRefusesCrossMarketDuplicateSymbol` | a052 identity regression contract | verified by focused engine suite |

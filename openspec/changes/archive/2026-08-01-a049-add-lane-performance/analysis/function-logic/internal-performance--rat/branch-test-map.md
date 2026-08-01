# Branch Test Map: `rat`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `if` at line 374: `if !ok {`; invariant: missing/corrupt/alternate path is explicit | `TestMeasureSideAdjustsBuyAndSellMarkoutsWithCostsAndProvenance`, `TestTradeValidationRejectsInvalidStoredAmountsAndIdentity`, `TestUnknownCostKeepsGrossButMarksCostAdjustedMarkoutNotMeasured` | not separately captured for this evidence refresh | package-targeted regression PASS before integration; rerun by gate |

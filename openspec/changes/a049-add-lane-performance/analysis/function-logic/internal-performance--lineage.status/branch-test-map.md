# Branch Test Map: `Lineage.Status`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `range` at line 62: `for _, value := range []string{`; invariant: missing/corrupt/alternate path is explicit | `TestMeasureSideAdjustsBuyAndSellMarkoutsWithCostsAndProvenance`, `TestTradeValidationRejectsInvalidStoredAmountsAndIdentity`, `TestUnknownCostKeepsGrossButMarksCostAdjustedMarkoutNotMeasured` | not separately captured for this evidence refresh | package-targeted regression PASS before integration; rerun by gate |
| B2 | `if` at line 67: `if strings.TrimSpace(value) == "" {`; invariant: missing/corrupt/alternate path is explicit | `TestMeasureSideAdjustsBuyAndSellMarkoutsWithCostsAndProvenance`, `TestTradeValidationRejectsInvalidStoredAmountsAndIdentity`, `TestUnknownCostKeepsGrossButMarksCostAdjustedMarkoutNotMeasured` | not separately captured for this evidence refresh | package-targeted regression PASS before integration; rerun by gate |

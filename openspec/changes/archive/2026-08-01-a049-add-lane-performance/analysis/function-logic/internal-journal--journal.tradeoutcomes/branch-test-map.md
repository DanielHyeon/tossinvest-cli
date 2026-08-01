# Branch Test Map: `Journal.TradeOutcomes`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `if` at line 704: `if err != nil {`; invariant: missing/corrupt/alternate path is explicit | `TestTradeOutcomesAreScopedToTheAccount`, `TestClosedStrategyTradeSourcesReturnsExactIdentifiersFactsAndNullableCost` | not separately captured for this evidence refresh | package-targeted regression PASS before integration; rerun by gate |
| B2 | `for` at line 710: `for rows.Next() {`; invariant: missing/corrupt/alternate path is explicit | `TestTradeOutcomesAreScopedToTheAccount`, `TestClosedStrategyTradeSourcesReturnsExactIdentifiersFactsAndNullableCost` | not separately captured for this evidence refresh | package-targeted regression PASS before integration; rerun by gate |
| B3 | `if` at line 712: `if err != nil {`; invariant: missing/corrupt/alternate path is explicit | `TestTradeOutcomesAreScopedToTheAccount`, `TestClosedStrategyTradeSourcesReturnsExactIdentifiersFactsAndNullableCost` | not separately captured for this evidence refresh | package-targeted regression PASS before integration; rerun by gate |
| B4 | `if` at line 717: `if err := rows.Err(); err != nil {`; invariant: missing/corrupt/alternate path is explicit | `TestTradeOutcomesAreScopedToTheAccount`, `TestClosedStrategyTradeSourcesReturnsExactIdentifiersFactsAndNullableCost` | not separately captured for this evidence refresh | package-targeted regression PASS before integration; rerun by gate |

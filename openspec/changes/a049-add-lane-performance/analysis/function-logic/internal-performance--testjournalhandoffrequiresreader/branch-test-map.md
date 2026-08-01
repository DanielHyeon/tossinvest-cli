# Branch Test Map: `TestJournalHandoffRequiresReader`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `if` at line 67: `if _, err := store.CollectClosedStrategyTrades(context.Background(), nil, ClosedTradeWindow{`; invariant: missing/corrupt/alternate path is explicit | `TestJournalHandoffRequiresReader` (this regression test) | not separately captured for this evidence refresh | package-targeted regression PASS before integration; rerun by gate |

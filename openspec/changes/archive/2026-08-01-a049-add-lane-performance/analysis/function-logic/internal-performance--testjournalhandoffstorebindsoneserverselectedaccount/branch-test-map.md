# Branch Test Map: `TestJournalHandoffStoreBindsOneServerSelectedAccount`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `range` at line 94: `for _, account := range []string{"acct-1", "acct-1"} {`; invariant: missing/corrupt/alternate path is explicit | `TestJournalHandoffStoreBindsOneServerSelectedAccount` (this regression test) | not separately captured for this evidence refresh | package-targeted regression PASS before integration; rerun by gate |
| B2 | `if` at line 95: `if _, err := store.CollectClosedStrategyTrades(context.Background(), reader, ClosedTradeWindow{`; invariant: missing/corrupt/alternate path is explicit | `TestJournalHandoffStoreBindsOneServerSelectedAccount` (this regression test) | not separately captured for this evidence refresh | package-targeted regression PASS before integration; rerun by gate |
| B3 | `if` at line 101: `if _, err := store.CollectClosedStrategyTrades(context.Background(), reader, ClosedTradeWindow{`; invariant: missing/corrupt/alternate path is explicit | `TestJournalHandoffStoreBindsOneServerSelectedAccount` (this regression test) | not separately captured for this evidence refresh | package-targeted regression PASS before integration; rerun by gate |

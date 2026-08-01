# Branch Test Map: `TestJournalHandoffValidatesAccountAndWindowBeforeReading`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `range` at line 48: `for _, window := range []ClosedTradeWindow{`; invariant: missing/corrupt/alternate path is explicit | `TestJournalHandoffValidatesAccountAndWindowBeforeReading` (this regression test) | not separately captured for this evidence refresh | package-targeted regression PASS before integration; rerun by gate |
| B2 | `if` at line 55: `if _, err := store.CollectClosedStrategyTrades(context.Background(), reader, window, nil, at); err == nil {`; invariant: missing/corrupt/alternate path is explicit | `TestJournalHandoffValidatesAccountAndWindowBeforeReading` (this regression test) | not separately captured for this evidence refresh | package-targeted regression PASS before integration; rerun by gate |
| B3 | `if` at line 58: `if reader.calls != 0 {`; invariant: missing/corrupt/alternate path is explicit | `TestJournalHandoffValidatesAccountAndWindowBeforeReading` (this regression test) | not separately captured for this evidence refresh | package-targeted regression PASS before integration; rerun by gate |

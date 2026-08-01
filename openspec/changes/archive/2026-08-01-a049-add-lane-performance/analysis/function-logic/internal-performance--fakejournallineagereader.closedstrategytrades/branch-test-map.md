# Branch Test Map: `fakeJournalLineageReader.ClosedStrategyTrades`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `if` at line 18: `if f.err != nil {`; invariant: missing/corrupt/alternate path is explicit | `TestJournalHandoffConsumesOneExactLineageReadAndCallerOwnedObservations` and the named handoff regression tests | not separately captured for this evidence refresh | package-targeted regression PASS before integration; rerun by gate |

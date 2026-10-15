# Branch Test Map: `adaptCandles`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | wire `[newest, oldest]` becomes chronological | `TestAdaptCandlesOrdersNewestFirstWireChronologically` | pending RED | pending GREEN |
| B2 | empty input remains empty | `TestAdaptCandlesEmpty` | unchanged branch; not RED target | pending GREEN |
| B3 | invalid timestamp remains zero | `TestAdaptCandlesBadTimestamp` | unchanged branch; not RED target | pending GREEN |
| B4 | post-loop reversal swaps local converted entries | `TestAdaptCandlesOrdersNewestFirstWireChronologically` | observed in RED output | pending GREEN |
| I1 | singleton stays intact | `TestAdaptCandlesOrdersNewestFirstWireChronologically` | pending RED | pending GREEN |
| I2 | strict/officialbars ordering and close-to-open conversion | `TestPollBindsEachSuccessorToTheNextNewerObservedBar`; `TestPollConvertsTheBrokerCloseLabelIntoAnOpenInstant` | unchanged branch; not RED target | pending focused run |

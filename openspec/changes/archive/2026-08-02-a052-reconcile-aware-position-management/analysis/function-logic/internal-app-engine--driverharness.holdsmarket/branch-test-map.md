# Branch Test Map: `driverHarness.holdsMarket`

| Branch | AST control path | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `if` line 186 | `if h.prices.currencies == nil {` true/entered and complementary path | driverHarness.holdsMarket | covered by a052 contract RED or pre-existing regression | verified by focused package suite |

# Branch Test Map: `driverHarness.positionMarket`

| Branch | AST control path | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `if` line 224 | `if err != nil {` true/entered and complementary path | driverHarness.positionMarket | covered by a052 contract RED or pre-existing regression | verified by focused package suite |

# Branch Test Map: `StrategyEntryQuantity`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | risk budget is the minimum | `TestStrategyEntryQuantityUsesExactMinimumOfGuardianCaps/risk_budget` | missing strategy sizing | pass |
| B2 | default max quantity 100 is the minimum | `.../default_quantity_cap` | missing | pass |
| B3 | non-divisible notional floors once | `.../notional_floor` | missing | pass |
| B4 | exact notional boundary is not undersized | `.../exact_notional_boundary` | missing | pass |
| B5 | risk or notional produces zero | `TestStrategyEntryQuantityRefusesZeroCapacity` | missing | pass |
| B6 | quantity cap parse/compare | exact-minimum table | missing | pass |
| B7 | entry parse/validation | zero-capacity and contract suites | missing | pass |
| B8 | notional parse/validation | exact-minimum table | missing | pass |
| B9 | quantity cap replaces risk cap | default quantity row | missing | pass |
| B10 | notional cap replaces current cap | notional rows | missing | pass |
| B11 | positive final capacity | all success rows | missing | pass |

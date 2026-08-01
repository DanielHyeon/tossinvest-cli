# Branch Test Map: `RiskBasedQuantity`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | malformed budget | contract invalid-input tests | baseline | pass |
| B2 | negative budget | policy/contract tests | baseline | pass |
| B3 | malformed entry | input unavailable tests | baseline | pass |
| B4 | bad/nonprotective stop | stop contract and zero-capacity tests | baseline | pass |
| B5 | non-divisible width floors once | `TestRiskBasedQuantityFloorsExactlyOnce` | baseline | pass |

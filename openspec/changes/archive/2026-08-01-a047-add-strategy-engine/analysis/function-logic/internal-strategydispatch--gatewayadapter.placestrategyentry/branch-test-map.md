# Branch Test Map: `GatewayAdapter.PlaceStrategyEntry`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | missing gateway/invalid plan | adapter focused tests | alternate fake path | pass |
| B2 | noncanonical/lossy quantity and price | `TestExactFloatRejectsLossyOrNonCanonicalDecimals` | silent float conversion | pass |
| B3 | actual execgw signature preserved | compile-time `Gateway.Place` assertion | string-only result draft | pass |
| B4 | normalized official call returns exact Outcome/error | compile assertion and dispatch outcome table | string-only result | pass |

# Branch Test Map: `Measure`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | foreign/out-of-window observations excluded | `TestMeasureFiltersForeignAndOutOfWindowObservations` | yes | yes |
| B2 | missing markout stays not_measured | `TestMeasureReusesInclusiveMarkoutToleranceAndKeepsMissingStates` | existing | existing |
| B3 | BUY and SELL gross/cost markouts are side-correct with provenance | `TestMeasureSideAdjustsBuyAndSellMarkoutsWithCostsAndProvenance` | yes | yes |
| B4 | BUY and SELL slippage have adverse-positive convention | same | yes | yes |
| B5 | BUY and SELL MFE/MAE use the same side/provenance | same | yes | yes |
| B6 | only measured windows enter gross derivation | markout missing-state test | no — existing | yes |
| B7 | supported side gates gross completion | BUY/SELL side test | yes | yes |
| B8 | valid cost produces cost-adjusted result | BUY/SELL side test | yes | yes |
| B9 | selected observation instant exists | provenance test | no — existing | yes |
| B10 | supplied rows are searched deterministically | provenance test | no — existing | yes |
| B11 | exact instant+price attaches identity/source/version | provenance test | no — existing | yes |
| B12 | decision price controls side-aware slippage | BUY/SELL side test | yes | yes |
| B13 | observations control side-aware MFE/MAE | side/excursion tests | yes | yes |

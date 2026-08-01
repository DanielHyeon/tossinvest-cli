# Branch Test Map: `TestParseRateBudgetResetUsesExactThresholdAndPlausibilityBoundaries`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | all exact threshold/plausibility fixtures run | same function | reset rules were duplicated downstream | yes |
| B2 | any result tuple mismatch fails with full evidence | same function | parser drift was not detected | yes |

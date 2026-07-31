# Branch Test Map: `Console.handleSettingsSave`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | settings seam absent | `TestAnUnwiredSettingsSeamIsExplained` | baseline | yes |
| B2 | config unreadable | existing settings load-error test | baseline | yes |
| B3 | empty/text/NaN/Inf/bounds/off-step/stale fraction | `TestInvalidStopPercentWritesNothing` | yes | yes |
| B4 | every 2..20 half-step saves fraction | `TestEveryAllowedStopPercentConvertsToFraction` | yes | yes |
| B5 | seam rejects or successful 7.5 conversion | `TestInvalidStopPercentWritesNothing`, `TestSavingTheFormWritesTheBlock` | yes | yes |

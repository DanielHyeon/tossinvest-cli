# Branch Test Map: `settingsPage.StopPctSlider`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | unset/rejected fraction renders 5 | existing raw-block/default tests | baseline | yes |
| B2 | 0.075 renders 7.5 and 0.076 renders 7.6 | `TestStopPercentControlIsCSPCompatible` | yes | yes |

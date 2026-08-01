# Branch Test Map: `consoleMarketScheduleSeam`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | invalid config path produces read error/default page | console path/full suite | existing | yes |
| B2 | shared official calendar reader is wired when provided | `TestConsoleMarketScheduleSeamDoesNotActivateApprovedDesiredState` | provenance absent | yes |
| success | missing desired file remains OFF/OFF/none/regular without a network call | `TestConsoleMarketScheduleSeamReadsClosedDefaults` | existing | yes |

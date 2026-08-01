# Branch Test Map: `consoleMarketScheduleReader.Read`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1-B2 | path/load error returns no raw state to UI | console safe-error tests + desired strict tests | existing | yes |
| B3 | new install/market none stays input-free and makes no calendar call | `TestConsoleMarketScheduleSeamReadsClosedDefaults` + DOM no-control test | existing | yes |
| B4-B5 | selected market without one exact official provenance fails closed | seam/full console tests | seam previously claimed desired digest | yes |
| B6-B10 | request clock, location, official read, completion clock and adapter errors are checked | production provenance + fetched-at completion + official/scheduler calendar tests | fetched-at sampled before response | yes |
| success | KR current date yields canonical source/digest/fetched-at while effective remains disabled | `TestConsoleMarketScheduleSeamDoesNotActivateApprovedDesiredState` | provenance absent | yes |
| B1 | retained constructor error | console path contract | existing | yes |
| B2 | desired load error | desired/UI error tests | existing | yes |
| B4 | selected market without calendar reader | fail-closed seam contract | new guard | yes |
| B5 | scope lacks one exact provenance | enum fail-closed contract | new guard | yes |
| B6 | injected deterministic clock | production provenance test | fetched-at absent | yes |
| B7 | market location error | clock market tests | existing | yes |
| B8 | official read error | official/console error tests | new wiring | yes |
| B9 | injected completion clock is sampled after official response | `TestConsoleMarketScheduleFetchedAtIsResponseCompletionTime` | request-start time used | yes |
| B10 | adapter validation error | scheduler calendar tests | existing | yes |

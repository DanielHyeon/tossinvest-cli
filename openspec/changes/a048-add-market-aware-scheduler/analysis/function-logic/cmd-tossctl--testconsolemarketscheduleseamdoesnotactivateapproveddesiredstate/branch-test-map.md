# Branch Test Map: `TestConsoleMarketScheduleSeamDoesNotActivateApprovedDesiredState`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | desired state cannot be persisted | `TestConsoleMarketScheduleSeamDoesNotActivateApprovedDesiredState` | fixture prerequisite | yes |
| B2 | market-schedule read fails | same test | not the authoritative fixture | yes |
| B3 | approved desired state remains desired-only and effective state disabled | same test | activation boundary established | yes |
| B4 | official calendar source/version/fetched-at are present | same test | provenance absent before feature | yes |
| B5 | KR calendar request uses the market-local exact date | same test | provenance absent before feature | yes |

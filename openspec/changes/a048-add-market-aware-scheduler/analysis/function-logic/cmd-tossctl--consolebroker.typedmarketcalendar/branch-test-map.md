# Branch Test Map: `consoleBroker.TypedMarketCalendar`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | resolver error closes provenance | console broker/full tests | existing | yes |
| B2 | non-calendar test broker cannot be treated as provenance source | fail-closed adapter contract | new seam | yes |
| success | shared broker's typed calendar is called with KR/exact date | `TestConsoleMarketScheduleSeamDoesNotActivateApprovedDesiredState` | provenance absent | yes |

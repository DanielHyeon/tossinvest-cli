# Branch Test Map: `stubScheduleCalendar.TypedMarketCalendar`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | concurrent and serial calendar calls are recorded under the fixture mutex | `TestOpeningEveryConsoleReadScreenResolvesTheAccountOnce` | fixture was not concurrency-safe | yes |

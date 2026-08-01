# Branch Test Map: `Client.MarketCalendar`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | unsupported country refuses | existing `MarketCalendar` tests | existing | yes |
| B2 | short, padded, timestamp and impossible dates refuse with zero HTTP requests | `TestMarketCalendarReadsRequireExactGregorianDate` | yes, 14 requests | yes |
| B3 | exact nonempty date is sent once; empty remains omitted | existing calendar fixtures + typed preservation fixture | existing | yes |
| B4/success | official errors propagate and valid payload decodes | official client/calendar suite | existing | yes |
| B4 | official GET error is propagated | official client/calendar suite | existing | yes |

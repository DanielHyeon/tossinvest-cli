# Branch Test Map: `Client.TypedMarketCalendar`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | unsupported country refuses before GET | existing official country contract | existing | yes |
| B2 | non-exact/impossible date refuses before any request | `TestMarketCalendarReadsRequireExactGregorianDate` | yes | yes |
| B3 | exact date is passed to official endpoint | `TestTypedMarketCalendarPreservesNullableSessions` | existing | yes |
| B4 | malformed timezone-aware time fails decoding | `TestTypedMarketCalendarRejectsMalformedTime` | existing | yes |
| success | holiday nulls and next regular session survive typed decode | nullable-session test | existing | yes |

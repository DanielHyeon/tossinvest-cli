# Branch Test Map: `validateMarketCalendarDate`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | empty date remains allowed | existing calendar calls without date | existing | yes |
| B2 | malformed width/separators/padding refuse | `TestMarketCalendarReadsRequireExactGregorianDate` | yes | yes |
| B3 | impossible Gregorian dates refuse | same exact-date test | yes | yes |
| success | exact date proceeds | typed nullable-session test | existing | yes |

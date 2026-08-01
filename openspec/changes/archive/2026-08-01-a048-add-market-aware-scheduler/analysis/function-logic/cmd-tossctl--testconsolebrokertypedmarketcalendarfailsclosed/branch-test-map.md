# Branch Test Map: `TestConsoleBrokerTypedMarketCalendarFailsClosed`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | resolver failure is propagated by identity | `TestConsoleBrokerTypedMarketCalendarFailsClosed/resolver_error` | branch previously claimed without a direct test | yes |
| B2 | generic broker lacking typed calendar is rejected | `TestConsoleBrokerTypedMarketCalendarFailsClosed/broker_lacks_typed_calendar` | branch previously claimed without a direct test | yes |

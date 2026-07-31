# Branch Test Map: `consoleBroker.TypedMarketCalendar`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | resolver error closes provenance before type assertion or calendar call | `TestConsoleBrokerTypedMarketCalendarFailsClosed/resolver_error` | branch previously unasserted | yes |
| B2 | broker without typed calendar capability is rejected | `TestConsoleBrokerTypedMarketCalendarFailsClosed/broker_lacks_typed_calendar` | branch previously unasserted | yes |
| success | shared adapter delegates twice while factory runs once and cached exact reference survives | `TestConsoleBrokerTypedMarketCalendarReusesResolutionAndKeepsExactAccountRef` | cross-change tuple contract broke compilation | yes |

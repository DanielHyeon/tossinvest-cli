# Branch Test Map: `consoleBroker.resolve`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | calendar adapter reuses the cached client and retains the trimmed account reference | `TestConsoleBrokerTypedMarketCalendarReusesResolutionAndKeepsExactAccountRef` | integration contract gap identified | yes |
| B2 | factory error propagates and leaves no invented calendar provenance | `TestConsoleBrokerTypedMarketCalendarFailsClosed/resolver_error` | fail-closed branch previously unasserted | yes |
| success | positions, orders, and market schedule share one serial/concurrent construction | `TestOpeningEveryConsoleReadScreenResolvesTheAccountOnce` | market-schedule seam missing from guard | yes |

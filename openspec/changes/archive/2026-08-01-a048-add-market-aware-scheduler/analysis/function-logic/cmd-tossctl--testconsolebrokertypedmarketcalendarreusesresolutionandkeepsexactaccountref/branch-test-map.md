# Branch Test Map: `TestConsoleBrokerTypedMarketCalendarReusesResolutionAndKeepsExactAccountRef`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | invoke the adapter twice through one cold resolver | `TestConsoleBrokerTypedMarketCalendarReusesResolutionAndKeepsExactAccountRef` | integration callsite had stale tuple arity | yes |
| B2 | either delegated read returns an error | same test | not the success fixture | yes |
| B3 | two reads build the factory more than once | same test | shared calendar reuse was unasserted | yes |
| B4 | post-read cache lookup errors | same test | not the success fixture | yes |
| B5 | padded factory reference is not cached as exact trimmed identity | same test | account-reference preservation was unasserted | yes |
| B6 | adapter changes country/date or call count | same test | direct delegation was only indirectly covered | yes |

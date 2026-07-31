# Branch Test Map: `consoleBroker.resolve`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | existing client reused | shared console reader tests | existing | yes |
| B2 | factory failure propagates | console broker failure tests | existing | yes |
| success | concurrent/serial screens resolve once | `TestOpeningEveryConsoleReadScreenResolvesTheAccountOnce` | existing | yes |

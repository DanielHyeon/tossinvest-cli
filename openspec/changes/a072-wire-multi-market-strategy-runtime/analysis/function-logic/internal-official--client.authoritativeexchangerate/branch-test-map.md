# Branch Test Map: `Client.AuthoritativeExchangeRate`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | nil client refuses without transport | ReadOfficial nil-client test | yes (atomic authority API absent) | yes |
| B2 | configured client refuses before token/data HTTP; sealed client holds boundary during replay | TestAuthoritativeExchangeRateRejectsConfiguredClientBeforeHTTP / TestAuthoritativeExchangeRateKeepsOriginAndReadInOneBoundary | yes (option replay changed transport and raced) | yes |

# Branch Test Map: `Client.ExchangeRate`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | HTTP/auth/decode failure returns no exchange-rate value | existing ExchangeRate integration/error tests | yes (atomic wrapper absent in TOCTOU RED) | yes |

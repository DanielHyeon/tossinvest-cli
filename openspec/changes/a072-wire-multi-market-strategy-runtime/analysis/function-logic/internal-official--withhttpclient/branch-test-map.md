# Branch Test Map: `WithHTTPClient`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | custom transport cannot attest production origin | `official.TestAuthorityOriginRejectsConfiguredTransport` | yes (custom transport could mint authority) | yes |

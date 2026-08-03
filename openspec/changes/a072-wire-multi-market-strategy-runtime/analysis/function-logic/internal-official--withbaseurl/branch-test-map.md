# Branch Test Map: `WithBaseURL`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | configured base remains readable but cannot attest production origin | `official.TestAuthorityOriginRejectsConfiguredTransport` | yes (configured client could mint authority) | yes |

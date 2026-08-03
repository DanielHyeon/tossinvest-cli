# Branch Test Map: `Client.authorityOriginLocked`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | only sealed constructor-owned origin is accepted | TestAuthorityOriginRejectsConfiguredTransport | yes (separate check/read allowed TOCTOU) | yes |

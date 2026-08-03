# Branch Test Map: `Client.AuthorityOrigin`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | reject nil, configured origin, or replaced transport | TestAuthorityOriginRejectsConfiguredTransport | yes (configured transport was authoritative) | yes |

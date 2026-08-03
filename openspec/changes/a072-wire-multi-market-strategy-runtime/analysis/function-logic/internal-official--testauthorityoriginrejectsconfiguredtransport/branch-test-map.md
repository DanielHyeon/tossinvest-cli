# Branch Test Map: `TestAuthorityOriginRejectsConfiguredTransport`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | untouched default client has private authoritative transport | TestAuthorityOriginRejectsConfiguredTransport | yes (origin API absent) | yes |
| B2 | iterate endpoint and HTTP-client override cases | TestAuthorityOriginRejectsConfiguredTransport | yes (overrides were authoritative) | yes |
| B3 | every override is rejected even when values look default | TestAuthorityOriginRejectsConfiguredTransport | yes (overrides were authoritative) | yes |

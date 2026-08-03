# Branch Test Map: `Client.AuthorityOrigin`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | reject nil receiver | TestReadOfficialRefusesInvalidAuthority | yes (atomic method absent) | yes |
| B2 | reject configured origin and preserve sealed origin during replay | authority-origin and concurrent replay tests | yes (origin raced with option writes) | yes |

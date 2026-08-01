# Branch Test Map: `digestAuditEvent`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | happy path digest changes for every persisted audit field and verifies untouched events | `TestAuditEventDigestRejectsValidLookingPersistedTampering`, v2 migration test | audit rows had no integrity digest | PASS |

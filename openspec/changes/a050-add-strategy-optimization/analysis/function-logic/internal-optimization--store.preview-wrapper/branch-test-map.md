# Branch Test Map: `Store.Preview`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | happy path delegates with the store actor and no caller-selected identity | `TestDirectStoreCandidatesRemainBoundToTheStoreActor`, preview/apply lifecycle tests | mutable actor binding existed before hardening | PASS |

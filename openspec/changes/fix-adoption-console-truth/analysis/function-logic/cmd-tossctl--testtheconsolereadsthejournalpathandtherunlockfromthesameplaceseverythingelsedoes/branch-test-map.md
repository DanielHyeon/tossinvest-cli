# Branch Test Map: `TestTheConsoleReadsTheJournalPathAndTheRunLockFromTheSamePlacesEverythingElseDoes`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | resolver and run-lock bindings remain present; writable open remains absent | this test | yes (old assertion named only default profile) | yes |
| B2 | either required binding is missing | this test | yes (resolver assertion changed) | yes |
| B3 | writable journal open appears in console assembly | this test | baseline | yes |

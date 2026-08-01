# Branch Test Map: `ReadOnly.LivePositionExits`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | exit-state result query failure | account-view fault test | yes | pending |
| B2 | positions query failure | account-view fault test | yes | pending |
| B3 | position row scan | existing account-view test | yes | pending |
| B4 | matching exit state | full snapshot read-model test | yes | pending |
| B5 | missing/corrupt exit state | typed unknown test | yes | pending |
| B6 | rows.Err/success | account-view test | yes | pending |
| B7 | legacy identity needs adoption context | legacy read compatibility test | yes | yes |
| B8 | pinned legacy identity resolves | known legacy identity test | yes | yes |
| B9 | unknown legacy identity | typed unknown without default backfill | yes | yes |

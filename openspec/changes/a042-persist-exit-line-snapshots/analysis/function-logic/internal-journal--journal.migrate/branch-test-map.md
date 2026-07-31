# Branch Test Map: `Journal.migrate`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | version read error | migration fault tests | no | pending |
| B2 | schema too new | existing refusal test | no | pending |
| B3 | already target | existing reopen test | no | pending |
| B4 | new database no backup | schema test | no | pending |
| B5 | live database backup | v10 migration backup test | no | pending |
| B6 | skip applied steps | v9→v10 test | no | pending |
| B7 | step failure | v10 rollback test | no | pending |
| B8 | wrong final target | migration plan test | no | pending |
| B9 | success | v10 reopen test | no | pending |

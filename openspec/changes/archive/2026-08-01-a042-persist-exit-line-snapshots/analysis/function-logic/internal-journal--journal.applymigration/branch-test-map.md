# Branch Test Map: `Journal.applyMigration`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | begin failure | migration fault test | no | pending |
| B2 | DDL failure | migration rollback test | no | pending |
| B3 | schema meta version failure | migration rollback test | no | pending |
| B4 | created-at failure | migration rollback test | no | pending |
| B5 | migrated-at failure | migration rollback test | no | pending |
| B6 | user_version failure | migration rollback test | no | pending |
| B7 | commit/success | migration reopen test | no | pending |

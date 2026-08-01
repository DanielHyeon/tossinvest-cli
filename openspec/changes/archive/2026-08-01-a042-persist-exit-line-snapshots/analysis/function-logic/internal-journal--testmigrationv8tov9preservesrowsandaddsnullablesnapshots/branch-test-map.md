# Branch Test Map: `TestMigrationV8ToV9PreservesRowsAndAddsNullableSnapshots`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | fixture close failure | same test | yes | yes |
| B2 | schema version read failure | same test | yes | yes |
| B3 | wrong v9 version | same test | yes | yes |
| B4 | row count mismatch | same test | yes | yes |
| B5 | pragma metadata read failure | same test | yes | yes |
| B6 | unexpected NOT NULL/default | same test | yes | yes |
| B7 | all columns valid | same test | yes | yes |

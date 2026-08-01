# Branch Test Map: `TestFailedV10MigrationRollsBackDDLMetadataAndUserVersion`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | close fixture | named test | existing | yes |
| B2 | broken migration fails | named test | existing | yes |
| B3 | backup path | named test | existing | yes |
| B4 | reconcile instruction | named test | existing | yes |
| B5 | survivor v9 | named test | existing | yes |
| B6 | survivor rows | named test | existing | yes |
| B7 | each artifact | named test | existing | yes |
| B8 | artifact lookup | named test | existing | yes |
| B9 | no partial artifact | named test | existing | yes |
| B10 | close survivor | named test | existing | yes |
| B11 | restored head v11 | named test | expected 10 failed at v11 | yes |
| B12 | restored rows | named test | existing | yes |

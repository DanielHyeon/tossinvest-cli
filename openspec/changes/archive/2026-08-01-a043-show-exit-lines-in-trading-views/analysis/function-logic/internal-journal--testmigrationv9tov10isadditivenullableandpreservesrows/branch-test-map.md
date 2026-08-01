# Branch Test Map: `TestMigrationV9ToV10IsAdditiveNullableAndPreservesRows`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | close v9 fixture | named test | existing | yes |
| B2 | current head version | named test | expected 10 failed at v11 | yes |
| B3 | row preservation | named test | existing | yes |
| B4 | each v10 table | named test | existing | yes |
| B5 | each v10 column | named test | existing | yes |
| B6 | column lookup | named test | existing | yes |
| B7 | nullable/default contract | named test | existing | yes |

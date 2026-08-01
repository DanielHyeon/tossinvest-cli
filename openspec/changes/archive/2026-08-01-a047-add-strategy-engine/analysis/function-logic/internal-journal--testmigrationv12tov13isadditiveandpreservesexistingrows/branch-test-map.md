# Branch Test Map: `TestMigrationV12ToV13IsAdditiveAndPreservesExistingRows`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | v12 seed closes durably | same test | baseline | pass |
| B2 | current schema version is recorded | same test | v13 hardcode fails after v14 | `SchemaVersion` pass |
| B3 | legacy rows survive | same test | baseline | pass |
| B4 | v12/v13 artifact loop remains complete | same test | baseline | pass |
| B5 | missing or duplicate artifact refuses | same test | baseline | pass |

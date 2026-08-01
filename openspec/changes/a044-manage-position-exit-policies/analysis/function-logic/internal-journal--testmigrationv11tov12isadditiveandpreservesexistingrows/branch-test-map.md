# Branch Test Map: `TestMigrationV11ToV12IsAdditiveAndPreservesExistingRows`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | v11 fixture closes durably | `TestMigrationV11ToV12IsAdditiveAndPreservesExistingRows` | yes | yes |
| B2 | resulting version is 12 | same | yes | yes |
| B3 | every prior table row count is preserved | same | yes | yes |
| B4 | v11/v12 artifacts are enumerated | same | yes | yes |
| B5 | sqlite_master query succeeds | same | yes | yes |
| B6 | each artifact exists exactly once | same | yes | yes |

# Branch Test Map: `TestFailedV11MigrationRollsBackIndexAndVersion`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | v10 fixture closes | `TestFailedV11MigrationRollsBackIndexAndVersion` | yes | yes |
| B2 | broken v11 open fails | same | yes | yes |
| B3 | survivor user_version remains 10 | same | yes | yes |
| B4 | index query succeeds | same | yes | yes |
| B5 | partial index is absent | same | yes | yes |
| B6 | survivor closes | same | yes | yes |
| B7 | exactly one backup exists | same | yes | yes |
| B8 | restored backup migrates to exactly v11 | same | yes | yes |

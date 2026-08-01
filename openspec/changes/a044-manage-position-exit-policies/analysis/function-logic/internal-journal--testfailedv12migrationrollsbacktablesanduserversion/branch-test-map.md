# Branch Test Map: `TestFailedV12MigrationRollsBackTablesAndUserVersion`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | v11 fixture closes | `TestFailedV12MigrationRollsBackTablesAndUserVersion` | yes | yes |
| B2 | broken v12 open fails | same | yes | yes |
| B3 | exactly one named backup exists | same | yes | yes |
| B4 | survivor user_version is 11 | same | yes | yes |
| B5 | prior rows survive | same | yes | yes |
| B6 | partial v12 table is absent | same | yes | yes |
| B7 | partial v12 table count is zero | same | yes | yes |
| B8 | both altered tables are enumerated | same | yes | yes |
| B9 | lifecycle column query succeeds | same | yes | yes |
| B10 | partial lifecycle columns are absent | same | yes | yes |
| B11 | survivor closes and backup is exactly v11 without v12 table | same | yes | yes |
| B12 | restored backup migrates to v12 | same | yes | yes |

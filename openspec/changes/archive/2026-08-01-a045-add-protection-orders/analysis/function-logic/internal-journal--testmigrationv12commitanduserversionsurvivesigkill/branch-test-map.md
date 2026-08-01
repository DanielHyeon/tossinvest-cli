# Branch Test Map: `TestMigrationV12CommitAndUserVersionSurviveSIGKILL`
| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | helper subprocess path | self | existing coverage | pass |
| B2 | helper setup failure | self | existing coverage | pass |
| B3 | migration failure | self | existing coverage | pass |
| B4 | parent setup failure | self | existing coverage | pass |
| B5 | process launch failure | self | existing coverage | pass |
| B6 | catalog checks | self | existing coverage | pass |
| B7 | missing v12 object | self | existing coverage | pass |
| B8 | version read failure | self | existing coverage | pass |
| B9 | version differs from 12 | self | current schema leaked into fixture | pass |

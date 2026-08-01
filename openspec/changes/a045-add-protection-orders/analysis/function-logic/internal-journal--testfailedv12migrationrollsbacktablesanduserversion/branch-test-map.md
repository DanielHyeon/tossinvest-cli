# Branch Test Map: `TestFailedV12MigrationRollsBackTablesAndUserVersion`
| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | open temporary DB | self | existing coverage | pass |
| B2 | migrate to v11 | self | existing coverage | pass |
| B3 | seed legacy row | self | existing coverage | pass |
| B4 | install failure injection | self | existing coverage | pass |
| B5 | expect migration failure | self | existing coverage | pass |
| B6 | read rolled-back version | self | existing coverage | pass |
| B7 | require v11 version | self | current schema leaked | pass |
| B8 | enumerate v12 objects | self | existing coverage | pass |
| B9 | catalog query failure | self | existing coverage | pass |
| B10 | reject partially-created object | self | existing coverage | pass |
| B11 | legacy row query failure | self | existing coverage | pass |
| B12 | legacy row changed | self | existing coverage | pass |

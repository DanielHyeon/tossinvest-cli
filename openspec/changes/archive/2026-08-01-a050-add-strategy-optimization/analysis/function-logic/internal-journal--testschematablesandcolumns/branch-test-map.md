# Branch Test Map: `TestSchemaTablesAndColumns`
| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | current v13 catalog and exact protection columns | self | new tables absent from golden list | pass |
| B2 | catalog query failure | self | existing coverage | pass |
| B3 | scan each table name | self | existing coverage | pass |
| B4 | table-name scan failure | self | existing coverage | pass |
| B5 | row iteration failure | self | existing coverage | pass |
| B6 | exact table list differs | self | new tables absent from golden list | pass |
| B7 | exact safety-critical column list differs | self | new columns unpinned | pass |

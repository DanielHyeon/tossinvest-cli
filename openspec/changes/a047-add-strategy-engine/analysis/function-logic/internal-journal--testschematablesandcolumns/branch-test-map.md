# Branch Test Map: `TestSchemaTablesAndColumns`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | table metadata query succeeds | same test | baseline | pass |
| B2 | table rows scan | same test | baseline | pass |
| B3 | iterator error refuses | same test | baseline | pass |
| B4 | exhaustive table set matches | same test | v14 tables missing from expectation | pass after exact addition |
| B5 | reader is closed before PRAGMA loop | same test | baseline | pass |
| B6 | each expected column set queried | same test | baseline | pass |
| B7 | any column mismatch fails | same test | baseline | pass |

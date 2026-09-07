# Branch Test Map: `TestSchemaIndexes`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | open/query | self | existing | yes |
| B2 | row scan | self | existing | yes |
| B3 | row error | self | existing | yes |
| B4 | build set | self | existing | yes |
| B5 | required index loop | self | v20 indexes absent | yes |
| B6 | missing index failure | self | v20 indexes absent | yes |

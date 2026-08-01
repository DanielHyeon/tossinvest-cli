# Branch Test Map: `TestSchemaIndexes`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | schema query failure | TestSchemaIndexes | existing | yes |
| B2 | catalog rows | TestSchemaIndexes | existing | yes |
| B3 | scan failure | TestSchemaIndexes | existing | yes |
| B4 | rows error | TestSchemaIndexes | existing | yes |
| B5 | every required index | TestSchemaIndexes | v11 index absent before migration | yes |
| B6 | missing index | TestSchemaIndexes | v11 index absent before migration | yes |

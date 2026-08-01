# Branch Test Map: `TestEvidenceQueryUsesOnlyRowsRemainingAfterFilter`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | build hidden rows | named test | pre-filter scope overflow | yes |
| B2 | visible evidence assertion | named test | pre-filter query failed | yes |
| B3 | count preservation | named test | regression guard | yes |

# Branch Test Map: `TestMigrationV10ToV11AddsIntentIndexAndPreservesRows`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | seed exit lineage succeeds | `TestMigrationV10ToV11AddsIntentIndexAndPreservesRows` | yes | yes |
| B2 | v10 fixture closes durably | same | yes | yes |
| B3 | target is exactly v11 | same | yes | yes |
| B4 | v11 index exists | same | yes | yes |
| B5 | index covers proposed intent | same | yes | yes |
| B6 | lineage row is preserved | same | yes | yes |

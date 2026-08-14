# Branch Test Map: `Block.Covers`

| Branch | Scenario | Test | RED | GREEN |
|---|---|---|---|---|
| B1 | permanent account block covers all | durable permanent/adoption tests | preserve | yes |
| B2 | market scope selected | existing scope matrix | preserve | yes |
| B3 | market equality decides coverage | existing scope matrix | preserve | yes |
| B4 | symbol scope selected | existing scope matrix | preserve | yes |
| B5 | blank-symbol ordinary covers arbitrary candidate | `TestA110BlankSymbolJournalStateNeverRestoresAsPermanent` | yes | yes |
| B6 | real symbol remains narrow | changing-symbol incident tests | preserve | yes |

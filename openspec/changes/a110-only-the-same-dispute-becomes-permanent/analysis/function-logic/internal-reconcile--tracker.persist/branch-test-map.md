# Branch Test Map: `Tracker.persist`

| Branch | Scenario | Test | RED | GREEN |
|---|---|---|---|---|
| B1 | nil journal unit path | A110 direct tracker cases | preserve | yes |
| B2 | additions are inspected before Enter | persistence suites | preserve | yes |
| B3 | blank ordinary symbol is not written; error is deferred through valid additions/releases | `TestA110BlankSymbolJournalStateNeverRestoresAsPermanent`, `TestA110BlankSymbolPendingDoesNotStarveValidSiblingRelease` | yes | yes |
| B4 | Enter error keeps pending gate | failed ordinary/permanent enter tests | preserve | yes |
| B5 | foreign durable cause wins | existing conflict regression | preserve | yes |
| B6 | releases are inspected after additions even when blank error is pending | `TestA110BlankSymbolPendingDoesNotStarveValidSiblingRelease` | yes | yes |
| B7 | release error is not published | existing release/credit suites | preserve | yes |
| B8 | exact release absence is not published | existing release/credit suites | preserve | yes |
| Return | exact release success is published before deferred blank error returns | sibling release + adjustment-credit suites | yes | yes |
